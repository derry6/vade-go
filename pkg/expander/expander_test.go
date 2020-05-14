package expander_test

import (
    "reflect"
    "testing"

    "github.com/derry6/vade-go/pkg/expander"
)

type testStore map[string]interface{}

func (s testStore) Set(k string, v interface{}) {
    s[k] = v
}

func (s testStore) Get(key string)(value interface{}, ok bool) {
    value, ok = s[key]
    return
}

func TestExpansion(t *testing.T) {
    var testCases =  []struct {
        Store testStore
        key   string
        value interface{}
    }{
        {testStore{"k": "100", "k2": "${k}"}, "k2", "100"},
        {testStore{"k": "100", "k2": "aa${k}"}, "k2", "aa100"},
        {testStore{"k": "100", "k2": "${k}bb"}, "k2", "100bb"},
        {testStore{"k": "100", "k2": "aa${k}bb"}, "k2", "aa100bb"},
        {testStore{"k": "100", "k2": "aa${k}bb${k}cc"}, "k2", "aa100bb100cc"},
        {testStore{"k": "100", "k2": "aa${k}bb", "k3": "${k2}cc"}, "k3", "aa100bbcc"},
        {testStore{"k": "100", "k2": "aa${k}bb", "k3": "${k}and${k2}"}, "k3", "100andaa100bb"},
    }
    for _, c := range testCases {
        v, err := expander.New(c.Store.Get).Expand(c.key)
        if err != nil {
            t.Fatal(err)
        }
        if !reflect.DeepEqual(v, c.value) {
            t.Fatalf("value is %v, want %v", v, c.value)
        }
    }
}

var customExpandHandler = func(in string)(out interface{}, err error) {
    return "__" + in + "__", nil
}

func TestExpansionCustomPattern(t *testing.T) {
    var testCases =  []struct {
        Store testStore
        key   string
        value interface{}
    }{
        {testStore{"k": "100", "k2": "$C{k}"}, "k2", "__k__"},
        {testStore{"k": "100", "k2": "aa$C{k}"}, "k2", "aa__k__"},
        {testStore{"k": "100", "k2": "$C{k}bb"}, "k2", "__k__bb"},
        {testStore{"k": "100", "k2": "aa$C{k}bb"}, "k2", "aa__k__bb"},
        {testStore{"k": "100", "k2": "aa$C{k}bb$C{k}cc"}, "k2", "aa__k__bb__k__cc"},
        {testStore{"k": "100", "k2": "aa$C{k}bb", "k3": "$C{k2}cc"}, "k3", "__k2__cc"},
        {testStore{"k": "100", "k2": "aa$C{k}bb", "k3": "$C{k}and$C{k2}"}, "k3", "__k__and__k2__"},
    }
    for _, c := range testCases {
        v, err := expander.New(c.Store.Get, expander.WithExpansion("$C{", "}", customExpandHandler)).Expand(c.key)
        if err != nil {
            t.Fatal(err)
        }
        if !reflect.DeepEqual(v, c.value) {
            t.Fatalf("value is %v, want %v", v, c.value)
        }
    }
}
