package client

import (
    "context"
    "os"
    "strings"

    "gopkg.in/yaml.v2"
)

const (
    Env = "env"
)

var (
    _ Client = (*envClient)(nil)
)

func init() {
    _ = RegisterClient(Env, newEnvClient)
}

type envClient struct{}

func (c *envClient) Close() error { return nil }
func (c *envClient) Pull(ctx context.Context, path string) (data []byte, err error) {
    ps := map[string]interface{}{}
    environ := os.Environ()
    for _, item := range environ {
        fullExpr := []rune(item)
        idx := strings.Index(item, "=")
        envKey := string(fullExpr[0:idx])
        propsKey := strings.Replace(envKey, "_", ".", -1)
        propsKey = strings.ToLower(propsKey)
        // 保存原始配置和变换后的配置
        ps[envKey] = string(fullExpr[idx+1:])
        if propsKey[0] != '.' {
            ps[propsKey] = ps[envKey]
        }
    }
    data, err = yaml.Marshal(ps)
    return
}
func (c *envClient) Push(ctx context.Context, path string, data []byte) error { return nil }
func (c *envClient) Watch(path string, cb ChangedCallback) error              { return nil }

func newEnvClient(cfg *Config) (Client, error) {
    return &envClient{}, nil
}
