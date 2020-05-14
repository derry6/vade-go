package source

import (
    "context"
    "errors"
    "sync"
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/derry6/vade-go/source/client"
)

type fakeClient struct {
    data map[string]string
    mu   sync.RWMutex
}

func (c *fakeClient) Close() error { return nil }
func (c *fakeClient) Name() string { return "fake" }
func (c *fakeClient) Pull(ctx context.Context, path string) (data []byte, err error) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    if d, ok := c.data[path]; ok {
        return []byte(d), nil
    }
    return nil, errors.New("not exists")
}
func (c *fakeClient) Push(ctx context.Context, path string, data []byte) error {
    c.mu.Lock()
    c.data[path] = string(data)
    c.mu.Unlock()
    return nil
}
func (c *fakeClient) Watch(path string, cb client.ChangedCallback) error { return nil }

func newFakeClient(opts *client.Config) (c client.Client, err error) {
    return &fakeClient{
        data: map[string]string{},
    }, nil
}

func newFakeSource(t *testing.T, opts ...Option) Source {
    _ = client.RegisterClient("fake", newFakeClient)
    c, _ := client.New("fake", nil)
    return New("fake", c, opts...)
}

func TestSourceNew(t *testing.T) {
    _ = newFakeSource(t).Close()
}

func TestSourceAddPath(t *testing.T) {
    var requires = []struct {
        name string
        data []byte
        err  bool
        msg  string
    }{
        {"r0", []byte(`a: from-r0`), false, "ok"},
        {"r1", []byte(`a = from-r1 # comment`), false, "ok"},
        {"r2", []byte(`invalid data`), false, "invalid data"},
        {"r3", []byte(``), false, "empty data"},
    }
    var optionals = []struct {
        name string
        data []byte
        err  bool
        msg  string
    }{
        {"p0", []byte(`a: from-p0`), false, "ok"},
        {"p1", []byte(`a = from-p1 # comment`), false, "ok"},
        {"p2", []byte(`invalid data`), false, "invalid data"},
        {"p3", []byte(``), false, "empty data"},
    }
    s := newFakeSource(t)
    for _, p := range requires {
        _ = s.Client().Push(context.TODO(), p.name, p.data)
        err := s.AddPath(p.name, WithPathRequired())
        assert.Equal(t, err != nil, p.err, p.msg)
    }
    // not exits
    err := s.AddPath("r-not-exists", WithPathRequired())
    assert.Equal(t, err != nil, true, "should be error")

    for _, p := range optionals {
        _ = s.Client().Push(context.TODO(), p.name, p.data)
        err := s.AddPath(p.name)
        assert.Equal(t, err != nil, p.err, p.msg)
    }
    err = s.AddPath("p-not-exists")
    assert.Equal(t, err != nil, false)
}

func TestSourceAddPathPriority(t *testing.T) {
    type V = map[string]interface{}
    var paths = []struct {
        name   string
        data   string
        pri    int
        values V
    }{
        {"p0", `a: p0-a`, 0, V{"a": "p0-a"}},
        {"p3", `a: p3-a`, 3, V{"a": "p3-a"}},
    }
    s := newFakeSource(t)

    for _, p := range paths {
        _ = s.Client().Push(context.TODO(), p.name, []byte(p.data))
        err := s.AddPath(p.name, WithPathPriority(p.pri))
        assert.NoError(t, err, "add path error")
        v, ok := s.Get("a")
        assert.True(t, ok)
        assert.Equal(t, p.values["a"], v)
    }
}
