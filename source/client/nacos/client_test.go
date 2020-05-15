package nacos_test

import (
    "context"
    "fmt"
    "math/rand"
    "os"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"

    "github.com/derry6/vade-go/source/client"
    "github.com/derry6/vade-go/source/client/nacos"
)

var (
    testServerAddress = "localhost:8848"
    testLogDir        = "/tmp/nacos_client_test_log"
    testTimeout       = time.Second
    testPullPath      = "test-pull.yaml"
    testWatchPath     = "test-watch.yaml"
)

func init() {
    rand.Seed(time.Now().UnixNano())
}

func cleanUp(c client.Client) {
    _ = os.RemoveAll("/tmp/nacos_client_test_log")
    _ = c.Close()
}

func newTestClient() (client.Client, error) {
    cfg := client.DefaultConfig()
    cfg.Address = testServerAddress
    cfg.Timeout = testTimeout
    cfg.LogDir = testLogDir
    return client.New(nacos.Name, cfg)
}

func TestClient_Pull(t *testing.T) {
    cli, err := newTestClient()
    assert.NoError(t, err)
    defer cleanUp(cli)
    testData := uuid.New().String()
    err = cli.Push(context.TODO(), testPullPath, []byte(testData))
    assert.NoError(t, err)
    time.Sleep(500 * time.Millisecond)
    getData, err := cli.Pull(context.TODO(), testPullPath)
    assert.NoError(t, err)
    assert.Equal(t, testData, string(getData))
}

func TestClient_Watch(t *testing.T) {
    watchData := fmt.Sprintf("test_watch: %s", uuid.New().String())
    cli, err := newTestClient()
    assert.NoError(t, err)
    defer cleanUp(cli)
    events := make(chan []byte, 1)
    // create dataId
    err = cli.Push(context.TODO(), testWatchPath, []byte(watchData))
    assert.NoError(t, err)

    time.Sleep(time.Second)
    // watch dataId
    err = cli.Watch(testWatchPath, func(data []byte) {
        // fmt.Printf("Nacos config %q changed: %s\n", testWatchPath, data)
        events <- data
    })
    assert.NoError(t, err)
    select {
    case <-time.After(5 * time.Second):
        assert.FailNow(t, "Watch timeout")
    case data := <-events:
        assert.Equal(t, watchData, string(data))
    }
}