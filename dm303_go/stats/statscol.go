// count map, this shall be safe for concurrent use
package stats

const defaultChanSize = 1000

type Unit struct {
	key     string
	value   int64
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
	stats       map[string]int64
	c			chan Unit
}

func NewStatsCollector(chanSize int64) *StatsCollector {
	if chanSize <= 0 {
		chanSize = defaultChanSize
	}
	sc := &StatsCollector{make(map[string]int64, 0), make(chan Unit, chanSize)}
	go sc.inc()
	return sc
}

func (this *StatsCollector) getStats() map[string]int64 {
	return this.stats
}

func (this *StatsCollector) getValue(key string) int64 {
	if v, ok := this.stats[key]; ok {
		return v
	} else {
		return 0
	}
}

func (this *StatsCollector) setValue(key string, v int64) {
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
	if _, ok := this.stats[key]; ok {
		delete(this.stats, key)
	}
}

func (this *StatsCollector) clearStats() {
	this.stats = make(map[string]int64, 0)
}

// get key from chan and update map
func (this *StatsCollector) inc() {
	// use range to empty the chan
	for unit := range this.c {
		this.stats[unit.key] += unit.value
	}
}
