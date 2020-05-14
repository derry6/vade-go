package parser

type defaultParser struct {
    yaml Parser
    json Parser
    prop Parser
}

func (p *defaultParser) Parse(data []byte, prefix string) (values map[string]interface{}, err error) {
    if values, err = p.yaml.Parse(data, prefix); err != nil {
        if values, err = p.json.Parse(data, prefix); err != nil {
            return p.prop.Parse(data, prefix)
        }
    }
    return
}

func NewDefault() Parser {
    return &defaultParser{
        yaml: NewYAML(),
        json: NewJSON(),
        prop: NewProps(),
    }
}