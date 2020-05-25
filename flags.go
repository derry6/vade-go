package vade

import (
	"flag"
	"os"
	"path/filepath"
	"time"
)

// Global flags
var (
	_flagSet   = flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ExitOnError)
	_bindFlags = map[string]interface{}{}
)

// Flag 定义命令行参数
func Flag(key string, def interface{}, usage string) {
	switch def.(type) {
	case int:
		_bindFlags[key] = _flagSet.Int(key, def.(int), usage)
	case uint:
		_bindFlags[key] = _flagSet.Uint(key, def.(uint), usage)
	case bool:
		_bindFlags[key] = _flagSet.Bool(key, def.(bool), usage)
	case string:
		_bindFlags[key] = _flagSet.String(key, def.(string), usage)
	case time.Duration:
		_bindFlags[key] = _flagSet.Duration(key, def.(time.Duration), usage)
	case float64:
		_bindFlags[key] = _flagSet.Float64(key, def.(float64), usage)
	case float32:
		_bindFlags[key] = _flagSet.Float64(key, float64(def.(float32)), usage)
	}
}
