package util_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/commis/tm-tools/libs/util"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFileNameNoExt(t *testing.T) {
	Convey("should equal /tmp", t, func() {
		rootDir := "/tmp/util_test.txt"
		defer removeTestDir(rootDir)
		So(util.FileNameNoExt(rootDir), ShouldEqual, "util_test")
	})
}

func TestGetParentDir(t *testing.T) {
	Convey("should equal /tmp", t, func() {
		rootDir := "/tmp/util_test"
		defer removeTestDir(rootDir)
		So(util.GetParentDir(rootDir, 1), ShouldEqual, "/tmp")
	})
}

func TestGetChildDir(t *testing.T) {
	rootDir := "/tmp/util_test"
	defer removeTestDir(rootDir)

	Convey("get child dir should be equal", t, func() {
		names := make([]string, 0)
		for i := 0; i < 5; i++ {
			dirName := strconv.Itoa(i)
			if util.CreateDirAll(rootDir + "/" + dirName) {
				names = append(names, dirName)
			}
		}
		dirs := util.GetChildDir(rootDir)
		So(util.StringSliceEqual(dirs, names), ShouldBeTrue)
	})

	Convey("get child dir should be not equal", t, func() {
		names := make([]string, 0)
		names = append(names, "yy")
		for i := 0; i < 3; i++ {
			dirName := strconv.Itoa(i)
			if util.CreateDirAll(rootDir + "/" + dirName) {
				names = append(names, dirName)
			}
		}
		dirs := util.GetChildDir(rootDir)
		So(util.StringSliceEqual(dirs, names), ShouldBeFalse)
	})
}

func TestRename(t *testing.T) {
	Convey("should exist new name", t, func() {
		rootDir := "/tmp/util_test"
		if util.CreateDirAll(rootDir) {
			newDir := rootDir + ".new"
			util.Rename(newDir, rootDir)
			defer removeTestDir(newDir)
			So(util.Exist(newDir), ShouldBeTrue)
		}
	})
}

func removeTestDir(dir string) {
	if util.Exist(dir) {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Printf("failed to remove dir %s\n", dir)
		}
	}
}
