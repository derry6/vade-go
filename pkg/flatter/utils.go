package flatter


func mergeKey(prefix, child string) string {
    if n := len(prefix); n == 0 {
        return child
    } else {
        if prefix[n-1] == '.' {
            return prefix + child
        }
        return prefix + "." + child
    }
}
