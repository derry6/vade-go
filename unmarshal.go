package vade

import (
    "encoding"
    "fmt"
    "reflect"
    "strconv"
    "strings"
    "time"

    pkgerrs "github.com/pkg/errors"
    "github.com/spf13/cast"

    "github.com/derry6/vade-go/pkg/structinfo"
)

var (
    durationType = reflect.TypeOf(time.Duration(0))
    zeroValue    reflect.Value
)

type UnmarshalGet func(key string) (value interface{}, ok bool)

type UnmarshalOption func(d *decoder)

func WithUnmarshalTag(tag string) UnmarshalOption {
    return func(opts *decoder) {
        opts.tag = tag
    }
}

func WithUnmarshalPrefix(key string) UnmarshalOption {
    return func(opts *decoder) {
        opts.prefix = key
    }
}


type TypeError struct {
    Errors []string
}

func (e *TypeError) Error() string {
    return fmt.Sprintf("props: unmarshal errors:\n  %s",
        strings.Join(e.Errors, "\n  "))
}

type noPanicError struct {
    err error
}


type Unmarshaler interface {
    UnmarshalProps(unmarshal func(interface{}) error) error
}

type decoder struct {
    errs   []string
    keys   []string
    get UnmarshalGet
    tag    string
    prefix string
}

func (u *decoder) getValue(key string)(value interface{}, ok bool) {
    if u.get == nil {
        return nil, false
    }
    return u.get(key)
}

// error handler
func (u *decoder) handleErr(err *error) {
    if v := recover(); v != nil {
        if e, ok := v.(noPanicError); ok {
            *err = e.err
        } else {
            panic(v)
        }
    }
}
func (u *decoder) fail(err error) {
    panic(noPanicError{err})
}
func (u *decoder) failf(format string, args ...interface{}) {
    panic(noPanicError{fmt.Errorf("eacc: "+format, args...)})
}
func (u *decoder) addErr(err error, key string, out reflect.Value) {
    u.errs = append(u.errs, fmt.Sprintf("cannot unmarshal %s into %s: %v", key, out.Type(), err))
}

func (u *decoder) parseTag(tag string) (string, structinfo.TagOptions) {
    if idx := strings.Index(tag, ","); idx != -1 {
        return tag[:idx], structinfo.TagOptions(tag[idx+1:])
    }
    return tag, structinfo.TagOptions("")
}
func (u *decoder) findTag(t reflect.StructTag) (tag string, tagOpts structinfo.TagOptions) {
    useTag := u.tag
    if useTag != "" {
        return u.parseTag(t.Get(useTag))
    }
    for _, tagKey := range structinfo.SupportedTags {
        tag, tagOpts = u.parseTag(t.Get(tagKey))
        if tag == "-" {
            continue
        }
        if tagOpts.Contains("inline") {
            return
        }
        if tag != "" {
            return
        }
    }
    return
}

func (u *decoder) arraySize(key string) (n int) {
    if v, ok := u.getValue(key); ok {
        switch x := v.(type) {
        case int:
            if x <= 0 {
                x = 0
            }
            return x
        case int64:
            if x <= 0 {
                x = 0
            }
            return int(x)
        }
        return 0
    }
    for i := 0; i < 32; i++ {
        pre := key + "[" + strconv.Itoa(i) + "]"
        got := false
        for _, k := range u.keys {
            if pre == k || strings.HasPrefix(k, pre+".") {
                got = true
                break
            }
        }
        if got {
            n++
        } else {
            return
        }
    }
    return
}
func (u *decoder) childKeysOf(key string) (keys []string) {
    if key == "" {
        return u.keys
    }
    prefix := key + "."
    for _, k := range u.keys {
        if strings.HasPrefix(k, prefix) {
            keys = append(keys, k)
        }
    }
    return
}
func (u *decoder) childName(full string, key string) string {
    if key != "" {
        key += "."
    }
    name := strings.TrimPrefix(full, key)
    i := strings.Index(name, ".")
    if i >= 0 {
        name = name[:i]
    }
    return name
}
func (u *decoder) mergeKey(prefix, name string) string {
    if prefix != "" {
        return prefix + "." + name
    }
    return name
}

func (u *decoder) callUnmarshaler(key string, exu Unmarshaler) (good bool) {
    tErrLen := len(u.errs)
    err := exu.UnmarshalProps(func(v interface{}) (err error) {
        defer u.handleErr(&err)
        u.unmarshal(key, reflect.ValueOf(v))
        if len(u.errs) > tErrLen {
            issues := u.errs[tErrLen:]
            u.errs = u.errs[:tErrLen]
            return &TypeError{issues}
        }
        return nil
    })
    if e, ok := err.(*TypeError); ok {
        u.errs = append(u.errs, e.Errors...)
        return false
    }
    if err != nil {
        u.fail(err)
    }
    return true
}

func (u *decoder) indirectPtr(key string, out reflect.Value) (newOut reflect.Value, done, good bool) {
    again := true
    for again {
        again = false
        if out.Kind() == reflect.Ptr {
            if out.IsNil() {
                out.Set(reflect.New(out.Type().Elem()))
            }
            out = out.Elem()
            again = true
        }
        if out.CanAddr() {
            addr := out.Addr().Interface()
            if exu, ok := addr.(Unmarshaler); ok {
                good = u.callUnmarshaler(key, exu)
                return out, true, good
            }
            if tu, ok := addr.(encoding.TextUnmarshaler); ok {
                if value, _ := u.getValue(key); value != nil {
                    switch v := value.(type) {
                    case string:
                        if err := tu.UnmarshalText([]byte(v)); err != nil {
                            u.fail(err)
                        }
                        return out, true, good
                    case []byte:
                        if err := tu.UnmarshalText(v); err != nil {
                            u.fail(err)
                        }
                        return out, true, good
                    }
                }
            }
        }
    }
    return out, false, false
}
func (u *decoder) handleBasic(key string, out reflect.Value) bool {
    // finds value from getter
    value, _ := u.getValue(key)
    if value == nil {
        if out.Kind() == reflect.Map && !out.CanAddr() {
            // 设置map为零值
            for _, k := range out.MapKeys() {
                out.SetMapIndex(k, zeroValue)
            }
        } else {
            out.Set(reflect.Zero(out.Type()))
        }
        return true
    }
    if rv := reflect.ValueOf(value); out.Type() == rv.Type() {
        // out的类型和value的类型一致， 可以直接设置
        out.Set(rv)
        return true
    }
    // 类型不一致， 需要类型转换
    if out.CanAddr() {
        // 如果out实现了TextUnmarshaler接口，
        // 并且value是text类型: (string, []byte)
        // 则调用UnmarshalText进行设置。
        addr := out.Addr().Interface()
        tu, ok := addr.(encoding.TextUnmarshaler)
        if ok {
            switch x := value.(type) {
            case string:
                if err := tu.UnmarshalText([]byte(x)); err != nil {
                    u.fail(err)
                }
                return true
            case []byte:
                if err := tu.UnmarshalText(x); err != nil {
                    u.fail(err)
                }
                return true
            }
        }
    }
    var dstErr error

    if out.Type() == durationType {
        switch v := value.(type) {
        case string:
            d, err := toDuration(v)
            if err == nil {
                out.SetInt(int64(d))
                return true
            }
            dstErr = pkgerrs.Wrapf(err, "v = %x", v)
        }
        u.addErr(dstErr, key, out)
        return false
    }
    // 其他的基本类型， 作类型转换后进行设置。
    switch out.Kind() {
    case reflect.Interface:
        out.Set(reflect.ValueOf(value))
        return true
    case reflect.String:
        if x, err := cast.ToStringE(value); err == nil {
            out.SetString(x)
            return true
        } else {
            dstErr = err
        }
    case reflect.Int8, reflect.Int16, reflect.Int, reflect.Int32, reflect.Int64:
        if x, err := cast.ToInt64E(value); err == nil {
            out.SetInt(x)
            return true
        } else {
            dstErr = err
        }
        // todo: 考虑溢出的问题
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
        if x, err := cast.ToUint64E(value); err == nil {
            out.SetUint(x)
            return true
        } else {
            dstErr = err
        }
    case reflect.Bool:
        if b, err := cast.ToBoolE(value); err == nil {
            out.SetBool(b)
            return true
        } else {
            dstErr = err
        }
    case reflect.Float32, reflect.Float64:
        if f64, err := cast.ToFloat64E(value); err == nil {
            out.SetFloat(f64)
            return true
        } else {
            dstErr = err
        }
    case reflect.Struct:
        // should not reach here
        if rv := reflect.ValueOf(value); out.Type() == rv.Type() {
            out.Set(rv)
            return true
        } else {
            dstErr = pkgerrs.New("type error")
        }
    case reflect.Ptr:
        if out.Type().Elem() == reflect.TypeOf(value) {
            // todo: Does this make sense?
            //  When is out a Ptr except when decoding a nil value?
            elem := reflect.New(out.Type().Elem())
            elem.Elem().Set(reflect.ValueOf(value))
            out.Set(elem)
            return true
        } else {
            dstErr = pkgerrs.New("element type error")
        }
    }
    u.addErr(dstErr, key, out)
    return false
}
func (u *decoder) handleMap(key string, out reflect.Value) (good bool) {
    outt := out.Type()
    kt := outt.Key()
    et := outt.Elem()
    if kt.Kind() != reflect.String {
        u.failf("map key type must be string: %#v", kt)
    }
    if out.IsNil() {
        out.Set(reflect.MakeMap(outt))
    }
    children := u.childKeysOf(key)
    for _, childKey := range children {
        k := reflect.New(kt).Elem()
        name := u.childName(childKey, key)
        k.SetString(name)
        full := u.mergeKey(key, name)
        e := reflect.New(et).Elem()
        if u.unmarshal(full, e) {
            out.SetMapIndex(k, e)
        }
    }
    return true
}
func (u *decoder) handleSlice(key string, out reflect.Value) (good bool) {
    outType := out.Type()
    elemType := outType.Elem()
    elemSize := u.arraySize(key)
    slice := reflect.MakeSlice(outType, elemSize, elemSize)
    for i := 0; i < elemSize; i++ {
        name := key + "[" + strconv.Itoa(i) + "]"
        v := reflect.New(elemType)
        if good = u.unmarshal(name, v); !good {
        }
        slice.Index(i).Set(v.Elem())
    }
    out.Set(slice)
    return true
}

func (u *decoder) handleStruct(key string, out reflect.Value) (good bool) {
    sInfo, err := structinfo.Get(out.Type(), u.tag)
    if err != nil {
        panic(err)
    }
    children := u.childKeysOf(key)
    // not inline
    for name, info := range sInfo.FieldsMap {
        var field reflect.Value
        if info.Inline == nil {
            field = out.Field(info.Num)
        } else {
            field = out.FieldByIndex(info.Inline)
        }
        fullName := u.mergeKey(key, name)
        u.unmarshal(fullName, field)
    }
    // handle inlined
    if sInfo.InlineMapIndex == -1 {
        return true
    }
    // 如果结构体中包含inline成员
    var inlineMap reflect.Value
    var elemType reflect.Type
    inlineMap = out.Field(sInfo.InlineMapIndex)
    inlineMap.Set(reflect.New(inlineMap.Type()).Elem())
    elemType = inlineMap.Type().Elem()

    if inlineMap.IsNil() {
        inlineMap.Set(reflect.MakeMap(inlineMap.Type()))
    }
    // getter all fields not decoded to inlined field
    for _, name := range children {
        name = u.childName(name, key)
        if len(name) == 0 {
            continue
        }
        if _, ok := sInfo.FieldsMap[name]; ok {
            continue
        }
        elemKey := u.mergeKey(key, name)
        value := reflect.New(elemType).Elem()
        u.unmarshal(elemKey, value)
        inlineMap.SetMapIndex(reflect.ValueOf(name), value)
    }
    return true
}

func (u *decoder) unmarshal(key string, out reflect.Value) (good bool) {
    out, done, good := u.indirectPtr(key, out)
    if done {
        return good
    }
    switch out.Kind() {
    case reflect.Map:
        good = u.handleMap(key, out)
    case reflect.Slice:
        good = u.handleSlice(key, out)
    case reflect.Struct:
        good = u.handleStruct(key, out)
    default:
        good = u.handleBasic(key, out)
    }
    return good
}

func unmarshal(get UnmarshalGet, keys []string, out interface{}, opts ...UnmarshalOption) (err error) {
    v := reflect.ValueOf(out)
    if v.Kind() == reflect.Ptr && !v.IsNil() {
        v = v.Elem()
    }
    d := &decoder{get: get, keys: keys}
    for _, optFn := range opts {
        optFn(d)
    }
    n := len(d.prefix) - 1
    if len(d.prefix) > 0 && d.prefix[n] == '.' {
        d.prefix = d.prefix[:n]
    }
    defer d.handleErr(&err)
    d.unmarshal(d.prefix, v)
    if len(d.errs) > 0 {
        return &TypeError{d.errs}
    }
    return nil
}

func Unmarshal(out interface{}, opts ...UnmarshalOption) error {
    return unmarshal(_mgr.Get, _mgr.Keys(), out, opts...)
}