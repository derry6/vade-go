package main

import (
    "time"

    "github.com/derry6/vade-go"
    "log"
)

func init() {
    vade.Flag("vade.logging.level", "xx", "logging level")
    vade.Flag("a.b.c", 100, "abc")
}

type listener struct {

}
func (e *listener) OnPropertyEvent(ev *vade.Event) {
    log.Printf("Property changed: %v", ev.String())
}

func main() {
    requires := []string{"service.yaml"}
    optionals :=[]string{"logger.yaml"}
    err := vade.Init(
        vade.WithFileSource(requires, optionals),
        vade.WithEnvSource(),
        vade.WithFlagSource())
    if err != nil {
        panic(err)
    }
    // 覆盖所有的配置
    vade.Set("time.now", time.Now().Format(time.RFC3339))

    values := vade.All()
    for k, v := range values {
       log.Printf("%s = %v", k, v)
    }
    log.Printf("time.now = %v", vade.MustTime("time.now").String())

    // Watch 配置的变更
    id := vade.Watch("^vade.logging.*$", &listener{})
    defer vade.Unwatch(id)
    select{}
}
