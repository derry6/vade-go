package nacos_test

import (
    "context"
    "testing"
    "time"

    "github.com/derry6/vade-go/source/client"
    "github.com/derry6/vade-go/source/client/nacos"
)

func newNacosClient(t *testing.T, addr string) client.Client {
    cfg := client.DefaultConfig()
    cfg.Address = "localhost:8848"
    cfg.Timeout = time.Second
    cli, err := client.New(nacos.Name, cfg)
    if err != nil{
        t.Fatalf("Can not create nacos client: %v", err)
    }
    return cli
}

func watchNacosPath(t *testing.T, path string, rsp chan []byte) error {
    cli := newNacosClient(t, "localhost:8848")
    go func() {
        defer cli.Close()
        if err := cli.Watch(path, func(data []byte) {
            rsp <- data
        }); err != nil {
            t.Fatalf("Can not watch nacos path: %v", err)
        }
    }()
    return nil
}

func TestNacosClientPushPull(t *testing.T) {
    cli := newNacosClient(t, "localhost:8848")
    defer cli.Close()
    testData := []byte("a: 100")
    err := cli.Push(context.TODO(), "test.yaml", testData)
    if err != nil {
        t.Fatalf("Can not push content to nacos: %v", err)
    }
    data, err := cli.Pull(context.TODO(), "test.yaml")
    if err != nil {
        t.Fatalf("Can not pull content from nacos: %v", err)
    }
    if string(data) != string(testData) {
        t.Fatalf("nacos data is %q: want %q", string(data), string(testData))
    }
}

func TestNacosClientWatch(t *testing.T) {
    var (
        testData = "test_b: 100"
        testPath = "test-watch.yaml"
    )
    cli := newNacosClient(t, "localhost:8848")
    defer cli.Close()
    err := cli.Push(context.TODO(), testPath, []byte("test_a: 10"))
    if err != nil {
        t.Fatalf("Can not push content to nacos: %v", err)
    }
    rsp := make(chan []byte)
    defer close(rsp)

    _ = watchNacosPath(t, testPath, rsp)
    err =cli.Push(context.TODO(), testPath, []byte(testData))
    if err != nil {
        t.Fatalf("Can not push content to nacos: %v", err)
    }
    select {
    case <-time.After(100*time.Millisecond):
        t.Fatalf("Watch nacos path timeout")
        case data := <-rsp:
            if string(data) != testData {
                t.Fatalf("Nacos path changed: %q, want %q", string(data), testData)
            }
    }
}