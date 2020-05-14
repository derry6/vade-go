package nacos

import (
    "context"
    "crypto/md5"
    "encoding/hex"
    "io/ioutil"
    "sync"

    nacosconts "github.com/nacos-group/nacos-sdk-go/common/constant"
    "github.com/pkg/errors"

    "github.com/nacos-group/nacos-sdk-go/clients"
    nacoscc "github.com/nacos-group/nacos-sdk-go/clients/config_client"
    nacosvo "github.com/nacos-group/nacos-sdk-go/vo"

    stdlog "log"

    "github.com/derry6/vade-go/source/client"
)

const (
    Name = "nacos"
)

var (
    _ client.Client = (*Client)(nil)
)

func init() {
    _ = client.RegisterClient(Name, NewClient)
}

type Client struct {
    namespace     string
    group         string
    watchDisabled bool
    mutex         sync.RWMutex
    client        nacoscc.IConfigClient
    callbacks     map[string]client.ChangedCallback
}

func (c *Client) Close() error { return nil }
func (c *Client) Pull(ctx context.Context, path string) (data []byte, err error) {
    p := newPath(path, c.group, c.namespace)
    if p.dataId == "" {
        return nil, errors.New("unknown nacos path")
    }
    param := nacosvo.ConfigParam{DataId: p.dataId, Group: p.group}
    content, err := c.client.GetConfig(param)
    return []byte(content), err
}

func (c *Client) Push(ctx context.Context, path string, data []byte) error {
    p := newPath(path, c.group, c.namespace)
    if p.dataId == "" {
        return errors.New("unknown nacos path")
    }
    param := nacosvo.ConfigParam{
        DataId:  p.dataId,
        Group:   c.group,
        Content: string(data),
    }
    _, err := c.client.PublishConfig(param)
    return err
}

func (c *Client) Watch(path string, cb client.ChangedCallback) error {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    p := newPath(path, c.group, c.namespace)
    if c.watchDisabled {
        return nil
    }
    paId := p.String()
    if _, ok := c.callbacks[paId]; ok {
        return nil
    }
    c.callbacks[paId] = cb
    param := nacosvo.ConfigParam{
        DataId:   p.dataId,
        Group:    p.group,
        OnChange: c.handleUpdated,
    }
    if err := c.client.ListenConfig(param); err != nil {
        return err
    }
    return nil
}

func (c *Client) md5(data string) (md5sum string) {
    m5 := md5.New()
    m5.Write([]byte(data))
    return hex.EncodeToString(m5.Sum(nil))
}

func (c *Client) handleUpdated(namespace, group, dataId, data string) {
    p := nPath{group: group, namespace: namespace, dataId: dataId}
    if p.namespace == "" {
        p.namespace = defaultNamespace
    }
    if p.group == "" {
        p.group = defaultGroup
    }
    c.mutex.Lock()
    cb, _ := c.callbacks[p.String()]
    c.mutex.Unlock()
    if cb != nil {
        cb([]byte(data))
    }
}

func NewClient(cfg *client.Config) (client.Client, error) {
    settings := make(map[string]interface{})
    if cfg.Address == "" {
        return nil, errors.New("missing server info")
    }

    sConfigs := getServerConfigs(cfg)
    cConfigs := getClientConfig(cfg)

    namespaceId := cfg.Namespace
    if namespaceId != "public" {
        cConfigs.NamespaceId = namespaceId
    }
    group := cfg.Group
    if group == "" {
        group = defaultGroup
    }
    settings[nacosconts.KEY_SERVER_CONFIGS] = sConfigs
    settings[nacosconts.KEY_CLIENT_CONFIG] = cConfigs

    stdlog.SetOutput(ioutil.Discard)
    configClient, err := clients.CreateConfigClient(settings)
    if err != nil {
        return nil, err
    }
    s := &Client{
        group:         group,
        namespace:     namespaceId,
        client:        configClient,
        callbacks:     make(map[string]client.ChangedCallback),
        watchDisabled: cfg.WatchDisabled,
    }
    return s, nil
}
