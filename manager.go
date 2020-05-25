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
	ksMap          map[string]source.Source
	defaults       map[string]interface{}
	overrides      map[string]interface{}
	expander       expander.Expander
	expandDisabled bool
	dispatcher     *dispatcher
	mutex          sync.RWMutex
}

func (mgr *manager) AddSource(newSrc source.Source) (err error) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	for _, p := range mgr.sources {
		if p == newSrc {
			return
		}
		if p.Name() == newSrc.Name() {
			return
		}
	}
	keys := newSrc.Keys()
	for _, k := range keys {
		px, ok := mgr.ksMap[k]
		if ok { // 优先级较高
			if newSrc.Priority() > px.Priority() {
				mgr.ksMap[k] = newSrc
			}
		} else {
			mgr.ksMap[k] = newSrc
		}
	}

	newSrc.OnEvents(func(events []*source.Event) {
		mgr.handleSourceEvents(newSrc, events)
	})

	sources := append(mgr.sources, newSrc)
	sort.Sort(sourceLess(sources))
	mgr.sources = sources
	return nil
}

func (mgr *manager) Source(name string) (source.Source, error) {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	for _, s := range mgr.sources {
		if s.Name() == name {
			return s, nil
		}
	}
	return nil, pkgerrs.New("source not found")
}
func (mgr *manager) Sources() (sources []source.Source) {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	for _, s := range mgr.sources {
		sources = append(sources, s)
	}
	return sources
}

func (mgr *manager) AddPath(sourceName string, path string, opts ...source.PathOption) error {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	for _, s := range mgr.sources {
		if s.Name() == sourceName {
			return s.AddPath(path, opts...)
		}
	}
	return nil
}

func (mgr *manager) unsafeGet(key string) (val interface{}, ok bool) {
	if val, ok = mgr.overrides[key]; ok {
		return val, ok
	}
	s, ok := mgr.ksMap[key]
	if ok {
		return s.Get(key)
	}
	val, ok = mgr.defaults[key]
	return
}

func (mgr *manager) Get(key string) (val interface{}, ok bool) {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	if mgr.expandDisabled {
		return mgr.unsafeGet(key)
	}
	v, err := mgr.expander.Expand(key)
	if err != nil {
		log.Get().Errorf("Can't expand key %q : %v", key, err)
		return nil, false
	}
	return v, true
}

func (mgr *manager) All() (values map[string]interface{}) {
	values = map[string]interface{}{}
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()

	// defaults
	for k, v := range mgr.defaults {
		values[k] = v
	}
	sources := mgr.sources
	for i := len(sources) - 1; i >= 0; i-- {
		for k, v := range sources[i].All() {
			values[k] = v
		}
	}
	// overrides
	for k, v := range mgr.overrides {
		values[k] = v
	}
	return values
}

func (mgr *manager) Keys() (keys []string) {
	mgr.mutex.RLock()
	defer mgr.mutex.RUnlock()
	// 过滤重复Key
	keysMap := map[string]bool{}
	for k := range mgr.defaults {
		keysMap[k] = true
	}
	for k := range mgr.ksMap {
		keysMap[k] = true
	}
	for k := range mgr.overrides {
		keysMap[k] = true
	}
	for k := range keysMap {
		keys = append(keys, k)
	}
	return keys
}

func (mgr *manager) Set(key string, value interface{}) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	mgr.overrides[key] = value
}
func (mgr *manager) Delete(key string) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	delete(mgr.overrides, key)
	delete(mgr.defaults, key)
}

func (mgr *manager) SetDefault(key string, value interface{}) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	mgr.defaults[key] = value
}

func (mgr *manager) Watch(pattern string, cb EventHandler) (watchId int64) {
	return mgr.dispatcher.Watch(pattern, cb)
}
func (mgr *manager) Unwatch(watchId int64) {
	mgr.dispatcher.Unwatch(watchId)
}

// 是否在高优先级的 source 中存在该key
func (mgr *manager) inHigherSource(key string, src source.Source) bool {
	for _, p := range mgr.sources {
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
func (mgr *manager) lowerSource(src source.Source, key string) (source.Source, interface{}) {
	for _, p := range mgr.sources {
		if p == src || p.Priority() > src.Priority() {
			continue
		}
		if v, ok := p.Get(key); ok {
			return p, v
		}
	}
	return nil, nil
}

func (mgr *manager) handleDeletedEvent(src source.Source, ev *source.Event) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	if lastSrc, ok := mgr.ksMap[ev.Key]; ok {
		if lastSrc != src {
			// 低优先级的source发生的事件, h忽略
			if lastSrc.Priority() > src.Priority() {
				return
			}
		}
		// 找到低优先级的source
		lowerSrc, v := mgr.lowerSource(src, ev.Key)
		if lowerSrc == nil {
			delete(mgr.ksMap, ev.Key)
		} else {
			ev.ValueTo = v
			ev.Action = Updated
			mgr.ksMap[ev.Key] = lowerSrc
		}
	}
}
func (mgr *manager) handleUpdatedEvent(src source.Source, ev *source.Event) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	lastSrc, ok := mgr.ksMap[ev.Key]
	if !ok {
		mgr.ksMap[ev.Key] = src
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
		mgr.ksMap[ev.Key] = src
		ev.Action = Created
		ev.ValueFrom = nil
	} else {
		ev.Action = Updated
		ev.ValueFrom = v
	}
	mgr.ksMap[ev.Key] = src
}

func (mgr *manager) handleSourceEvents(src source.Source, events []*source.Event) {
	_events := []*source.Event{}
	for _, ev := range events {
		if ev.Action == Created ||
			ev.Action == Updated {
			mgr.handleUpdatedEvent(src, ev)
			_events = append(_events, ev)
		} else if ev.Action == Deleted {
			mgr.handleDeletedEvent(src, ev)
			_events = append(_events, ev)
		}
	}
	if len(_events) > 0 {
		mgr.dispatcher.Dispatch(_events)
	}
}

func newManager(opts ...Option) (Manager, error) {
	vOpts := newOptions(opts...)
	mgr := &manager{
		sources:    make([]source.Source, 0),
		ksMap:      make(map[string]source.Source),
		overrides:  make(map[string]interface{}),
		defaults:   make(map[string]interface{}),
		dispatcher: newDispatcher(),
		mutex:      sync.RWMutex{},
	}
	if err := mgr.init(vOpts); err != nil {
		return nil, err
	}
	return mgr, nil
}
