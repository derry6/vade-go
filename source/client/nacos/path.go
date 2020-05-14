package nacos

import (
    "fmt"
    "net/url"
    "path"
)

type nPath struct {
    namespace string
    group     string
    dataId    string
}

func (p *nPath) String() string {
    return fmt.Sprintf("%s@%s@%s", p.dataId, p.group, p.namespace)
}

func (p *nPath) Ext() string {
    ext := path.Ext(p.dataId)
    if len(ext) != 0 {
        return ext[1:]
    }
    return "properties"
}

func newPath(path string, defaultGroup, defaultNamespace string) *nPath {
    u, err := url.Parse(path)
    if err != nil {
        return nil
    }
    q := u.Query()
    p := &nPath{dataId: u.Path, group: defaultGroup, namespace: defaultNamespace}
    if group := q.Get("group"); group != "" {
        p.group = group
    }
    if ns := q.Get("namespace"); ns != "" {
        p.namespace = ns
    }
    return p
}
