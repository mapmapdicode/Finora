// Package metrics provides a deliberately small dependency-free Prometheus
// exposition surface for MVP operational counters. Values are process-local;
// production aggregation is delegated to the metrics scraper.
package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

var counters sync.Map // map[string]*atomic.Int64

func Inc(name string) { Add(name, 1) }
func Add(name string, value int64) {
	if value <= 0 {
		return
	}
	item, _ := counters.LoadOrStore(name, &atomic.Int64{})
	item.(*atomic.Int64).Add(value)
}

func Handler(w http.ResponseWriter, _ *http.Request) {
	keys := []string{}
	counters.Range(func(key, _ any) bool { keys = append(keys, key.(string)); return true })
	sort.Strings(keys)
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	for _, key := range keys {
		item, _ := counters.Load(key)
		_, _ = fmt.Fprintf(w, "%s %d\n", sanitize(key), item.(*atomic.Int64).Load())
	}
}

func sanitize(name string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, name)
}
