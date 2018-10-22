package util_test

import (
	"testing"

	"github.com/commis/tm-tools/libs/util"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStringSliceEqual(t *testing.T) {
	Convey("should be return true", t, func() {
		a := []string{"hello", "goconvey"}
		b := []string{"hello", "goconvey"}
		So(util.StringSliceEqual(a, b), ShouldBeTrue)
	})

	Convey("should be return false", t, func() {
		a := []string{"hello", "123"}
		b := []string{"hello", "goconvey"}
		So(util.StringSliceEqual(a, b), ShouldBeFalse)
	})
}
