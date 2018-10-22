package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/commis/tm-tools/libs/log"
)

func FileNameNoExt(fpath string) string {
	base := filepath.Base(fpath)
	return strings.TrimSuffix(base, filepath.Ext(fpath))
}

func GetParentDir(currDir string, level int) string {
	dirs := strings.Split(currDir, "/")
	return strings.Join(dirs[:len(dirs)-level], "/")
}

func GetChildDir(root string) []string {
	subdirs := make([]string, 0)
	files, _ := ioutil.ReadDir(root)
	for _, f := range files {
		if f.IsDir() {
			subdirs = append(subdirs, f.Name())
		}
	}
	return subdirs
}

func Rename(dstName, srcName string) {
	err := os.Rename(srcName, dstName)
	if err != nil {
		log.Errorf("failed to Rename file %s, %v", srcName, err)
	}
}

func CreateDirAll(dir string) bool {
	if !Exist(dir) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Printf("failed to create dir %s\n", dir)
			return false
		}
	}
	return true
}

func Exist(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return true
}
