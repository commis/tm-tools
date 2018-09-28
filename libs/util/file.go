package util

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/commis/tm-tools/libs/log"
)

func FileNameNoExt(fpath string) string {
	base := filepath.Base(fpath)
	return strings.TrimSuffix(base, filepath.Ext(fpath))
}

func GetParentDirectory(currDir string, level int) string {
	dirs := strings.Split(currDir, "/")
	return strings.Join(dirs[:len(dirs)-level], "/")
}

func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return 0, err
	}
	defer src.Close()

	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return 0, err
	}
	defer dst.Close()

	return io.Copy(dst, src)
}

func Rename(dstName, srcName string) {
	err := os.Rename(srcName, dstName)
	if err != nil {
		log.Errorf("failed to Rename file %s to %s", srcName, dstName)
	}
}
