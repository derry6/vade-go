package apollo

import (
    "context"
    "fmt"
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
    watchURLFmt = "%s/notifications/v2"
    getURLFmt   = "%s/configs/%s/%s/%s"
    defaultHost = "http://localhost:8080"
)

var (
    errNotFound    = errors.New("not found")
    errNotModified = errors.New("not modified")
)

var (
    _ client.Client = (*Client)(nil)
)

func init() {
    _ = client.RegisterClient(Name, NewClient)
}

// appId@cluster
type Client struct {
    appId         string
    cluster       string
    token         string
    timeout       time.Duration
    cli           *httpClient
    selector      Selector
    snapshot      *Snapshot
    mutex         sync.RWMutex
    nIDs          map[string]int64                  // namespace@cluster@app -> id
    releases      map[string]string                 // namespace@cluster@app -> id
    callbacks     map[string]client.ChangedCallback // namespace@cluster@app -> callbacks
    watches       map[string][]string               // cluster@app -> namespaces
    watchDisabled bool
}

func (c *Client) Close() error { return nil }

func (c *Client) setRelease(fullKey string, release string, nId int64) {
    c.mutex.Lock()
    if release != "" {
        c.releases[fullKey] = release
    }
    if nId != 0 {
        c.nIDs[fullKey] = nId
    }
    c.mutex.Unlock()
}
func (c *Client) lastReleaseOf(fullKey string) (release string) {
    c.mutex.RLock()
    release, _ = c.releases[fullKey]
    c.mutex.RUnlock()
    return release
}

func (c *Client) Pull(ctx context.Context, path string) ([]byte, error) {
    var (
        release string
        data    []byte
        err     error
        nId     int64
    )
    p := buildConfigPath(path, c.cluster, c.appId, c.token)
    fullKey := p.fullKey()
    release, data, err = c.getConfig(p, "")
    if err != nil {
        log.Get().Warnf("Can not load apollo config for %q: %v", fullKey, err)
        serverErr := err
        if data, err = c.snapshot.get(p); err != nil {
            return nil, serverErr
        }
        release, nId = c.snapshot.getReleaseKey(p)
        c.setRelease(fullKey, release, nId)
        return data, nil
    }
    c.setRelease(fullKey, release, 0)
    return data, err
}

func (c *Client) Push(ctx context.Context, path string, data []byte) error {
    return errors.New("not implement yet")
}

func (c *Client) Watch(path string, cb client.ChangedCallback) error {
    p := buildConfigPath(path, c.cluster, c.appId, c.token)
    c.mutex.Lock()
    defer c.mutex.Unlock()
    if c.watchDisabled {
        return errors.New("watch disabled")
    }
    c.callbacks[p.fullKey()] = cb
    found := false

    watchKey := p.watchKey()
    namespaces, ok := c.watches[watchKey]
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
        c.watches[watchKey] = namespaces
    }
    if !ok {
        go c.listen(p)
    }
    return nil
}

// get config from apollo server
func (c *Client) getConfig(p *configPath, release string) (newRelease string, content []byte, err error) {
    var (
        values = url.Values{}
        req    *http.Request
        result *httpResult
    )
    appId := url.QueryEscape(p.appId)
    cluster := url.QueryEscape(p.cluster)
    ns := url.QueryEscape(p.namespace)

    values.Set("releaseKey", release)
    server := c.selector.Select()
    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()

    reqURL := fmt.Sprintf(getURLFmt, server, appId, cluster, ns)
    req, err = c.cli.NewRequest(http.MethodGet, reqURL, values)
    if err != nil {
        return
    }
    sign(req, p.appId, p.token)

    result, err = c.cli.Raw(ctx, req)
    if err != nil {
        return "", nil, err
    }
    switch result.Code {
    case http.StatusNotFound:
        return release, nil, errNotFound
    case http.StatusOK:
        break
    case http.StatusNotModified:
        return release, nil, errNotModified
    default:
        return release, nil, errors.Errorf("status code is %d", result.Code)
    }
    newRelease, content, err = c.parseResponse(p, result.Body)
    if err != nil {
        return "", nil, err
    }
    return newRelease, content, err
}

func (c *Client) onConfigChanged(p *configPath, nId int64) {
    fullKey := p.fullKey()
    lastKey := c.lastReleaseOf(fullKey)
    newKey, content, err := c.getConfig(p, lastKey)
    if err != nil {
        log.Get().Errorf("Can not load apollo configs for %q: %v", fullKey, err)
        return
    }
    c.setRelease(fullKey, newKey, nId)
    if newKey != lastKey {
        _ = c.snapshot.save(p, newKey, nId, content)
    }
    c.mutex.RLock()
    cb, _ := c.callbacks[fullKey]
    c.mutex.RUnlock()
    if cb != nil {
        cb(content)
    }
}

func validateConfig(cfg *client.Config) (addresses []string, err error) {
    if cfg.Address == "" {
        return nil, errors.New("missing server info")
    }
    addresses = strings.Split(cfg.Address, ";")
    if len(addresses) == 0 {
        addresses = append(addresses, defaultHost)
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
    return addresses, nil
}

func NewClient(cfg *client.Config) (client.Client, error) {
    addrs, err := validateConfig(cfg)
    if err != nil {
        return nil, err
    }
    return &Client{
        appId:         cfg.AppID,
        cluster:       cfg.Cluster,
        timeout:       cfg.Timeout,
        token:         cfg.Token,
        selector:      NewRandom(addrs),
        snapshot:      newSnapshot(cfg.CacheDir),
        nIDs:          make(map[string]int64),
        releases:      make(map[string]string),
        cli:           newHttpClient(),
        watches:       map[string][]string{},
        callbacks:     map[string]client.ChangedCallback{},
        watchDisabled: cfg.WatchDisabled,
    }, nil
}
