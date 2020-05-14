package vade

import (
    "sort"
    "sync"

    pkgerrs "github.com/pkg/errors"

    "github.com/derry6/vade-go/pkg/expander"
    "github.com/derry6/vade-go/pkg/log"
    "github.com/derry6/vade-go/source"
)

// Manager 管理不同的 Source
type Manager interface {
    // 添加配置源
    AddSource(src source.Source) error
    Source(name string) (source.Source, error)
    Sources() []source.Source

    AddPath(sourceName string, path string, opts ...source.PathOption) error

    // 配置kv相关操作
    All() (values map[string]interface{})
    Keys() (keys []string)
    Get(key string) (value interface{}, ok bool)
    Set(key string, value interface{})
    SetDefault(key string, value interface{})
    Delete(key string)

    // 监听事件
    Watch(pattern string, handler EventHandler) (watchId int64)
    Unwatch(id int64)
}

type sourceLess []source.Source

func (s sourceLess) Len() int           { return len(s) }
func (s sourceLess) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sourceLess) Less(i, j int) bool { return s[i].Priority() > s[j].Priority() }

type manager struct {
    sources        []source.Source // 按照优先级排序
    ksMap    map[string]source.Source
    defaults map[string]interface{}
    overrides      map[string]interface{}
    expander       expander.Expander
    expandDisabled bool
    dispatcher     *dispatcher
    mutex          sync.RWMutex
}

func (m *manager) setOptions(opts *options) {
    m.expander = expander.New(m.unsafeGet, opts.epOpts...)
    m.expandDisabled = opts.epDisabled
}

func (m *manager) AddSource(newSrc source.Source) (err error) {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    for _, p := range m.sources {
        if p == newSrc {
            return
        }
        if p.Name() == newSrc.Name() {
            return
        }
    }
    keys := newSrc.Keys()
    for _, k := range keys {
        px, ok := m.ksMap[k]
        if ok { // 优先级较高
            if newSrc.Priority() > px.Priority() {
                m.ksMap[k] = newSrc
            }
        } else {
            m.ksMap[k] = newSrc
        }
    }

    newSrc.OnEvents(func(events []*source.Event) {
        m.handleSourceEvents(newSrc, events)
    })

    sources := append(m.sources, newSrc)
    sort.Sort(sourceLess(sources))
    m.sources = sources
    return nil
}

func (m *manager) Source(name string) (source.Source, error) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    for _, s := range m.sources {
        if s.Name() == name {
            return s, nil
        }
    }
    return nil, pkgerrs.New("source not found")
}
func (m *manager) Sources() (sources []source.Source) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    for _, s := range m.sources {
        sources = append(sources, s)
    }
    return sources
}

func (m *manager)  AddPath(sourceName string, path string, opts ...source.PathOption) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    for _, s := range m.sources {
        if s.Name() == sourceName {
            return s.AddPath(path, opts...)
        }
    }
    return nil
}

func (m *manager) unsafeGet(key string) (val interface{}, ok bool) {
    if val, ok = m.overrides[key]; ok {
        return val, ok
    }
    s, ok := m.ksMap[key]
    if ok {
        return s.Get(key)
    }
    val, ok = m.defaults[key]
    return
}

func (m *manager) Get(key string) (val interface{}, ok bool) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    if m.expandDisabled {
        return m.unsafeGet(key)
    }
    v, err := m.expander.Expand(key)
    if err != nil {
        log.Get().Errorf("Can't expand key %q : %v", key, err)
        return nil, false
    }
    return v, true
}

func (m *manager) All() (values map[string]interface{}) {
    values = map[string]interface{}{}
    m.mutex.RLock()
    defer m.mutex.RUnlock()

    // defaults
    for k, v := range m.defaults {
        values[k] = v
    }
    sources := m.sources
    for i := len(sources) - 1; i >= 0; i-- {
        for k, v := range sources[i].All() {
            values[k] = v
        }
    }
    // overrides
    for k, v := range m.overrides {
        values[k] = v
    }
    return values
}

func (m *manager) Keys() (keys []string) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    // 过滤重复Key
    keysMap := map[string]bool{}
    for k := range m.defaults {
        keysMap[k] = true
    }
    for k := range m.ksMap {
        keysMap[k] = true
    }
    for k := range m.overrides {
        keysMap[k] = true
    }
    for k := range keysMap {
        keys = append(keys, k)
    }
    return keys
}

func (m *manager) Set(key string, value interface{}) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    m.overrides[key] = value
}
func (m *manager) Delete(key string) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    delete(m.overrides, key)
    delete(m.defaults, key)
}

func (m *manager) SetDefault(key string, value interface{}) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    m.defaults[key] = value
}

func (m *manager) Watch(pattern string, cb EventHandler) (watchId int64) {
    return m.dispatcher.Watch(pattern, cb)
}
func (m *manager) Unwatch(watchId int64) {
    m.dispatcher.Unwatch(watchId)
}

// 是否在高优先级的 source 中存在该key
func (m *manager) inHigherSource(key string, src source.Source) bool {
    for _, p := range m.sources {
        if p == src || p.Priority() < src.Priority() {
            continue
        }
        if _, ok := p.Get(key); !ok {
            continue
        }
        return true
    }
    return false
}

// 找到低优先级的 source
func (m *manager) lowerSource(src source.Source, key string) (source.Source, interface{}) {
    for _, p := range m.sources {
        if p == src || p.Priority() > src.Priority() {
            continue
        }
        if v, ok := p.Get(key); ok {
            return p, v
        }
    }
    return nil, nil
}

func (m *manager) handleDeletedEvent(src source.Source, ev *source.Event) {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    if lastSrc, ok := m.ksMap[ev.Key]; ok {
        if lastSrc != src {
            // 低优先级的source发生的事件, h忽略
            if lastSrc.Priority() > src.Priority() {
                return
            }
        }
        // 找到低优先级的source
        lowerSrc, v := m.lowerSource(src, ev.Key)
        if lowerSrc == nil {
            delete(m.ksMap, ev.Key)
        } else {
            ev.ValueTo = v
            ev.Action = Updated
            m.ksMap[ev.Key] = lowerSrc
        }
    }
}
func (m *manager) handleUpdatedEvent(src source.Source, ev *source.Event) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    lastSrc, ok := m.ksMap[ev.Key]
    if !ok {
        m.ksMap[ev.Key] = src
        return
    }
    if lastSrc == src {
        return
    }
    // src > lastSrc
    // 低优先级发生的事件 ev.src
    if lastSrc.Priority() > src.Priority() {
        log.Get().Debugf("Property in high priority source, ignored: %q", ev.Key)
        return
    }
    // 如果是相同优先级发生的事件
    // 或者高优先级的source发生的事件
    v, ok2 := lastSrc.Get(ev.Key)
    if !ok2 {
        // should never reach here
        m.ksMap[ev.Key] = src
        ev.Action = Created
        ev.ValueFrom = nil
    } else {
        ev.Action = Updated
        ev.ValueFrom = v
    }
    m.ksMap[ev.Key] = src
}

func (m *manager) handleSourceEvents(src source.Source, events []*source.Event) {
    for _, ev := range events {
        if ev.Action == Created || ev.Action == Updated {
            m.handleUpdatedEvent(src, ev)
            m.dispatcher.Dispatch(ev)
        } else if ev.Action == Deleted {
            m.handleDeletedEvent(src, ev)
            m.dispatcher.Dispatch(ev)
        }
    }
}

func newManager() *manager {
    m := &manager{
        sources:    make([]source.Source, 0),
        ksMap:      make(map[string]source.Source),
        overrides:  make(map[string]interface{}),
        defaults:   make(map[string]interface{}),
        dispatcher: newDispatcher(),
        mutex:      sync.RWMutex{},
    }
    m.expander = expander.New(m.unsafeGet)
    return m
}
