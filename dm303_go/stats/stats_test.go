//stats_test.go
package stats

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestStats(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	SetValue("concurrent", 1)
	for i:=0;i<100;i++ {
		go inc(i)
	}
	time.Sleep(5 * time.Second)
	fmt.Printf("%d\n", GetValue("concurrent"))
}

func inc(j int) {
	for i:=0;i<1000;i++ {
		IncBy("concurrent", 3)
	}
}
