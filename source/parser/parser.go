package parser

// Parser properties parser
type Parser interface {
    Parse(data []byte, prefix string) (props map[string]interface{}, err error)
}
