package apollo

import (
    "context"
    "fmt"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"

    "github.com/derry6/vade-go/source/client"
)

const (
    // apollo demo server
    testHost      = "http://127.0.0.1:8080"
    testAppId     = "SampleApp"
    testCluster   = "default"
    testNamespace = "application"
)

func newTestClient(t *testing.T) client.Client {
    cfg := client.DefaultConfig()
    cfg.Address = testHost
    cfg.Timeout = 3 * time.Second
    cfg.AppID = testAppId
    cfg.Cluster = testCluster
    c, err := NewClient(cfg)
    assert.NoError(t, err)
    return c
}

func TestClient_Pull(t *testing.T) {
    c := newTestClient(t)
    data, err := c.Pull(context.TODO(), testNamespace)
    assert.NoError(t, err)
    assert.True(t, len(data) > 0)
}

func TestClient_Watch(t *testing.T) {
    c := newTestClient(t)
    events := make(chan []byte)
    err := c.Watch(testNamespace, func(data []byte) {
        fmt.Printf("Apollo config changed: %q\n", string(data))
        events <- data
    })
    assert.NoError(t, err)
    select {
    case <-time.After(3 * time.Second):
        t.Fatalf("Watch timeout")
    case data := <-events:
        assert.True(t, len(data) > 0)
    }
}
