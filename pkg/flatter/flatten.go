package flatter

import (
    "fmt"
    "reflect"
)

var (
    _flatter = &flatter{}
)
type flatter struct{}

func (f *flatter) do(curKey string, value interface{}, dst map[string]interface{}) (err error) {
    var nextMap map[string]interface{}
    switch value.(type) {
    case map[string]interface{}: // json
        nextMap = value.(map[string]interface{})
    case map[interface{}]interface{}: // yaml
        nextMap = make(map[string]interface{})
        for k, v := range value.(map[interface{}]interface{}) {
            if key, ok := k.(string); ok {
                nextMap[key] = v
            } else {
                return fmt.Errorf("map key '%#v' is not string", k)
            }
        }
    case map[string]string:
        nextMap = make(map[string]interface{})
        for k, v := range value.(map[string]string) {
            nextMap[k] = v
        }
    case []byte:
        dst[curKey] = string(value.([]byte))
    case []interface{}:
        arr := value.([]interface{})
        if err = f.doArray(curKey, arr, dst); err != nil {
            return err
        }
    default:
        dst[curKey] = value
    }
    return f.doMap(curKey, nextMap, dst)
}

func (f *flatter) doArray(curKey string, value []interface{}, out map[string]interface{}) (err error) {
    subKey := ""
    for i := 0; i < len(value); i++ {
        subKey = fmt.Sprintf("%s[%d]", curKey, i)
        if err = f.do(subKey, value[i], out); err != nil {
            return err
        }
    }
    out[curKey] = len(value)
    return nil
}

func (f *flatter) doMap(parent string, src, dst map[string]interface{}) (err error) {
    for key, value := range src {
        fullKey := mergeKey(parent, key)
        if err = f.do(fullKey, value, dst); err != nil {
            return
        }
    }
    return err
}

// Flatten 多级的map转为properties形式
func Flatten(src, dst map[string]interface{}, prefix string, useReflect bool) (err error) {
    if useReflect {
        vIn := reflect.ValueOf(src)
        return _reflectFlatter.do(prefix, vIn, dst)
    }
    return _flatter.doMap(prefix, src, dst)
}
