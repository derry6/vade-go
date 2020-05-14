package client

import (
    "context"
    "crypto/md5"
    "encoding/hex"
    "io"
    "io/ioutil"
    "os"
    "sync"
    "time"

    "github.com/fsnotify/fsnotify"
    pkgerrs "github.com/pkg/errors"

    "github.com/derry6/vade-go/pkg/log"
)

const File = "file"

var (
    _ Client    = (*fileClient)(nil)
    _ fsWatcher = (*fsnotify.Watcher)(nil)
)

func init() {
    _ = RegisterClient(File, newFileClient)
}

// 定义fsnotify.Watcher接口
type fsWatcher interface {
    io.Closer
    Add(name string) error
    Remove(name string) error
}

// interface
type Watcher interface {
    fsWatcher
    Events() chan fsnotify.Event
    Errors() chan error
}

type watcherImpl struct {
    *fsnotify.Watcher
}

func (w *watcherImpl) Events() chan fsnotify.Event { return w.Watcher.Events }
func (w *watcherImpl) Errors() chan error          { return w.Watcher.Errors }

func createWatcher(disabled bool) Watcher {
    if disabled {
        return nil
    }
    fw, err := fsnotify.NewWatcher()
    if err != nil {
        log.Get().Warnf("Can't create file watcher, so disabled: %v", err)
        return nil
    }
    return &watcherImpl{fw}
}

type fileClient struct {
    w    Watcher
    md5s map[string]string
    mu   sync.RWMutex
    cbs  map[string]func(data []byte)
}

func (c *fileClient) Close() error { return c.stop() }
func (c *fileClient) Pull(ctx context.Context, path string) (data []byte, err error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    if data, err = ioutil.ReadFile(path); err != nil {
        return
    }
    m5 := md5.New()
    m5.Write(data)
    c.md5s[path] = hex.EncodeToString(m5.Sum(nil))
    return
}
func (c *fileClient) Push(ctx context.Context, path string, data []byte) error {
    return nil
}

func (c *fileClient) Watch(path string, cb ChangedCallback) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.w == nil {
        return pkgerrs.New("watch disabled")
    }
    // already watched
    if _, ok := c.cbs[path]; ok {
        return nil
    }
    c.cbs[path] = cb
    return c.w.Add(path)
}

// 文件内容发生变化
func (c *fileClient) handleUpdated(filePath string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    cb, _ := c.cbs[filePath]
    if cb == nil {
        return
    }
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        log.Get().Debugf("Can't read file %q content: %v", filePath, err)
    }
    cb(data)
}

func (c *fileClient) isFileCanRead(fileName string) bool {
    if f, err := os.Open(fileName); err == nil && f != nil {
        _ = f.Close()
        return true
    }
    return false
}

// 处理文件系统事件
func (c *fileClient) handleEvents(event fsnotify.Event) {
    if event.Op&fsnotify.Remove == fsnotify.Remove {
        time.Sleep(50 * time.Millisecond)
        if !c.isFileCanRead(event.Name) {
            c.w.Remove(event.Name)
        } else {
            c.w.Add(event.Name)
            c.handleUpdated(event.Name)
        }
        return
    }
    if event.Op&fsnotify.Chmod == fsnotify.Chmod {
        time.Sleep(50 * time.Millisecond)
        if !c.isFileCanRead(event.Name) {
            c.w.Remove(event.Name)
        }
        return
    }
    if event.Op&fsnotify.Rename == fsnotify.Rename {
        time.Sleep(50 * time.Millisecond)
        if !c.isFileCanRead(event.Name) {
            c.w.Remove(event.Name)
        }
        return
    }
    if event.Op&fsnotify.Write == fsnotify.Write {
        time.Sleep(50 * time.Millisecond)
        c.handleUpdated(event.Name)
    }
}
func (c *fileClient) start() {
    if c.w == nil {
        return
    }
    events := c.w.Events()
    errors := c.w.Errors()
    for {
        select {
        case fsEvent, ok := <-events:
            if !ok {
                return
            }
            c.handleEvents(fsEvent)
        case err, ok := <-errors:
            if !ok {
                return
            }
            log.Get().Errorf("File watcher error: %v", err)
        }
    }
}

func (c *fileClient) stop() error {
    if c.w != nil {
        return c.w.Close()
    }
    return nil
}

func newFileClient(cfg *Config) (Client, error) {
    c := &fileClient{
        md5s: map[string]string{},
        cbs:  map[string]func(data []byte){},
        w:    createWatcher(cfg.WatchDisabled),
    }
    go c.start()
    return c, nil
}
