package apollo

import (
    "net/url"
    "path"
)

type configPath struct {
    namespace string // namespace
    appId     string
    cluster   string //
    token     string
}

func (p *configPath) extension() string {
    ext := path.Ext(p.namespace)
    if len(ext) > 0 {
        return ext[1:]
    }
    return "properties"
}

func (p *configPath) fullKey() string {
    return p.namespace + "@" + p.appId + "@" + p.cluster
}

func (p *configPath) watchKey() string {
    return p.cluster + "@" + p.appId
}

func buildConfigPath(path, cluster, appId, token string) *configPath {
    u, err := url.Parse(path)
    if err != nil {
        return nil
    }
    q := u.Query()
    p := &configPath{
        namespace: u.Path,
        appId:     q.Get("appId"),
        cluster:   q.Get("cluster"),
        token:     q.Get("token"),
    }
    if p.cluster == "" {
        p.cluster = cluster
    }
    if p.appId == "" {
        p.appId = appId
    }
    if p.token == "" {
        p.token = token
    }
    return p
}
