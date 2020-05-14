package source

import (
    "fmt"
    "io"
    "math/rand"

    "github.com/derry6/vade-go/source/client"
)

// Source 管理客户端的多个配置集合, 目前配置集没有优先级
type Source interface {
    io.Closer
    Name() string
    Client() client.Client

    // 优先级
    Priority() int

    // 获取所有的key
    Keys() (keys []string)
    // 获取所有的配置
    All() (values map[string]interface{})
    // 获取单个配置
    Get(key string) (value interface{}, ok bool)
    // 设置配置
    Set(key string, value interface{})
    // 添加配置集合
    AddPath(path string, opts ...PathOption) (err error)
    // 设置回调
    OnEvents(cb func([]*Event))
}

func buildRandomName() string {
    return fmt.Sprintf("source-%d", rand.Int())
}

func New(name string, c client.Client, opts ...Option) Source {
    sOpts := newOptions(opts...)
    if name == "" {
        name = buildRandomName()
    }
    return newBaseSource(name, c, sOpts)
}
