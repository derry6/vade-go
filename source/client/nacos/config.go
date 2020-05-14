package nacos

import (
    "net/url"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"

    nacosconts "github.com/nacos-group/nacos-sdk-go/common/constant"

    "github.com/derry6/vade-go/source/client"
)

const (
    defaultNamespace = "public"
    defaultGroup     = "DEFAULT_GROUP"
)

func getServerConfigs(cfg *client.Config) (nscs []nacosconts.ServerConfig) {
    addrs := strings.Split(cfg.Address, ";")
    for _, s := range addrs {
        if !strings.HasPrefix(s, "http://") &&
            !strings.HasPrefix(s, "https://") {
            s = "http://" + s
        }
        var nsc nacosconts.ServerConfig
        URL, err := url.Parse(s)
        if err != nil {
            continue
        }
        nsc.ContextPath = URL.Path
        if nsc.ContextPath == "" {
            nsc.ContextPath = "/nacos"
        }
        nsc.IpAddr = URL.Hostname()
        port, err := strconv.Atoi(URL.Port())
        if err != nil || port == 0 {
            port = 8848
        }
        nsc.Port = uint64(port)
        nscs = append(nscs, nsc)
    }
    if len(nscs) == 0 {
        nscs = append(nscs, nacosconts.ServerConfig{
            ContextPath: "/nacos",
            IpAddr:      "127.0.0.1",
            Port:        8848,
        })
    }
    return
}

func getClientConfig(cfg *client.Config) (ncc nacosconts.ClientConfig) {
    if cfg.Timeout == 0 {
        cfg.Timeout = time.Second
    }
    interval := 10 * cfg.Timeout
    ncc.TimeoutMs = uint64(cfg.Timeout.Nanoseconds() / 1e6)
    if interval >= 30*time.Second {
        interval = 30 * time.Second
        if interval < cfg.Timeout {
            interval = cfg.Timeout
        }
    }
    ncc.ListenInterval = uint64(interval.Nanoseconds() / 1e6)
    ncc.NotLoadCacheAtStart = false
    ncc.UpdateCacheWhenEmpty = false
    ncc.LogDir = cfg.LogDir
    if ncc.LogDir == "" {
        ncc.LogDir = "logs"
    }
    defaultCacheDir := filepath.Join(os.TempDir(), "eacc-snapshots", "nacos")
    if cfg.CacheDir == "" {
        cfg.CacheDir = defaultCacheDir
    }
    ncc.CacheDir = cfg.CacheDir
    ncc.Endpoint = cfg.Endpoint
    ncc.AccessKey = cfg.AccessKey
    ncc.SecretKey = cfg.SecretKey
    if cfg.AccessKey != "" && cfg.SecretKey != "" {
        ncc.OpenKMS = true
    }
    ncc.RegionId = cfg.RegionID
    return
}
