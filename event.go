package vade

import (
    "math/rand"
    "regexp"
    "sync"
    "time"

    "github.com/derry6/vade-go/source"
)


func init() {
    rand.Seed(time.Now().UnixNano())
}

// Event property event
type Event = source.Event

// Event actions
var (
    Created = source.Created
    Updated = source.Updated
    Deleted = source.Deleted
)

// EventHandler 事件处理
type EventHandler interface {
    OnPropertyEvent(ev *Event)
}

type wrapper struct {
    id      int64
    pattern string
    handler EventHandler
}

// dispatcher event dispatcher
type dispatcher struct {
    mutex    sync.RWMutex
    handlers map[int64]*wrapper
}

func newDispatcher() *dispatcher {
    return &dispatcher{
        mutex:    sync.RWMutex{},
        handlers: make(map[int64]*wrapper),
    }
}

func (d *dispatcher) Unwatch(id int64) {
    d.mutex.Lock()
    defer d.mutex.Unlock()
    delete(d.handlers, id)
}

func (d *dispatcher) Watch(pattern string, handler EventHandler) (nextId int64) {
    d.mutex.Lock()
    defer d.mutex.Unlock()

    for id, hdr := range d.handlers {
        if hdr.pattern == pattern && handler == hdr.handler {
            return id
        }
    }
    for {
        nextId = rand.Int63()
        if _, ok := d.handlers[nextId]; ok {
            continue
        }
        d.handlers[nextId] = &wrapper{
            id:      nextId,
            pattern: pattern,
            handler: handler,
        }
        break
    }
    return nextId
}

func (d *dispatcher) handlersOf(key string) (hdrs []EventHandler) {
    d.mutex.RLock()
    defer d.mutex.RUnlock()
    for _, w := range d.handlers {
        if ok, _ := regexp.MatchString(w.pattern, key); ok {
            hdrs = append(hdrs, w.handler)
        }
    }
    return hdrs
}

func (d *dispatcher) Dispatch(event *Event) {
    hdrs := d.handlersOf(event.Key)
    for _, hdr := range hdrs {
        if hdr == nil {
            continue
        }
        hdr.OnPropertyEvent(event)
    }
}


