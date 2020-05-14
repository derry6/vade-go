# vade-go
分布式应用配置管理框架，从多种配置源中拉取配置转化为`properties`的格式进行管理，监听配置变更进行实时同步。

## 特性
1. 多种配置源, 容易扩展。
2. 自定义优先级。
3. 多种配置格式,　默认支持json,yaml,properties格式。
4. 支持变量替换，默认支持${}形式。
5. 支持将配置Unmarshal到结构体或者map中。
6. 支持配置覆盖或设置默认配置。
7. 可Watch配置变更。

## 安装
```shell script
go get -v github.com/derry6/vade-go
```

## 使用

#### 1. 基本使用
```go

func init() {
    // 定义命令行参数
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
    // 初始化
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
```

#### 2. 远程配置源
```go
    // 初始化
    vade.Init()
    
    // 创建配置源客户端
    cfg := client.DefaultConfig()
    cfg.Address = "localhost:8848"
    cfg.Timeout = 3 * time.Second
    cli, _ := client.New("nacos", cfg)
    s := source.New("nacos", cli)
    // 添加配置
    _ = s.AddPath("test.yaml", source.WithPathRequired())
    
    _ = vade.AddSource(s)
```

#### 3. 自定义变量替换
```go
func doExpand(in string)(v interface{}, err error) {
    return "replaced-" + in ,nil
}

vade.Init(vade.WithExpansion("${", "}",doExpand))

```

#### 4. Unmarshal
1. 支持数据类型bool/int/float/string/map/struct/slice, 支持内嵌struct
2. 支持指定tag, 默认使用yaml.

```go
    type Value struct {
        Int    int                    `yaml:"int"`
        Float  float64                `yaml:"float"`
        Bool   bool                   `yaml:"bool"`
        String string                 `yaml:"string"`
        Map    map[string]interface{} `yaml:"map"`
        Slice  []string               `yaml:"slice"`
        Other  map[string]interface{} `yaml:",inline"`
    }
    // properties:
    /*
        abc.config.int
        abc.config.float
        abc.config.bool
        abc.config.string
        abc.config.map.a
        abc.config.map.b
        abc.config.slice[0] 
        abc.config.slice[1]
        abc.config.slice[2]
        abc.config.other1
        abc.config.other2
        abc.config.otherX
    */

    var v Value
    // Unmarshal value.* to v
    vade.Unmarshal(&v,vade.WithUnmarshalPrefix("abc.config"), vade.WithUnmarshalTag("yaml"))

```

## 参考
1. [https://github.com/spf13/viper](https://github.com/spf13/viper)
2. [https://github.com/magiconair/properties](https://github.com/magiconair/properties)
