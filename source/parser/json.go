package parser

import (
    "encoding/json"

    "github.com/derry6/vade-go/pkg/flatter"
)

type jsonParser struct {
}

func (p *jsonParser) Parse(data []byte, prefix string) (values map[string]interface{}, err error) {
    raw := make(map[string]interface{})
    if err = json.Unmarshal(data, &raw); err != nil {
        return
    }
    values = make(map[string]interface{})
    err = flatter.Flatten(raw, values, prefix, false)
    return
}

func NewJSON() Parser { return &jsonParser{} }