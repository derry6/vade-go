package structinfo

import (
    "errors"
    "fmt"
    "reflect"
    "strings"
    "sync"
)

// 从encoding/json拷贝
// StructInfo 结构体的信息
type StructInfo struct {
    FieldsMap      map[string]FieldInfo
    FieldsList     []FieldInfo
    InlineMapIndex int // 如果结构体包含inline map, 这是index。
}

type FieldInfo struct {
    Key       string
    Num       int
    OmitEmpty bool
    Flow      bool
    // Id holds the unique field identifier, so we can cheaply
    // check for field duplicates without maintaining an extra map.
    Id int
    // Inline holds the field index if the field is part of an inlined struct.
    Inline []int
}

var cache = make(map[reflect.Type]*StructInfo)
var cacheMu sync.RWMutex


func Get(st reflect.Type, useTag string) (*StructInfo, error) {
    cacheMu.RLock()
    sInfo, found := cache[st]
    cacheMu.RUnlock()
    if found {
        return sInfo, nil
    }
    n := st.NumField()
    fieldsMap := make(map[string]FieldInfo)
    fieldsList := make([]FieldInfo, 0, n)
    inlinedMapNum := -1
    for i := 0; i != n; i++ {
        field := st.Field(i)
        if field.PkgPath != "" && !field.Anonymous {
            continue // Private field
        }
        info := FieldInfo{Num: i}
        var tag string
        if useTag != "" {
            tag = field.Tag.Get(useTag)
            if tag == "" && strings.Index(string(field.Tag), ":") < 0 {
                tag = string(field.Tag)
            }
            if tag == "-" {
                continue
            }
        } else {
            for _, tn := range SupportedTags {
                tag = field.Tag.Get(tn)
                if tag == "" {
                    continue
                }
                break
            }
            if tag == "" && strings.Index(string(field.Tag), ":") < 0 {
                tag = string(field.Tag)
            }
            if tag == "-" {
                continue
            }
        }

        inline := false
        fields := strings.Split(tag, ",")
        if len(fields) > 1 {
            for _, flag := range fields[1:] {
                switch flag {
                case "omitempty":
                    info.OmitEmpty = true
                case "flow":
                    info.Flow = true
                case "inline":
                    inline = true
                default:
                    return nil,
                        errors.New(fmt.Sprintf("Unsupported flag %q in tag %q of type %s", flag, tag, st))
                }
            }
            tag = fields[0]
        }

        if inline {
            switch field.Type.Kind() {
            case reflect.Map:
                if inlinedMapNum >= 0 {
                    return nil, errors.New("multiple inline maps in struct " + st.String())
                }
                if field.Type.Key() != reflect.TypeOf("") {
                    return nil, errors.New("Option ,inline needs a map with string keys in struct " + st.String())
                }
                inlinedMapNum = info.Num
            case reflect.Struct:
                subInfo, err := Get(field.Type, useTag)
                if err != nil {
                    return nil, err
                }
                for _, subField := range subInfo.FieldsList {
                    if _, found = fieldsMap[subField.Key]; found {
                        msg := "Duplicated curKey '" + subField.Key + "' in struct " + st.String()
                        return nil, errors.New(msg)
                    }
                    if subField.Inline == nil {
                        subField.Inline = []int{i, subField.Num}
                    } else {
                        subField.Inline = append([]int{i}, subField.Inline...)
                    }
                    subField.Id = len(fieldsList)
                    fieldsMap[subField.Key] = subField
                    fieldsList = append(fieldsList, subField)
                }
            default:
                // return nil, errors.New("option ,inline needs a struct value or map field")
                return nil, errors.New("option ,inline needs a struct value field")
            }
            continue
        }

        if tag != "" {
            info.Key = tag
        } else {
            // 如果没有tag, 使用成员名称的 驼峰命名法 格式。
            info.Key = toCamelString(field.Name)
        }

        if _, found = fieldsMap[info.Key]; found {
            msg := "Duplicated curKey '" + info.Key + "' in struct " + st.String()
            return nil, errors.New(msg)
        }

        info.Id = len(fieldsList)
        fieldsList = append(fieldsList, info)
        fieldsMap[info.Key] = info
    }

    sInfo = &StructInfo{
        FieldsMap:      fieldsMap,
        FieldsList:     fieldsList,
        InlineMapIndex: inlinedMapNum,
    }

    cacheMu.Lock()
    cache[st] = sInfo
    cacheMu.Unlock()
    return sInfo, nil
}
