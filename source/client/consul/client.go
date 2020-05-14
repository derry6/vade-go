package consul

import (
    "context"
    "crypto/tls"
    "errors"
    "net/http"
    "strings"
    "sync"
    "time"

    consulapi "github.com/hashicorp/consul/api"
    pkgerrs "github.com/pkg/errors"

    "github.com/derry6/vade-go/source/client"
)

const (
    Name = "consul"
)

var (
    _ client.Client = (*Client)(nil)
)

type Client struct {
    client        *consulapi.KV
    watchDisabled bool // watch enable
    dataCenter    string
    timeout       time.Duration
    mutex         sync.RWMutex
    watchers      map[string]func(data []byte)
}

func (c *Client) Close() error { return nil }
func (c *Client) Name() string { return Name }

func (c *Client) Pull(ctx context.Context, path string) (data []byte, err error) {
    opts := &consulapi.QueryOptions{Datacenter: c.dataCenter}
    kv, _, err := c.client.Get(path, opts)
    if err != nil {
        return nil, err
    }
    if kv == nil {
        return nil, errors.New("not found")
    }
    return kv.Value, nil
}

func (c *Client) Push(ctx context.Context, path string, data []byte) error {
    opts := &consulapi.WriteOptions{
        Datacenter: c.dataCenter,
    }
    kv := &consulapi.KVPair{Key: path, Value: data}
    _, err := c.client.Put(kv, opts)
    return err
}

func (c *Client) Watch(path string, cb client.ChangedCallback) error {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    if c.watchDisabled {
        return errors.New("watch disabled")
    }
    _, ok := c.watchers[path]
    if ok {
        if cb == nil {
            delete(c.watchers, path)
        }
        return nil
    }
    c.watchers[path] = cb
    go c.doListen(path)
    return nil
}

func (c *Client) doListen(path string) {
    index := uint64(0)
    waitTime := 10 * c.timeout
    // todo: close watcher
    for {
        opts := consulapi.QueryOptions{WaitIndex: index, WaitTime: waitTime}
        kvp, meta, err := c.client.Get(path, &opts)
        if kvp == nil && err == nil {
            time.Sleep(c.timeout)
            continue
        }
        if err != nil {
            time.Sleep(waitTime)
            continue
        }
        index = meta.LastIndex
        c.mutex.RLock()
        cb, _ := c.watchers[path]
        c.mutex.RUnlock()
        if cb != nil {
            cb(kvp.Value)
            continue
        }
    }
}

// config.
func newHttpClient(transport *http.Transport, tlsConfig *tls.Config) (*http.Client, error) {
    c := &http.Client{
        Transport: transport,
    }
    transport.TLSClientConfig = tlsConfig
    return c, nil
}

func NewClient(cfg *client.Config) (client.Client, error) {
    if cfg.Address == "" {
        return nil, pkgerrs.New("missing server address")
    }
    addrs := strings.Split(cfg.Address, ",")
    if len(addrs) == 0 {
        addrs = append(addrs, "127.0.0.1:8500")
    }
    if cfg.Timeout == 0 {
        cfg.Timeout = time.Second
    }
    c1 := consulapi.DefaultConfig()
    c1.Token = cfg.Token
    c1.Address = addrs[0]
    c1.Datacenter = cfg.DataCenter
    c1.WaitTime = cfg.Timeout
    // todo: support tls config

    ac, err := consulapi.NewClient(c1)
    if err != nil {
        return nil, err
    }
    cli := &Client{
        client:        ac.KV(),
        watchDisabled: cfg.WatchDisabled,
        dataCenter:    cfg.DataCenter,
        timeout:       cfg.Timeout,
        mutex:         sync.RWMutex{},
        watchers:      map[string]func(data []byte){},
    }
    return cli, nil
}
