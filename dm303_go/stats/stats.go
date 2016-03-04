//stats.go
package stats

var (
    Global StatsBase
)

func init() {
	Global = NewStatsCollector(0)
}

func GetStats() map[string]int64 {
	return Global.getStats()
}

func GetValue(key string) int64 {
	return Global.getValue(key)
}

func IncValue(key string) {
	Global.incValue(key)
}

func SetValue(key string, v int64) {
	Global.setValue(key, v)
}

func ResetKey(key string) {
	Global.resetKey(key)
}

func ClearStats() {
	Global.clearStats()
}

func IncBy(key string, v int64) {
	Global.incBy(key, v)
}
