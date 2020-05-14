package apollo

import (
    "encoding/json"

    "github.com/go-errors/errors"
    pkgerrs "github.com/pkg/errors"
)

type response struct {
    AppID      string          `json:"appId"`
    Cluster    string          `json:"cluster"`
    Namespace  string          `json:"namespaceName"`
    Configs    json.RawMessage `json:"configurations"`
    ReleaseKey string          `json:"releaseKey"`
}

func (c *Client) parseRsp(p *aPath, raw []byte) (releaseKey string, data []byte, err error) {
    var rsp response
    if err = json.Unmarshal(raw, &rsp); err != nil {
        return "", nil, err
    }
    releaseKey = rsp.ReleaseKey
    switch p.Ext() {
    case "yaml", "yml", "json":
        raw = rsp.Configs
    default:
        return releaseKey, rsp.Configs, nil
    }
    cMap := map[string]interface{}{}
    if err = json.Unmarshal(raw, &cMap); err != nil {
        return releaseKey, nil, err
    }
    txt, ok := cMap["content"]
    if !ok {
        return releaseKey, nil, pkgerrs.Errorf("unknown response: %s", raw)
    }
    s, ok2 := txt.(string)
    if !ok2 {
        return releaseKey, nil, pkgerrs.Errorf("unknown response: %s", raw)
    }
    return releaseKey, []byte(s), nil
}

func (c *Client) parseCachedRsp(p *aPath, raw []byte) (data []byte, err error) {
    switch p.Ext() {
    case "yaml", "yml", "json":
    default:
        return raw, err
    }
    cMap := map[string]interface{}{}
    if err = json.Unmarshal(raw, &cMap); err != nil {
        return nil, err
    }
    txt, ok := cMap["content"]
    if !ok {
        return nil, errors.Errorf("unknown response: %s", raw)
    }
    s, ok2 := txt.(string)
    if !ok2 {
        return nil, errors.Errorf("unknown response: %s", raw)
    }
    return []byte(s), err
}
