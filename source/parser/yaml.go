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
	var (
		raw     map[string]interface{}
		decoder = yaml.NewDecoder(bytes.NewReader(data))
	)
	values = make(map[string]interface{})
	for {
		raw = make(map[string]interface{})
		if err = decoder.Decode(raw); err != nil {
			if err == io.EOF {
				return values, nil
			}
			return nil, err
		}
		if err = flatter.Flatten(raw, values, prefix, false); err != nil {
			return nil, err
		}
	}
}

func NewYAML() Parser {
	return &yamlParser{}
}
