package vade

import (
    "testing"
    "time"
)

type testUnmarshalGetter struct {
    values map[string]interface{}
}

func newTestUnmarshalGetter() *testUnmarshalGetter {
    return &testUnmarshalGetter{
        values: map[string]interface{}{},
    }
}
func (t *testUnmarshalGetter) Set(key string, value interface{}) {
    t.values[key] = value
}
func (t *testUnmarshalGetter) Get(key string) (value interface{}, ok bool) {
    value, ok = t.values[key]
    return
}
func (t *testUnmarshalGetter) Keys() (keys []string) {
    for k := range t.values {
        keys = append(keys, k)
    }
    return keys
}

func TestUnmarshalBasic(t *testing.T) {
    store := newTestUnmarshalGetter()
    store.Set("i", 100)
    store.Set("f", 2.0)
    store.Set("s", "debug")
    store.Set("d", "3s")
    type Value struct {
        I int           `yaml:"i"`
        F float64       `yaml:"f"`
        S string        `yaml:"s"`
        D time.Duration `yaml:"d"`
    }
    var v Value
    if err := unmarshal(store.Get, store.Keys(), &v); err != nil {
        t.Error(err)
    }
    if v.D != 3*time.Second {
        t.Errorf("Unmarshal error: c.d=%v, expect 3s", v.D)
    }
}
