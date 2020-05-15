package etcd

import (
    "context"
    "strings"
    "sync"
    "time"

    "github.com/coreos/etcd/mvcc/mvccpb"
    pkgerrs "github.com/pkg/errors"
    etcdv3 "go.etcd.io/etcd/clientv3"

    "github.com/derry6/vade-go/source/client"
)

const (
    Name = "etcd"
)

var (
    _ client.Client = (*Client)(nil)
)

func init() {
    _ = client.RegisterClient(Name, NewClient)
}

type Client struct {
    timeout time.Duration
    etcd    *etcdv3.Client
    revs    map[string]int64
    watcher etcdv3.Watcher
    watches map[string]struct{}
    close   chan struct{}
    mu      sync.RWMutex
}

func setTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
    var cancel context.CancelFunc = func() {}
    if deadline, ok := ctx.Deadline(); ok {
        if time.Now().Add(timeout).Before(deadline) {
            ctx, cancel = context.WithTimeout(ctx, timeout)
        }
    }
    return ctx, cancel
}

func (c *Client) Close() error {
    close(c.close)
    return nil
}

func (c *Client) Pull(ctx context.Context, path string) ([]byte, error) {
    var cancel context.CancelFunc
    ctx, cancel = setTimeout(ctx, c.timeout)
    defer cancel()
    rsp, err := c.etcd.KV.Get(ctx, path)
    if err != nil {
        return nil, err
    }
    if rsp.Count != 1 {
        return nil, pkgerrs.Errorf("invalid response: count=%d", rsp.Count)
    }
    c.mu.Lock()
    c.revs[path] = rsp.Header.Revision
    c.mu.Unlock()
    return rsp.Kvs[0].Value, nil
}

func (c *Client) Push(ctx context.Context, path string, data []byte) error {
    var cancel context.CancelFunc
    ctx, cancel = setTimeout(ctx, c.timeout)
    defer cancel()
    _, err := c.etcd.KV.Put(ctx, path, string(data))
    if err != nil {
        return err
    }
    return nil
}
func (c *Client) Watch(path string, cb client.ChangedCallback) error {
    c.mu.Lock()
    if _, ok := c.watches[path]; ok {
        c.mu.Unlock()
        return nil
    }
    c.watches[path] = struct{}{}
    startRev, _ := c.revs[path]
    c.mu.Unlock()
    wc := c.watcher.Watch(context.Background(), path, etcdv3.WithRev(startRev))
    go func() {
        for {
            select {
            case rsp := <-wc:
                for _, ev := range rsp.Events {
                    switch ev.Type {
                    case mvccpb.DELETE:
                    case mvccpb.PUT:
                        cb(ev.Kv.Value)
                    }
                }
            case <-c.close:
                return
            }
        }
    }()
    return nil
}

func NewClient(cfg *client.Config) (client.Client, error) {
    if cfg.Address == "" {
        return nil, pkgerrs.New("missing etcd server info")
    }
    eps := strings.Split(cfg.Address, ";")
    if len(eps) == 0 {
        eps = append(eps, "127.0.0.1:2379")
    }
    if cfg.Timeout == 0 {
        cfg.Timeout = time.Second
    }
    ec, err := etcdv3.New(etcdv3.Config{
        Endpoints:   eps,
        DialTimeout: cfg.Timeout,
        Username:    cfg.Username,
        Password:    cfg.Password,
    })
    if err != nil {
        return nil, err
    }
    return &Client{
        timeout: cfg.Timeout,
        etcd:    ec,
        close:   make(chan struct{}),
        revs:    map[string]int64{},
        watcher: etcdv3.NewWatcher(ec),
        watches: map[string]struct{}{},
        mu:      sync.RWMutex{},
    }, nil
}
