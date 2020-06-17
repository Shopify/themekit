package file

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// TODO  swap the order of these methods for easier reading?
func fileChecksum(dir, src string) (string, error) {
	fmt.Printf("fileChecksum for dir '%s', src '%s'\n", dir, src)
	sum := md5.New()
	s, err := os.Open(filepath.Join(dir, src))
	if err != nil {
		return "", err
	}
	defer s.Close()
	_, err = io.Copy(sum, s)
	result := fmt.Sprintf("%x", sum.Sum(nil))
	fmt.Printf("checksum is %s\n", result)
	return result, err
}

func dirSums(src string) (map[string]string, error) {
	fmt.Printf("dirSums\n")
	paths := map[string]string{}
	return paths, filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
	  fmt.Printf("Calculating checksum %s\n", path)
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
