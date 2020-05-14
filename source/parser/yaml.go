package parser

import (
    "bytes"
    "io"

    "gopkg.in/yaml.v2"

    "github.com/derry6/vade-go/pkg/flatter"
)

type yamlParser struct {
}

func (p *yamlParser) Parse(data []byte, prefix string) (values map[string]interface{}, err error) {
    raw := make(map[string]interface{})
    d := yaml.NewDecoder(bytes.NewReader(data))

    n := 0
    for {
        var part map[string]interface{}
        if n == 0 {
            part = raw
        } else {
            part = make(map[string]interface{})
        }
        if err = d.Decode(part); err != nil {
            if err != io.EOF {
                return
            }
            err = nil
            break
        } else if n > 0 {
            for k, v := range part {
                raw[k] = v
            }
        }
        n++
    }
    values = make(map[string]interface{})
    err = flatter.Flatten(raw, values, prefix, false)
    return
}

func NewYAML() Parser {
    return &yamlParser{}
}