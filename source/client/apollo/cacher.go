package apollo

import (
    "fmt"
    "io/ioutil"
    "os"
    "path"
    "path/filepath"
    "sync"

    pkgerrs "github.com/pkg/errors"
)

type Snapshot struct {
    disable bool
    dir     string
    mutex   sync.RWMutex
}

func (c *Snapshot) get(p *aPath) (data []byte, err error) {
    if c.disable {
        return nil, pkgerrs.New("snapshot disabled") // disabled
    }
    fileName := c.file(p)
    c.mutex.RLock()
    data, err = ioutil.ReadFile(fileName)
    c.mutex.RUnlock()
    if err != nil {
        return nil, err
    }
    return data, nil
}

func (c *Snapshot) getReleaseKey(p *aPath) (key string, nId int64, err error) {
    if c.disable {
        return "", 0, pkgerrs.New("snapshot disabled")
    }
    c.mutex.RLock()
    data, err := ioutil.ReadFile(fmt.Sprintf("%s.release", c.file(p)))
    c.mutex.Unlock()
    if err != nil {
        return
    }
    _, err = fmt.Sscanf(string(data), "%s:%d", &key, &nId)
    return
}

func (c *Snapshot) save(p *aPath, release string, nId int64, data []byte) error {
    if c.disable {
        return pkgerrs.New("snapshot disabled")
    }
    fileName := c.file(p)
    dir := path.Dir(fileName)
    _, err := os.Stat(dir)
    if err != nil {
        if os.IsNotExist(err) {
            err = os.MkdirAll(dir, 0755)
        }
    }
    if err != nil {
        return pkgerrs.New("directory not exists")
    }
    c.mutex.Lock()
    err = ioutil.WriteFile(fileName, data, 0644)
    if err == nil {
        d := []byte(fmt.Sprintf("%s:%d", release, nId))
        err = ioutil.WriteFile(fmt.Sprintf("%s.release", fileName), d, 0644)
    }
    c.mutex.Unlock()
    return err
}

func (c *Snapshot) delete(p *aPath) error {
    if c.disable {
        return pkgerrs.New("snapshot disabled")
    }
    c.mutex.Lock()
    err := os.Remove(c.file(p))
    err = os.Remove(fmt.Sprintf("%s.release", c.file(p)))
    c.mutex.Unlock()
    return err
}

func (c Snapshot) clean() (err error) {
    if c.disable {
        return pkgerrs.New("snapshot disabled")
    }
    dir := path.Join(c.dir, "config")
    c.mutex.Lock()
    err = os.RemoveAll(dir)
    c.mutex.Unlock()
    return err
}

func (c *Snapshot) file(p *aPath) string {
    fileName := p.namespace + "@" + p.cluster + "@" + p.appId
    return filepath.Join(c.dir, "apollo", "config", fileName)
}

func newSnapshot(dir string) *Snapshot {
    disable := dir == ""
    return &Snapshot{disable: disable, dir: dir}
}
