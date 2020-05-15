package apollo

import (
    "context"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "net/url"

    pkgerrs "github.com/pkg/errors"
)

type httpResult struct {
    Code int
    Body []byte
}

type httpClient struct {
    client *http.Client
}

func newHttpClient() *httpClient {
    return &httpClient{client: &http.Client{}}
}

func (c *httpClient) NewRequest(method string, reqURL string, values url.Values) (req *http.Request, err error) {
    req, err = http.NewRequest(method, reqURL, nil)
    if err != nil {
        return req, err
    }
    req.URL.RawQuery = values.Encode()
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
    var (
        rsp *http.Response
    )
    res = &httpResult{}
    rsp, err = c.client.Do(req.WithContext(ctx))
    if err != nil {
        return nil, err
    }
    return c.readAll(rsp)
}

func (c *httpClient) Get(ctx context.Context, req *http.Request, obj interface{}) (*httpResult, error) {
    var (
        rsp *http.Response
        err error
    )
    rsp, err = c.client.Do(req.WithContext(ctx))
    if err != nil {
        return nil, err
    }
    return c.json(rsp, obj)
}
