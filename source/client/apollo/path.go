package apollo

import (
    "net/url"
    "path"
    "strings"
)

type aPath struct {
    namespace string // namespace
    appId     string
    cluster   string //
}

func (p *aPath) Ext() string {
    ext := path.Ext(p.namespace)
    if len(ext) > 0 {
        return ext[1:]
    }
    return "properties"
}

func (p *aPath) String() string {
    b := strings.Builder{}
    b.WriteString(p.namespace)
    b.WriteString("?")
    b.WriteString("appId=")
    b.WriteString(p.appId)
    b.WriteString("&cluster=")
    b.WriteString(p.cluster)
    return b.String()
}

func (p *aPath) watchID() string {
    return p.cluster + "@" + p.appId
}

func newPath(path, defaultCluster, defaultApp string) *aPath {
    u, err := url.Parse(path)
    if err != nil {
        return nil
    }
    q := u.Query()
    p := &aPath{namespace: u.Path, appId: q.Get("appId"), cluster: q.Get("cluster")}
    if p.cluster == "" {
        p.cluster = defaultCluster
    }
    if p.appId == "" {
        p.appId = defaultApp
    }
    return p
}
