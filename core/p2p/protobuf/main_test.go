package protobuf_test

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(
		m,
		goleak.IgnoreTopFunction("time.Sleep"),
		goleak.IgnoreTopFunction("io.(*pipe).read"),
		goleak.IgnoreTopFunction("io.(*pipe).write"),
		goleak.IgnoreTopFunction("github.com/redesblock/mop/core/p2p/protobuf_test.newMessageWriter.func1"),
	)
}
