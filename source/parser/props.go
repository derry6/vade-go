package parser

import (
    "bufio"
    "bytes"
    "io"
    "strings"
)

type propsParser struct {
}

func (p *propsParser) Parse(data []byte, prefix string) (values map[string]interface{}, err error) {
    var line string
    values = make(map[string]interface{})
    buf := bufio.NewReader(bytes.NewReader(data))
    if n := len(prefix); n > 0 && prefix[n-1] != '.' {
        prefix += "."
    }
    for {
        if err == io.EOF {
            break
        }
        line, err = buf.ReadString('\n')
        if err != nil && err != io.EOF {
            return nil, err
        }
        parts := strings.Split(line, "=")
        if len(parts) != 2 {
            continue
        }
        key := strings.TrimSpace(parts[0])
        if len(key) == 0 || key[0] == '#' {
            continue
        }
        key = prefix + key
        i := strings.Index(parts[1], "#")
        if i > 0 {
            parts[1] = parts[1][:i]
        }
        values[key] = strings.TrimSpace(parts[1])
    }
    return values, nil
}

func NewProps() Parser {
    return &propsParser{}
}