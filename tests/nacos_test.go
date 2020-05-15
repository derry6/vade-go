package tests

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"

    "github.com/derry6/vade-go"
    "github.com/derry6/vade-go/source/client"
    "github.com/derry6/vade-go/source/client/nacos"
)

var (
    testNacosAddress = "localhost:8848"
    testNacosTimeout = time.Second
)

func initNacos(t *testing.T) {
    cfg := client.DefaultConfig()
    cfg.Address = testNacosAddress
    cfg.Timeout = testNacosTimeout
    err := vade.Init(vade.WithRemoteSource(nacos.Name, cfg))
    assert.NoError(t, err)
}

func TestNacosSource(t *testing.T) {

}
