package main

import (
	"log"
	"time"

	"github.com/derry6/vade-go"
	"github.com/derry6/vade-go/source"
	"github.com/derry6/vade-go/source/client"
	_ "github.com/derry6/vade-go/source/client/nacos"
)

type Value struct {
	Int    int                    `yaml:"int"`
	Float  float64                `yaml:"float"`
	Bool   bool                   `yaml:"bool"`
	String string                 `yaml:"string"`
	Map    map[string]interface{} `yaml:"map"`
	Slice  []string               `yaml:"slice"`
	Other  map[string]interface{} `yaml:",inline"`
}

// func doExpand(in string) (v interface{}, err error) {
// 	return "replaced-" + in, nil
// }

func main() {
	err := vade.Init()
	if err != nil {
		log.Fatalf("Can not init vade: %v", err)
	}
	cfg := client.DefaultConfig()
	cfg.Address = "localhost:8848"
	cfg.Timeout = 3 * time.Second

	sourceName := "nacos"

	cli, _ := client.New(sourceName, cfg)
	s := source.New(sourceName, cli)
	_ = vade.AddSource(s)
	_ = vade.AddPath(sourceName, "test.yaml", source.WithPathRequired())

	var v Value
	opts := []vade.UnmarshalOption{
		vade.WithUnmarshalPrefix("value"),
		vade.WithUnmarshalTag("yaml"),
	}
	_ = vade.Unmarshal(&v, opts...)
}
