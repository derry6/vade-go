package parser_test

import (
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/derry6/vade-go/source/parser"
)

func TestParseData(t *testing.T) {
    type v = map[string]interface{}

    var testCases = []struct {
        data   string
        values v
    }{
        {`invalid data`, v{}},
        {`a: 10 # from yaml`, v{"a": 10}},
        {`b = 100 # from properties`, v{"b": "100"}},
        {`{"c": "valuec"}`, v{"c": "valuec"}},
        {``, v{}},
    }
    for _, testCase := range testCases {
        x, err := parser.NewDefault().Parse([]byte(testCase.data), "")
        assert.NoError(t, err, "parse data error")
        assert.Equal(t, x, testCase.values, "values not match")
    }
}
