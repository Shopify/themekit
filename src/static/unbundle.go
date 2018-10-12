package static

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/colors"
)

var zipData string

// Register will set the zip data that can be decompressed
func Register(data string) {
	zipData = data
}

// Unbundle will saftely put all files in place without overwriting files that
// already exist
func Unbundle(ctx *cmdutil.Ctx) error {
	files, err := getZipContents(zipData)
	if err != nil {
		return err
	}
	return createDirs(ctx, files)
}

func getZipContents(data string) (map[string]map[string]*zip.File, error) {
	files := map[string]map[string]*zip.File{}
	zipReader, err := zip.NewReader(strings.NewReader(data), int64(len(data)))
	if err != nil {
		return files, err
	}
	for _, zipFile := range zipReader.File {
		dir := filepath.Dir(zipFile.Name)
		if _, ok := files[dir]; !ok {
			files[dir] = map[string]*zip.File{}
		}
		files[dir][zipFile.Name] = zipFile
	}
	return files, nil
}

func createDirs(ctx *cmdutil.Ctx, dirFiles map[string]map[string]*zip.File) error {
	for dir, files := range dirFiles {
		dir = filepath.Join(ctx.Flags.Directory, dir)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			ctx.Log.Printf("%s %s.\n", colors.Green("Created"), dir)
		} else {
			ctx.Log.Printf("%s %s.\n", colors.Blue("Exists"), dir)
		}
		for path, file := range files {
			path = filepath.Join(ctx.Flags.Directory, path)
			if err := writeFile(ctx, path, file); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeFile(ctx *cmdutil.Ctx, path string, zfile *zip.File) error {
	if _, err := os.Stat(path); err == nil {
		ctx.Log.Printf("\t%s %s.\n", colors.Blue("Exists"), path)
		return nil
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	defer file.Sync()

	contents, err := zfile.Open()
	if err != nil {
		return err
	}
	defer contents.Close()

	_, err = io.Copy(file, contents)
	if err == nil {
		ctx.Log.Printf("\t%s %s.\n", colors.Green("Created"), file.Name())
	}
	return nil
}
