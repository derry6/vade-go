package structinfo

import "strings"

var (
    SupportedTags = []string{"prop", "yaml", "json"}
)

// TagOptions is the string following a comma in a struct field's "json"
// tag, or the empty string. It does not include the leading comma.
type TagOptions string

func (o TagOptions) Contains(optionName string) bool {
    if len(o) == 0 {
        return false
    }
    s := string(o)
    for s != "" {
        var next string
        i := strings.Index(s, ",")
        if i >= 0 {
            s, next = s[:i], s[i+1:]
        }
        if s == optionName {
            return true
        }
        s = next
    }
    return false
}
