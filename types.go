package vade

import (
    "fmt"
    "time"

    "github.com/spf13/cast"
)

func onMustError(key string, val interface{}, dst interface{}) {
    panic(fmt.Sprintf("mustError: value of key %q is %v(%T), want %T", key, val, val, dst))
}

// Int 获取int值
func Int(key string, def int) int {
    if v, ok := Get(key); ok {
        if x, err := cast.ToIntE(v); err == nil {
            return x
        }
    }
    return def
}

// MustInt 获取int值
func MustInt(key string) int {
    var (
        v interface{}
        ok bool
    )
    if v, ok = Get(key); ok {
        if x, err := cast.ToIntE(v); err == nil {
            return x
        }
    }
    onMustError(key, v, int(0))
    return 0
}

// Float 获取float值
func Float(key string, def float64) float64 {
    if v, ok := Get(key); ok {
        if x, err := cast.ToFloat64E(v); err == nil {
            return x
        }
    }
    return def
}

// MustFloat 获取float值
func MustFloat(key string) float64 {
    var (
        v interface{}
        ok bool
    )
    if v, ok = Get(key); ok {
        if x, err := cast.ToFloat64E(v); err == nil {
            return x
        }
    }
    onMustError(key, v, 0.0)
    return 0.0
}

// String 获取 string 值
func String(key string, def string) string {
    if v, ok := Get(key); ok {
        if x, err := cast.ToStringE(v); err == nil {
            return x
        }
    }
    return def
}
// MustString 获取 string 值
func MustString(key string) string {
    var (
        v interface{}
        ok bool
    )
    if v, ok = Get(key); ok {
        if x, err := cast.ToStringE(v); err == nil {
            return x
        }
    }
    onMustError(key, v, "")
    return ""
}

// Bool 获取 bool 值
func Bool(key string, def bool) bool {
    if v, ok := Get(key); ok {
        if x, err := cast.ToBoolE(v); err == nil {
            return x
        }
    }
    return def
}
// MustBool 获取 bool 值
func MustBool(key string) bool {
    var (
        v interface{}
        ok bool
    )
    if v, ok = Get(key); ok {
        if x, err := cast.ToBoolE(v); err == nil {
            return x
        }
    }
    onMustError(key, v, false)
    return false
}

// Duration 获取 duration 值
func Duration(key string, def time.Duration) time.Duration {
    if v, ok := Get(key); ok {
        if x, err := cast.ToDurationE(v); err == nil {
            return x
        }
    }
    return def
}

// MustDuration 获取 duration 值
func MustDuration(key string) time.Duration {
    var (
        v interface{}
        ok bool
    )
    if v, ok = Get(key); ok {
        if x, err := cast.ToDurationE(v); err == nil {
            return x
        }
    }
    onMustError(key, v, time.Duration(0))
    return time.Duration(0)
}

// Time 获取 time 值
func Time(key string, def time.Time) time.Time {
    if v, ok := Get(key); ok {
        if x, err := cast.ToTimeE(v); err == nil {
            return x
        }
    }
    return def
}

// MustTime 获取 time 值
func MustTime(key string) time.Time {
    var (
        v interface{}
        ok bool
    )
    if v, ok = Get(key); ok {
        if x, err := cast.ToTimeE(v); err == nil {
            return x
        }
    }
    onMustError(key, v, time.Time{})
    return time.Time{}
}