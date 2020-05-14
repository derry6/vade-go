package structinfo

import "strings"

func toSnakeString(s string) string {
    data := make([]byte, 0, len(s)*2)
    j := false
    num := len(s)
    for i := 0; i < num; i++ {
        d := s[i]
        if i > 0 && d >= 'A' && d <= 'Z' && j {
            data = append(data, '_')
        }
        if d != '_' {
            j = true
        }
        data = append(data, d)
    }
    return strings.ToLower(string(data[:]))
}

func toCamelString(s string) string {
    data := make([]byte, 0, len(s))
    bs := []byte(s)
    abbrLen := 0
    isNoChar := false

    for i, d := range bs {
        switch {
        case d >= 'A' && d <= 'Z':
            if abbrLen == 0 && i != 0 {
                data = append(data, d)
            } else {
                data = append(data, d + 32)
            }
            abbrLen += 1
            isNoChar = false
        case d >= 'a' && d <= 'z':
            if abbrLen > 1 {
                n := len(data)
                data[n-1] =  data[n-1] - 32
            }

            if isNoChar && i != 0 {
                data = append(data, d - 32)
            } else {
                data = append(data, d)
            }

            abbrLen = 0
            isNoChar = false
        case d >= '0' && d <= '9':
            data = append(data, d)
            abbrLen = 0
            isNoChar = true
        default:
            abbrLen = 0
            isNoChar = true
        }
    }
    return string(data[:])
}
