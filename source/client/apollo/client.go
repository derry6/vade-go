package apollo

import (
    "context"
    "encoding/json"
    "net/http"
    "net/url"
    "strings"
    "sync"
    "time"

    "github.com/pkg/errors"

    "github.com/derry6/vade-go/pkg/log"
    "github.com/derry6/vade-go/source/client"
)

const (
    Name = "apollo"
)

const (
    watchURL    = "%s/notifications/v2?appId=%s&cluster=%s&notifications=%s"
    cacheURL    = "%s/configfiles/json/%s/%s/%s"
    nonCacheURL = "%s/configs/%s/%s/%s"
)

var (
    ErrNotFound    = errors.New("not found")
    ErrNotModified = errors.New("not modified")
)

var (
    _ client.Client = (*Client)(nil)
)

func init() {
    _ = client.RegisterClient(Name, NewClient)
}

// appId@cluster
type Client struct {
    appId           string
    cluster         string
    timeout  time.Duration
    httpc    *httpClient
    selector Selector
    snapshot *Snapshot
    mutex    sync.RWMutex
    notificationIDs map[string]int64                  // namespace@cluster@app -> id
    releaseKeys     map[string]string                 // namespace@cluster@app -> id
    watches         map[string][]string               // cluster@app -> namespaces
    callbacks       map[string]client.ChangedCallback // namespace@cluster@app -> callbacks
    watchDisabled   bool
}

func (c *Client) Close() error { return nil }
func (c *Client) Name() string { return Name }
func (c *Client) Pull(ctx context.Context, path string) ([]byte, error) {
    p := newPath(path, c.cluster, c.appId)
    return c.pull(p, true)
}

func (c *Client) Push(ctx context.Context, path string, data []byte) error {
    return nil
}

func (c *Client) Watch(path string, cb client.ChangedCallback) error {
    p := newPath(path, c.cluster, c.appId)

    c.mutex.Lock()
    defer c.mutex.Unlock()
    if c.watchDisabled {
        return errors.New("watch disabled")
    }
    c.callbacks[p.String()] = cb
    found := false
    namespaces, ok := c.watches[p.watchID()]
    if ok {
        for _, namespace := range namespaces {
            if namespace == p.namespace {
                found = true
                break
            }
        }
    }
    if !found {
        namespaces = append(namespaces, p.namespace)
        c.watches[p.watchID()] = namespaces
    }
    if !ok {
        go c.listen(p)
    }
    return nil
}

func (c *Client) pull(p *aPath, cached bool) (data []byte, err error) {
    // 从服务端拉取配置
    if cached {
        if data, err = c.snapshot.get(p); err == nil {
            releaseKey, nId, e2 := c.snapshot.getReleaseKey(p)
            if e2 == nil {
                c.mutex.Lock()
                c.releaseKeys[p.String()] = releaseKey
                c.notificationIDs[p.String()] = nId
                c.mutex.Unlock()
            }
            err = e2
            return
        }
    }
    _, data, err = c.pullServer(p, cached)
    return data, err
}

// fetch data from apollo server
func (c *Client) doHttpGet(p *aPath, releaseKey string, cached bool) (code int, body []byte, err error) {
    var (
        values = url.Values{}
        req    *http.Request
    )
    appId := url.QueryEscape(p.appId)
    cluster := url.QueryEscape(p.cluster)
    ns := url.QueryEscape(p.namespace)

    values.Set("releaseKey", releaseKey)
    server := c.selector.Select()
    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()

    if cached {
        req, err = c.httpc.NewRequest(http.MethodGet, cacheURL, values, server, appId, cluster, ns)
    } else {
        req, err = c.httpc.NewRequest(http.MethodGet, nonCacheURL, values, server, appId, cluster, ns)
    }
    if err != nil {
        return
    }
    result, err := c.httpc.Raw(ctx, req)
    if err != nil {
        return 0, nil, err
    }
    return result.Code, result.Body, err
}
func (c *Client) pullServer(p *aPath, cached bool) (rKey string, data []byte, err error) {
    var (
        code int
        body []byte
    )
    c.mutex.RLock()
    rKeyID := p.String()
    rKey = c.releaseKeys[rKeyID]
    c.mutex.RUnlock()
    if code, body, err = c.doHttpGet(p, rKey, cached); err != nil {
        return
    }
    switch code {
    case http.StatusNotFound:
        return rKey, nil, ErrNotFound
    case http.StatusOK:
        break
    case http.StatusNotModified:
        return rKey, nil, ErrNotModified
    default:
        return rKey, nil, errors.Errorf("status code is %d", code)
    }
    if cached {
        body, err = c.parseCachedRsp(p, body)
    } else {
        rKey, body, err = c.parseRsp(p, body)
        if err != nil && rKey != "" {
            c.mutex.Lock()
            c.releaseKeys[rKeyID] = rKey
            c.mutex.Unlock()
        }
    }
    return rKey, body, err
}
func (c *Client) loadNew(p *aPath, lastKey string, nId int64) (data []byte, err error) {
    var newKey string
    newKey, data, err = c.pullServer(p, false)
    if err != nil {
        log.Get().Errorf("Can't load apollo %q configs: %v", p, err.Error())
        return
    }
    if newKey != lastKey {
        _ = c.snapshot.save(p, newKey, nId, data)
    }
    return
}

func (c *Client) handleUpdated(p *aPath, lastKey string, nId int64) {
    data, err := c.loadNew(p, lastKey, nId)
    if err != nil {
        return
    }
    c.mutex.RLock()
    cb, ok := c.callbacks[p.String()]
    c.mutex.RUnlock()
    if ok && cb != nil {
        cb(data)
    }
}

// watch
func (c *Client) buildListenData(p *aPath, namespaces []string) string {
    type N struct {
        Namespace string `json:"namespaceName,omitempty"`
        ID        int64  `json:"notificationId,omitempty"`
    }
    var v []*N
    c.mutex.RLock()
    for _, ns := range namespaces {
        p.namespace = ns
        id, _ := c.notificationIDs[p.String()]
        v = append(v, &N{Namespace: ns, ID: id})
    }
    c.mutex.RUnlock()
    data, _ := json.Marshal(v)
    return string(data)
}
func (c *Client) listen(p *aPath) {
    type notification struct {
        Name string `json:"namespaceName,omitempty"`
        ID   int64  `json:"notificationId,omitempty"`
    }
    timeout := c.timeout * 10
    for {
        var namespaces []string
        c.mutex.RLock()
        namespaces, _ = c.watches[p.watchID()]
        c.mutex.RUnlock()

        if len(namespaces) == 0 {
            time.Sleep(timeout)
            continue
        }
        param := c.buildListenData(p, namespaces)
        var changes []notification

        server := c.selector.Select()
        args := []interface{}{
            server,
            url.QueryEscape(p.appId),
            url.QueryEscape(p.cluster),
            url.QueryEscape(param)}

        ctx, cancel := context.WithTimeout(context.Background(), timeout)
        result, err := c.httpc.Get(ctx, watchURL, nil, &changes, args...)
        if err != nil {
            if ctx.Err() != context.DeadlineExceeded {
                time.Sleep(timeout)
            }
            cancel()
            continue
        }
        cancel()
        switch result.Code {
        case http.StatusNotModified:
            continue
        case http.StatusOK:
        default:
            time.Sleep(timeout)
            continue
        }
        for _, n := range changes {
            p.namespace = n.Name
            var lastKey string
            c.mutex.Lock()
            profileId := p.String()
            lastKey = c.releaseKeys[profileId]
            c.notificationIDs[profileId] = n.ID
            c.mutex.Unlock()
            c.handleUpdated(p, lastKey, n.ID)
        }
    }
}

func NewClient(cfg *client.Config) (c client.Client, err error) {
    if cfg.Address == "" {
        return nil, errors.New("missing server info")
    }
    addrs := strings.Split(cfg.Address, ";")
    if len(addrs) == 0 {
        addrs = append(addrs, "httpc://127.0.0.1:8080")
    }
    if cfg.AppID == "" {
        return nil, errors.New("missing appId")
    }
    if cfg.Cluster == "" {
        cfg.Cluster = "default"
    }
    if cfg.Timeout == 0 {
        cfg.Timeout = time.Second
    }
    return &Client{
        appId:           cfg.AppID,
        cluster:         cfg.Cluster,
        timeout:         cfg.Timeout,
        selector:        NewRandom(addrs),
        snapshot:        newSnapshot(cfg.CacheDir),
        notificationIDs: make(map[string]int64),
        releaseKeys:     make(map[string]string),
        httpc:           newHttpClient(10*cfg.Timeout, nil),
        watches:         map[string][]string{},
        callbacks:       map[string]client.ChangedCallback{},
        watchDisabled:   cfg.WatchDisabled,
    }, nil
}
