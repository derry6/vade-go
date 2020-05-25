package vade

import "github.com/derry6/vade-go/source"

var (
	_mgr Manager
)

func init() {
	_mgr, _ = newManager()
}

// NewManager create manager instance
func NewManager(opts ...Option) (Manager, error) {
	return newManager(opts...)
}

// AddSource 添加source
func AddSource(s source.Source) error {
	return _mgr.AddSource(s)
}

// Source 返回指定名称的source, 或者错误
func Source(name string) (source.Source, error) {
	return _mgr.Source(name)
}

// Sources 返回添加的所有source
func Sources() []source.Source {
	return _mgr.Sources()
}

func AddPath(source string, path string, opts ...source.PathOption) error {
	return _mgr.AddPath(source, path, opts...)
}

// All 返回所有的kv
func All() map[string]interface{} {
	return _mgr.All()
}

// Keys 返回所有的key
func Keys() []string {
	return _mgr.Keys()
}

// Get 获取指定key的value, 不存在时ok为false
func Get(key string) (value interface{}, ok bool) {
	return _mgr.Get(key)
}

// Set 设置kv， 覆盖配置。
func Set(key string, value interface{}) {
	_mgr.Set(key, value)
}

// SetDefault 设置默认值
func SetDefault(key string, value interface{}) {
	_mgr.SetDefault(key, value)
}

// Delete 删除覆盖的key和默认key
func Delete(key string) {
	_mgr.Delete(key)
}

// Watch 监听某个满足pattern模式的key变化的事件。
func Watch(pattern string, cb EventHandler) (id int64) {
	return _mgr.Watch(pattern, cb)
}

// Unwatch 取消监听
func Unwatch(id int64) {
	_mgr.Unwatch(id)
}
