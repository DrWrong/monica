package core

import (
	"testing"
)

func TestRouter(t *testing.T) {
	fn := func() {
		println("i am the handler")
	}
	Group("^/product",
		func() {
			Handle(`^/update/(?P<id>\d+)$`, fn)
			Handle(`^/create$`, fn)
			Group("^/partial", func(){
				Handle("^/test$", fn)
			})
		},
	)
	DebugRoute()
	processor, err := GetProcessor("/product/update/1")
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v", processor)

	processor, err = GetProcessor("/product/partial/test")

	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", processor)
}
