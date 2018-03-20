// count map, this shall be safe for concurrent use
package stats

import (
	"sync"
)

const defaultChanSize = 1000

type Unit struct {
	key   string
	value int64
}

type StatsBase interface {
	getStats() map[string]int64
	getValue(string) int64
	setValue(string, int64)
	incValue(string)
	incBy(string, int64)
	resetKey(string)
	clearStats()
}

type StatsCollector struct {
	stats map[string]int64
	c     chan Unit
	lock  sync.Mutex
}

func NewStatsCollector(chanSize int64) *StatsCollector {
	if chanSize <= 0 {
		chanSize = defaultChanSize
	}
	sc := &StatsCollector{
		stats: make(map[string]int64, 0),
		c:     make(chan Unit, chanSize)}
	go sc.inc()
	return sc
}

func (this *StatsCollector) getStats() map[string]int64 {
	this.lock.Lock()
	defer this.lock.Unlock()

	data := make(map[string]int64, len(this.stats))
	for key, value := range this.stats {
		data[key] = value
	}
	return data
}

func (this *StatsCollector) getValue(key string) int64 {
	this.lock.Lock()
	defer this.lock.Unlock()

	if v, ok := this.stats[key]; ok {
		return v
	} else {
		return 0
	}
}

func (this *StatsCollector) setValue(key string, v int64) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.stats[key] = v
}

// use chan to ensure safty of concurrency
func (this *StatsCollector) incValue(key string) {
	this.c <- Unit{key, 1}
}

func (this *StatsCollector) incBy(key string, v int64) {
	this.c <- Unit{key, v}
}

func (this *StatsCollector) resetKey(key string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	if _, ok := this.stats[key]; ok {
		delete(this.stats, key)
	}
}

func (this *StatsCollector) clearStats() {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.stats = make(map[string]int64, 0)
}

// get key from chan and update map
func (this *StatsCollector) inc() {
	// use range to empty the chan
	for unit := range this.c {
		func() {
			this.lock.Lock()
			defer this.lock.Unlock()
			this.stats[unit.key] += unit.value
		}()
	}
}
