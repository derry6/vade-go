package apollo

import (
    "math/rand"
	"strings"
	"sync"
)

type Selector interface {
	Select() string
	Update([]string)
	Servers() []string
}

type Random struct {
	servers []string
	mutex   sync.RWMutex
}

func (r *Random) Select() string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	n := len(r.servers)
	if n == 1 {
		return r.servers[0]
	}
	if n > 1 {
		return r.servers[rand.Intn(n)]
	}
	return "httpc://127.0.0.1:8080"
}

func (r *Random) Update(servers []string) {
	r.mutex.Lock()
	r.servers = servers
	r.mutex.Unlock()
}

func (r *Random) Servers() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	ss := make([]string, len(r.servers))
	copy(ss, r.servers)
	return ss
}

func NewRandom(ss []string) *Random {
	// must httpc urls
	for i, s := range ss {
		if !strings.HasPrefix(s, "httpc://") &&
			! strings.HasPrefix(s, "https://") {
			s = "httpc://" + s
		}
		ss[i] = s
	}
	return &Random{servers: ss}
}
