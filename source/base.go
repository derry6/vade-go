package source

import (
    "context"
    "reflect"
    "sort"
    "sync"

    pkgerrs "github.com/pkg/errors"

    "github.com/derry6/vade-go/pkg/log"
    client "github.com/derry6/vade-go/source/client"
    "github.com/derry6/vade-go/source/parser"
)

var (
    // 确保BaseSource实现了Source接口
    _ Source = (*BaseSource)(nil)
)

type BaseSource struct {
    name          string
    withDeleted   bool
    stores        []*pathStore
    values        map[string]*configValue // 根据优先级聚合后的配置
    callback      func([]*Event)
    defaultParser parser.Parser
    client        client.Client
    prefix        string
    priority      int
    mutex         sync.RWMutex
}

func (bs *BaseSource) Close() error          { return bs.client.Close() }
func (bs *BaseSource) Name() string          { return bs.name }
func (bs *BaseSource) Client() client.Client { return bs.client }
func (bs *BaseSource) Priority() int         { return bs.priority }

func (bs *BaseSource) parse(p parser.Parser, data []byte) (v map[string]interface{}, err error) {
    if p == nil {
        return bs.defaultParser.Parse(data, bs.prefix)
    }
    return p.Parse(data, bs.prefix)
}

// 设置回调
func (bs *BaseSource) OnEvents(cb func([]*Event)) {
    bs.callback = cb
}

// 获取所有的key
func (bs *BaseSource) Keys() (keys []string) {
    bs.mutex.RLock()
    defer bs.mutex.RUnlock()
    for key := range bs.values {
        keys = append(keys, key)
    }
    return
}

// 获取所有的配置
func (bs *BaseSource) All() (values map[string]interface{}) {
    bs.mutex.RLock()
    defer bs.mutex.RUnlock()
    values = make(map[string]interface{})
    for key, value := range bs.values {
        if value == nil {
            values[key] = nil
            continue
        }
        values[key] = value.value
    }
    return
}

// 获取单个配置
func (bs *BaseSource) Get(key string) (value interface{}, ok bool) {
    bs.mutex.RLock()
    defer bs.mutex.RUnlock()
    v, ok := bs.values[key]
    if ok {
        if v == nil {
            return nil, false
        }
        return v.value, true
    }
    return nil, false
}

// 设置配置
func (bs *BaseSource) Set(key string, value interface{}) {
    bs.mutex.Lock()
    defer bs.mutex.Unlock()
    // 设置到每个namespace当中
    for _, n := range bs.stores {
        if _, ok := n.values[key]; ok {
            n.values[key] = value
        }
    }
    if n, ok := bs.values[key]; ok {
        n.value = value
    }
}

// 添加配置
func (bs *BaseSource) AddPath(path string, opts ...PathOption) error {
    bs.mutex.RLock()
    p := bs.findStore(path)
    if p != nil {
        // 已经添加过了
        bs.mutex.RUnlock()
        return nil
    }
    bs.mutex.RUnlock()
    pOpts := newPathOptions(opts...)
    return bs.addPath(path, pOpts)
}

func (bs *BaseSource) watchPath(path string, pOpts *pathOptions) error {
    if pOpts.watchDisabled {
        return nil
    }
    return bs.client.Watch(path, func(data []byte) {
        e2 := bs.handlePathUpdated(path, data)
        if e2 != nil {
            log.Get().Errorf("Can't handle path config updated, source: %q, path: %q :%v", bs.Name(), path, e2)
        }
    })
}
func (bs *BaseSource) addPath(path string, pOpts *pathOptions) error {
    var (
        values map[string]interface{}
        store  = &pathStore{
            path:   path,
            pri:    pOpts.priority,
            parser: pOpts.parser,
            values: map[string]interface{}{},
        }
    )
    data, err := bs.client.Pull(context.Background(), path)
    if err != nil {
        if pOpts.required {
            return pkgerrs.Wrapf(err, "pull required path configs")
        } else {
            log.Get().Warnf("Can not pull path %q configs: %v", path, err)
        }
    } else {
        values, err = bs.parse(pOpts.parser, data)
        if err != nil {
            if pOpts.required {
                return pkgerrs.Wrapf(err, "parse required path configs")
            } else {
                log.Get().Warnf("Can not parse path %q configs: %v", path, err)
            }
        }
    }
    bs.mutex.Lock()
    bs.stores = append(bs.stores, store)
    sort.Sort(pathHighToLow(bs.stores))
    events := bs.populateEvents(store, values)
    bs.mutex.Unlock()
    // todo: handle errors
    _ = bs.watchPath(path, pOpts)
    bs.dispatchEvents(events)
    return nil
}

func (bs *BaseSource) findStore(path string) *pathStore {
    for _, store := range bs.stores {
        if store.path == path {
            return store
        }
    }
    return nil
}
func (bs *BaseSource) findAnotherStore(key string, s1 *pathStore) (v2 interface{}, s2 *pathStore) {
    for _, store := range bs.stores {
        if store == s1 {
            continue
        }
        if v, ok := store.values[key]; ok {
            return v, store
        }
    }
    return
}

// 处理namespace内容变更事件
func (bs *BaseSource) handlePathUpdated(path string, data []byte) error {
    bs.mutex.RLock()
    p := bs.findStore(path)
    bs.mutex.RUnlock()
    if p == nil {
        return pkgerrs.New("path not found")
    }
    values, err := bs.parse(p.parser, data)
    if err != nil {
        return err
    }
    bs.mutex.Lock()
    events := bs.populateEvents(p, values)
    bs.mutex.Unlock()

    if len(events) > 0 {
        bs.dispatchEvents(events)
    }
    return nil
}

func (bs *BaseSource) handleCreated(store *pathStore, ev *Event) bool {
    // 创建
    lastv, ok := bs.values[ev.Key]
    if !ok {
        bs.values[ev.Key] = &configValue{store: store, value: ev.ValueTo}
        return true
    }
    if lastv == nil {
        ev.Action = Updated
        ev.ValueFrom = nil
        bs.values[ev.Key] = &configValue{store: store, value: ev.ValueTo}
        return true
    }
    if lastv.store == nil {
        lastv.store = store
    }
    if store.pri < lastv.store.pri {
        return false
    }
    // 优先级相同或者高于旧的优先级
    lastv.store = store
    if !reflect.DeepEqual(lastv.value, ev.ValueTo) {
        ev.ValueFrom = lastv.value
        ev.Action = Updated
        lastv.value = ev.ValueTo
        return true
    }
    return false
}

func (bs *BaseSource) handleUpdated(store *pathStore, ev *Event) bool {
    lastv, ok := bs.values[ev.Key]
    if !ok || lastv == nil {
        ev.Action = Created
        ev.ValueFrom = nil
        bs.values[ev.Key] = &configValue{store: store, value: ev.ValueTo}
        return true
    }
    if lastv.store == nil {
        lastv.store = store
    }
    if store.pri < lastv.store.pri {
        return false
    }
    lastv.store = store
    if !reflect.DeepEqual(lastv.value, ev.ValueTo) {
        ev.Action = Updated
        ev.ValueFrom = lastv.value
        lastv.value = ev.ValueTo
        return true
    }
    return false
}

func (bs *BaseSource) handleDeleted(store *pathStore, ev *Event) bool {
    lastv, ok := bs.values[ev.Key]
    if !ok {
        return false
    }
    if lastv == nil || lastv.store == nil {
        delete(bs.values, ev.Key)
        ev.ValueFrom = nil
        if lastv != nil {
            ev.ValueFrom = lastv.value
        }
        return true
    }
    if lastv.store != store {
        return false
    }
    isReal := false
    // 如果是该set 维护, 找到其他set
    newValue, newStore := bs.findAnotherStore(ev.Key, store)
    if newStore != nil {
        if !reflect.DeepEqual(newValue, lastv.value) {
            ev.Action = Updated
            ev.ValueFrom = lastv.value
            ev.ValueTo = newValue
            isReal = true
        }
        bs.values[ev.Key].store = newStore
        bs.values[ev.Key].value = newValue
    } else {
        ev.ValueFrom = lastv.value
        ev.ValueTo = nil
        ev.Action = Deleted
        delete(bs.values, ev.Key)
        isReal = true
    }
    return isReal
}

func (bs *BaseSource) populateEvents(store *pathStore, values map[string]interface{}) (events []*Event) {
    if store.values == nil {
        store.values = map[string]interface{}{}
    }
    // 相对于本set的变化
    pathEvents := store.Update(values, bs.withDeleted)
    for _, ev := range pathEvents {
        if ev.Key != "" {
            ev.Path = store.path
            switch ev.Action {
            case Created:
                if bs.handleCreated(store, ev) {
                    events = append(events, ev)
                }
            case Updated:
                if bs.handleUpdated(store, ev) {
                    events = append(events, ev)
                }
            case Deleted:
                if bs.handleDeleted(store, ev) {
                    events = append(events, ev)
                }
            }
        }
    }
    store.values = values
    return
}

func (bs *BaseSource) dispatchEvents(events []*Event) {
    if cb := bs.callback; len(events) > 0 && cb != nil {
        go func() {
            defer func() {
                if err := recover(); err != nil {
                    log.Get().Errorf("Can not handle events: %v", err)
                }
            }()
            cb(events)
        }()
    }
}

func newBaseSource(name string, c client.Client, opts *options) *BaseSource {
    return &BaseSource{
        client:        c,
        name:          name,
        withDeleted:   opts.withDeleted,
        prefix:        opts.prefix,
        priority:      opts.priority,
        stores:        make([]*pathStore, 0),
        values:        map[string]*configValue{},
        callback:      nil,
        defaultParser: parser.NewDefault(),
        mutex:         sync.RWMutex{},
    }
}
