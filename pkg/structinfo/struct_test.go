package structinfo

import (
    "reflect"
    "testing"
)

func TestGetStructInfo(t *testing.T) {
    type User struct {
        ID        string `json:"id"`
        Name      string `yaml:"name"`
        Age       int
        BirthDate struct {
            Year  int `json:"year"`
            Month int `yaml:"month"`
            Day   int `yaml:"day"`
        }
        Ignored string `yaml:"-"`
        Map2 map[string]interface{} `yaml:",inline"`
    }
    v := reflect.ValueOf(User{})
    info, err := Get(v.Type(), "")
    if err != nil {
        t.Fatal(err)
    }
    if info.InlineMapIndex != 5 {
        t.Fatalf("inlined map number is %d, expected 5", info.InlineMapIndex)
    }
    for name, field := range info.FieldsMap {
        t.Logf("%s: %#v", name, field)
    }
}
