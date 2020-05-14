package vade

import "github.com/derry6/vade-go/source"

var (
    mgr Manager = newManager()
)

// AddSource 添加source
func AddSource(s source.Source) error {
    return mgr.AddSource(s)
}

// Source 返回指定名称的source, 或者错误
func Source(name string) (source.Source, error) {
    return mgr.Source(name)
}

// Sources 返回添加的所有source
func Sources() []source.Source {
    return mgr.Sources()
}

func AddPath(source string, path string, opts ...source.PathOption) error {
    return mgr.AddPath(source, path, opts...)
}

// All 返回所有的kv
func All() map[string]interface{} {
    return mgr.All()
}

// Keys 返回所有的key
func Keys() []string {
    return mgr.Keys()
}

// Get 获取指定key的value, 不存在时ok为false
func Get(key string) (value interface{}, ok bool) {
    return mgr.Get(key)
}

// Set 设置kv， 覆盖配置。
func Set(key string, value interface{}) {
    mgr.Set(key, value)
}

// SetDefault 设置默认值
func SetDefault(key string, value interface{}) {
    mgr.SetDefault(key, value)
}

// Delete 删除覆盖的key和默认key
func Delete(key string) {
    mgr.Delete(key)
}

// Watch 监听某个满足pattern模式的key变化的事件。
func Watch(pattern string, cb EventHandler) (id int64) {
    return mgr.Watch(pattern, cb)
}

// Unwatch 取消监听
func Unwatch(id int64) {
    mgr.Unwatch(id)
}
