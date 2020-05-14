package client

import (
    "crypto/tls"
    "flag"
    "net"
    "os"
    "strings"
    "time"

    "github.com/hashicorp/go-rootcerts"
)

// Config client configurations
type Config struct {
    // env
    Flags   map[string]interface{} `json:"-"`
    FlagSet *flag.FlagSet          `json:"-"`
    // common
    Timeout       time.Duration `json:"timeout,omitempty" yaml:"timeout"`
    WatchTimeout  time.Duration `json:"watchTimeout,omitempty" yaml:"watchTimeout"`
    WatchDisabled bool          `json:"watchDisabled,omitempty" yaml:"watchDisabled"`
    LogDir        string        `json:"logDir,omitempty" yaml:"logDir"`
    CacheDir      string        `json:"cacheDir,omitempty" yaml:"cacheDir"`
    // remote
    Endpoint string                 `json:"endpoint,omitempty" yaml:"endpoint"`
    Address  string                 `json:"address,omitempty" yaml:"address"`
    AppID    string                 `json:"appId,omitempty" yaml:"appId"`
    Cluster  string                 `json:"cluster,omitempty" yaml:"cluster"`
    RegionID string                 `json:"regionId,omitempty" yaml:"regionId"`
    Extra    map[string]interface{} `json:"extra,omitempty" yaml:"extra"`
    // ssl
    SSLEnabled    bool   `json:"sslEnabled,omitempty" yaml:"sslEnabled"`
    SSLCACert     string `json:"sslCACert,omitempty" yaml:"sslCACert"`
    SSLCert       string `json:"sslCert,omitempty" yaml:"sslCert"`
    SSLKey        string `json:"sslKey,omitempty" yaml:"sslKey"`
    SSLVerifyPeer bool   `json:"sslVerifyPeer,omitempty" yaml:"sslVerifyPeer"`
    ServerName    string `json:"serverName,omitempty" yaml:"serverName"`
    // auth
    Username   string `json:"username,omitempty" yaml:"username"`
    Password   string `json:"password,omitempty" yaml:"password"`
    AccessKey  string `json:"accessKey,omitempty" yaml:"accessKey"`
    SecretKey  string `json:"secretKey,omitempty" yaml:"secretKey"`
    Token      string `json:"token,omitempty" yaml:"token"`
    DataCenter string `json:"dataCenter,omitempty" yaml:"dataCenter"`
    Namespace  string `json:"namespace,omitempty" yaml:"namespace"`
    Group      string `json:"group,omitempty" yaml:"group"`
}

func DefaultConfig() *Config {
    return &Config{
        Timeout:       1 * time.Second,
        WatchDisabled: false,
    }
}

func (c *Config) getServerName() string {
    if c.ServerName != "" {
        return c.ServerName
    }
    if c.Address == "" {
        return ""
    }
    addr := strings.Split(c.Address, ",")[0]
    hasPort := strings.LastIndex(addr, ":") > strings.LastIndex(addr, "]")
    if hasPort {
        var err error
        addr, _, err = net.SplitHostPort(addr)
        if err != nil {
            return ""
        }
    }
    c.ServerName = addr
    return addr
}

func (c *Config) TLSConfig() *tls.Config {
    tlsConfig := &tls.Config{InsecureSkipVerify: !c.SSLVerifyPeer}
    tlsConfig.ServerName = c.getServerName()

    if c.SSLCert != "" && c.SSLKey != "" {
        tlsCert, err := tls.LoadX509KeyPair(c.SSLCert, c.SSLKey)
        if err != nil {
            return nil
        }
        tlsConfig.Certificates = []tls.Certificate{tlsCert}
    } else {
        return nil
    }

    if c.SSLCACert != "" {
        rootConfig := &rootcerts.Config{}
        stat, err := os.Stat(c.SSLCACert)
        if err == nil {
            if stat.IsDir() {
                rootConfig.CAPath = c.SSLCACert
            } else {
                rootConfig.CAFile = c.SSLCACert
            }
        }
        if err := rootcerts.ConfigureTLS(tlsConfig, rootConfig); err != nil {
            return nil
        }
    }
    return tlsConfig
}
