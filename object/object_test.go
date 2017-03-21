package object

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHashKeys(t *testing.T) {
	Convey("hash keys", t, func() {
		hello1 := NewString("Hello world")
		hello2 := NewString("Hello world")
		diff := NewString("My name is johnny")

		So(hello1.HashKey() == hello2.HashKey(), ShouldBeTrue)
		So(hello1.HashKey() == diff.HashKey(), ShouldBeFalse)
	})
}
