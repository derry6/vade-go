package expander

import (
    "strings"

    pkgerrs "github.com/pkg/errors"
    "github.com/spf13/cast"
)

type Getter func(key string) (value interface{}, ok bool)
type Handler func(in string) (interface{}, error)

type Expander interface {
    Expand(in string) (out interface{}, err error)
}

// 定义一个替换的区间[a, b]
type pattern struct {
    a, b   string
    la, lb int // a,b本身的长度
    cb Handler
}

type options struct {
    patterns map[string]pattern
}

type Option func(opts *options)

func WithExpansion(pre, post string, handler Handler) Option {
    return func(opts *options) {
        if pre != "" && post != "" && handler != nil {
            key := pre + ":" + post
            if _, ok := opts.patterns[key]; !ok {
                opts.patterns[key] = pattern{a: pre, b: post, la: len(pre), lb: len(post), cb: handler}
            }
        }
    }
}

func newOptions(opts ...Option) *options {
    eOpts := &options{patterns: map[string]pattern{}}
    for _, optFn := range opts {
        optFn(eOpts)
    }
    return eOpts
}

type expander struct {
    get  Getter
    opts *options
    pending map[string]bool
    cache   map[string]interface{}
}

func (e *expander) doReplace(in string) (result interface{}, err error) {
    // 默认需要替换的内容是一个key, 需要继续替换该key
    dstKey := strings.TrimSpace(in)
    return e.doExpand(dstKey)
}

func (e *expander) valueOf(key string) (v interface{}, ok bool) {
    if v, ok = e.cache[key]; ok {
        return
    }
    return e.get(key)
}

func (e *expander) hasNext(value string, ep pattern) (start, end int, has bool) {
    p1 := strings.Index(value, ep.a)
    if p1 < 0 {
        return 0, 0, false
    }
    p2 := strings.Index(value[p1+ep.la:], ep.b)
    if p2 < 0 {
        return 0, 0, false
    }
    return p1, p1 + ep.la + p2, true
}

func (e *expander) doOne(value string, ep pattern) (interface{}, bool, error) {
    var (
        expanded   = false
        start, end = 0, 0
        ok         = false
        builder    = strings.Builder{}
    )
    for {
        start, end, ok = e.hasNext(value, ep)
        if !ok {
            builder.WriteString(value)
            return builder.String(), expanded, nil
        }
        // 需要替换的内容, 也就是 ep.a 和 ep.b之间的内容
        src := value[start+ep.la : end]

        // 调用callback进行替换
        dst, err := ep.cb(src)
        if err != nil {
            return dst, expanded, err
        }
        expanded = true

        // 只有一个部分: 如 a=${b}
        if start == 0 && end+ep.lb >= len(value) {
            return dst, expanded, nil
        }
        // 如: a=23490${b}797402
        // 包含多个部分， 肯定是个string
        builder.WriteString(value[:start])
        dstStr, err := cast.ToStringE(dst)
        if err != nil {
            return dst, expanded, pkgerrs.Errorf("value of %q must be string", src)
        }
        builder.WriteString(dstStr)
        // 处理下一部分内容
        value = value[end+ep.lb:]
    }
}

// b a
func (e *expander) doExpand(key string) (interface{}, error) {
    // 是否存在循环依赖
    if _, ok := e.pending[key]; ok {
        return nil, pkgerrs.Errorf("circular dependency was detected: %s", key)
    }
    raw, ok := e.valueOf(key)
    if !ok {
        return nil, pkgerrs.Errorf("key %q not exists", key)
    }
    switch raw.(type) {
    case string:
    default:
        return raw, nil
    }
    value := raw.(string)
    expanded := false
    e.pending[key] = true // 还未得到结果
    for _, p := range e.opts.patterns {
        result, changed, err := e.doOne(value, p)
        if err != nil {
            return nil, err
        }
        if !changed {
            continue
        }
        expanded = true
        switch result.(type) {
        case string:
            value = result.(string)
        default:
            e.cache[key] = result
            delete(e.pending, key)
            return result, nil
        }
    }
    if expanded {
        e.cache[key] = value
    }
    delete(e.pending, key)
    return value, nil
}

func (e *expander) Expand(key string) (interface{}, error) {
    e.cache = make(map[string]interface{})
    e.pending = make(map[string]bool)
    return e.doExpand(key)
}


// New create expander instance
func New(get Getter, opts ...Option) Expander {
    e := &expander{
        get:     get,
        pending: make(map[string]bool),
        cache:   make(map[string]interface{}),
    }
    opts = append(opts, WithExpansion("${", "}", e.doReplace))
    e.opts = newOptions(opts...)
    return e
}
