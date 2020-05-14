package flatter

import (
    "fmt"
    "reflect"
)

var (
    _reflectFlatter = &reflectFlatter{}
)

type reflectFlatter struct {}

func (rf *reflectFlatter) doMap(prefix string, raw reflect.Value, out map[string]interface{}) (err error) {
    var key string
    for iter := raw.MapRange(); iter.Next(); {
        rkv := iter.Key()
        switch rkv.Kind() {
        case reflect.String:
            key = rkv.String()
        case reflect.Interface:
            elem := rkv.Elem()
            if elem.Kind() != reflect.String {
                return fmt.Errorf("key of map must string, but %v", elem.Kind())
            }
            key = elem.String()
        default:
            return fmt.Errorf("key of map must string, but %v", rkv.Kind())
        }
        err = rf.do(mergeKey(prefix, key), iter.Value(), out)
        if err != nil {
            return
        }
    }
    return err
}

func (rf *reflectFlatter) doArray(prefix string, raw reflect.Value, out map[string]interface{}) (err error) {
    var (
        subKey string
    )
    nums := raw.Len()
    for i := 0; i < nums; i++ {
        subKey = fmt.Sprintf("%s[%d]", prefix, i)
        item := raw.Index(i)
        if err = rf.do(subKey, item, out); err != nil {
            return err
        }
    }
    out[prefix] = nums
    return nil
}

func (rf *reflectFlatter) do(prefix string, src reflect.Value, dst map[string]interface{}) (err error) {
    switch src.Kind() {
    case reflect.Interface:
        return rf.do(prefix, src.Elem(), dst)
    case reflect.Map:
        return rf.doMap(prefix, src, dst)
    case reflect.Slice:
        return rf.doArray(prefix, src, dst)
    case reflect.Bool,
        reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
        reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
        reflect.Float32, reflect.Float64,
        reflect.String:
        if src.CanInterface() {
            dst[prefix] = src.Interface()
        } else {
            return fmt.Errorf("unsupported type: %v", src.Kind())
        }
    default:
        return fmt.Errorf("unsupported type: %v", src.Kind())
    }
    return nil
}
