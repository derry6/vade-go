package apollo

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "time"

    "github.com/derry6/vade-go/pkg/log"
)

// watch
func (c *Client) listenParamString(p *configPath, namespaces []string) string {
    type ListenParam struct {
        Namespace string `json:"namespaceName,omitempty"`
        ID        int64  `json:"notificationId,omitempty"`
    }
    var v []*ListenParam
    c.mutex.RLock()
    for _, namespace := range namespaces {
        p.namespace = namespace
        id, _ := c.nIDs[p.fullKey()]
        v = append(v, &ListenParam{Namespace: namespace, ID: id})
    }
    c.mutex.RUnlock()
    data, _ := json.Marshal(v)
    return string(data)
}

func (c *Client) listen(p *configPath) {
    type notification struct {
        Name string `json:"namespaceName,omitempty"`
        ID   int64  `json:"notificationId,omitempty"`
    }
    timeout := c.timeout * 10
    watchKey := p.watchKey()

    log.Get().Debugf("Listen %q, timeout = %s", p.fullKey(), timeout)
    for {
        var (
            changes []notification
            namespaces []string
        )
        c.mutex.RLock()
        namespaces, _ = c.watches[watchKey]
        c.mutex.RUnlock()

        if len(namespaces) == 0 {
            time.Sleep(timeout)
            continue
        }
        param := c.listenParamString(p, namespaces)
        ctx, cancel := context.WithTimeout(context.Background(), timeout)
        values := url.Values{}
        values.Set("appId", p.appId)
        values.Set("cluster", p.cluster)
        values.Set("notifications", param)
        reqURL := fmt.Sprintf(watchURLFmt, c.selector.Select())

        req, err := c.cli.NewRequest(http.MethodGet, reqURL, values)
        if err != nil {
            log.Get().Errorf("Failed to create listen request: %v", err)
            time.Sleep(timeout)
            continue
        }
        sign(req, p.appId, p.token)

        result, err := c.cli.Get(ctx, req, &changes)
        if err != nil {
            if ctx.Err() != context.DeadlineExceeded {
                log.Get().Debugf("Listen %q error: %v", p.fullKey(), err)
                time.Sleep(timeout)
            } else {
                log.Get().Debugf("Listen %q: timeout after %v", p.fullKey(), timeout)
            }
            cancel()
            continue
        }
        cancel()
        switch result.Code {
        case http.StatusNotModified:
            continue
        case http.StatusOK:
        default:
            time.Sleep(timeout)
            continue
        }
        for _, n := range changes {
            p.namespace = n.Name
            c.onConfigChanged(p, n.ID)
        }
    }
}
