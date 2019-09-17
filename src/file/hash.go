package file

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func fileChecksum(dir, src string) (string, error) {
	sum := md5.New()
	s, err := os.Open(filepath.Join(dir, src))
	if err != nil {
		return "", err
	}
	defer s.Close()
	_, err = io.Copy(sum, s)
	return fmt.Sprintf("%x", sum.Sum(nil)), err
}

func dirSums(src string) (map[string]string, error) {
	paths := map[string]string{}
	return paths, filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if src != path && info.Mode().IsRegular() {
			checksum, err := fileChecksum("", path)
			if err != nil {
				return err
			}
			paths[pathToProject(src, path)] = checksum
		}
		return nil
	})
}
