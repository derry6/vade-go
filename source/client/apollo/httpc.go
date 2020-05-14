package apollo

import (
    "context"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"
    "time"

    pkgerrs "github.com/pkg/errors"
)

type httpResult struct {
    Code int
    Body []byte
}

type httpClient struct {
    client *http.Client
}

func newHttpClient(_ time.Duration, _ *tls.Config) *httpClient {
    return &httpClient{client: &http.Client{}}
}

func (c *httpClient) NewRequest(method string, urlFmt string, values url.Values, args ...interface{}) (req *http.Request, err error) {
    URL := fmt.Sprintf(urlFmt, args...)
    if len(values) > 0 {
        URL += "?" + values.Encode()
    }
    req, err = http.NewRequest(method, URL, nil)
    if err != nil {
        return req, err
    }
    return req, err
}

func (c *httpClient) readAll(rsp *http.Response) (res *httpResult, err error) {
    res = &httpResult{Code: rsp.StatusCode}
    defer func() { _ = rsp.Body.Close() }()
    if res.Code > 404 {
        return res, nil
    }
    res.Body, err = ioutil.ReadAll(rsp.Body)
    return
}

func (c *httpClient) json(rsp *http.Response, js interface{}) (res *httpResult, err error) {
    if res, err = c.readAll(rsp); err != nil {
        return
    }
    if len(res.Body) > 0 {
        if err = json.Unmarshal(res.Body, js); err != nil {
            return res, pkgerrs.Errorf("unmarshal '%s': %v", string(res.Body), err)
        }
    }
    return
}

func (c *httpClient) Raw(ctx context.Context, req *http.Request) (res *httpResult, err error) {
    res = &httpResult{}
    rsp, err := c.client.Do(req.WithContext(ctx))
    if err != nil {
        return nil, err
    }
    return c.readAll(rsp)
}

func (c *httpClient) Get(ctx context.Context, urlFmt string, values url.Values, js interface{}, args ...interface{}) (res *httpResult, err error) {
    req, err := c.NewRequest(http.MethodGet, urlFmt, values, args...)
    if err != nil {
        return nil, err
    }
    rsp, err := c.client.Do(req.WithContext(ctx))
    if err != nil {
        return nil, err
    }
    return c.json(rsp, js)
}
