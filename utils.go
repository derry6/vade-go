package vade

import (
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"

    "github.com/go-errors/errors"

    "github.com/derry6/vade-go/pkg/log"
)

const (
    maxFileSize       = 1 * 1024 * 1024
    nanoSecondsPerDay = 24 /*h*/ * 60 /*m*/ * 60 /*s*/ * 1e9
)

func filesInDir(root string, cbFn func(f string) error) (err error) {
    return filepath.Walk(root, func(path string, info os.FileInfo, iErr error) error {
        log.Get().Infof("walk: %s", path)
        if iErr != nil {
            return iErr
        }
        if info.IsDir() {
            if path == root {
                return nil
            }
            return filepath.SkipDir
        }
        if info.Size() < maxFileSize {
            if iErr = cbFn(path); iErr != nil {
                return iErr
            }
        }
        return nil
    })
}

// toDuration 支持带d的duration解析
func toDuration(s string) (v time.Duration, err error) {
    var d int64
    di := strings.Index(s, "d")
    if di < 0 {
        return time.ParseDuration(s)
    }
    v, err = time.ParseDuration(s[di+1:])
    if err != nil {
        return 0, err
    }
    if di == 0 {
        d = 1
    } else {
        ds := s[:di]
        d, err = strconv.ParseInt(ds, 10, 64)
        if err != nil {
            return 0, errors.Errorf("invalid day: %v", err)
        }
        if d < 0 {
            return 0, errors.Errorf("invalid day: %v", d)
        }
    }
    return time.Duration(int64(v) + d*nanoSecondsPerDay), nil
}
