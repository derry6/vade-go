package consul

import (
    "context"
    "fmt"
    "testing"
    "time"

    "github.com/derry6/vade-go/source/client"
)

var (
    testPath = "testApp/testCluster/services/testSvc/application.yaml"
)

const (
    testSvc      = "testSvc"
    testFileName = "application.yaml"
    testData     = `
app:
  name: testApp
  service: testSvc
  logging:
    level: debug
`
)

func buildAbsolutePath(serviceName, fileName string) string {
    return fmt.Sprintf("vade/apps/testApp/services/%s/%s",
        serviceName, fileName)
}

func createTestClient() (client.Client, error) {
    return NewClient(&client.Config{
        Address:    "172.22.0.4:8500",
        DataCenter: "consul-shenzhen",
    })
}

func TestClientPushPull(t *testing.T) {
    c, err := createTestClient()
    if err != nil {
        t.Error(err)
    }
    defer c.Close()
    testPath := buildAbsolutePath(testSvc, testFileName)
    err = c.Push(context.Background(), testPath, []byte(testData))
    if err != nil {
        t.Error(err)
    }
    data, err := c.Pull(context.Background(), testPath)
    if err != nil {
        t.Error(err)
    }
    if string(data) != testData {
        t.Errorf("content error: data=%q, expect=%q", string(data), testData)
    }
}

func TestClientWatch(t *testing.T) {
    c, err := createTestClient()
    if err != nil {
        t.Error(err)
    }
    defer c.Close()

    ch := make(chan []byte)
    testPath := buildAbsolutePath(testSvc, testFileName)
    err = c.Watch(testPath, func(data []byte) {
        ch <- data
    })
    if err != nil {
        t.Error(err)
    }
    err = c.Push(context.Background(), testPath, []byte(testData))
    if err != nil {
        t.Error(err)
    }
    select {
    case data := <-ch:
        if string(data) != testData {
            t.Errorf("content error: data=%q, expect=%q", string(data), testData)
        }
    case <-time.After(100 * time.Millisecond):
        t.Errorf("watch timeout")
    }
}
