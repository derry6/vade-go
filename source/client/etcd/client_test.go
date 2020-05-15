package etcd

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/derry6/vade-go/source/client"
)

const (
    testHost = "localhost:2379"
    testPath = "/abc.yaml"
)

func newTestClient(t *testing.T) client.Client {
    cfg := client.DefaultConfig()
    cfg.Address = testHost
    c, err := NewClient(cfg)
    assert.NoError(t, err)
    return c
}

func TestClient_Pull(t *testing.T) {
    c := newTestClient(t)
    defer c.Close()
    err := c.Push(context.TODO(), testPath, []byte("abc"))
    assert.NoError(t, err)
    data, err := c.Pull(context.TODO(), testPath)
    assert.NoError(t, err)
    assert.Equal(t, "abc", string(data))
}
