package themekit

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"
)

const MessageSeparator string = "\n----------------------------------------------------------------\n"

func RedText(s string) string {
	return fmt.Sprintf("\033[31m%s\033[0m", s)
}

func YellowText(s string) string {
	return fmt.Sprintf("\033[33m%s\033[0m", s)
}

func BlueText(s string) string {
	return fmt.Sprintf("\033[34m%s\033[0m", s)
}

func GreenText(s string) string {
	return fmt.Sprintf("\033[32m%s\033[0m", s)
}

func TestFixture(name string) string {
	path := fmt.Sprintf("fixtures/%s.json", name)
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
}

func BinaryTestData() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	buff := bytes.NewBuffer([]byte{})
	png.Encode(buff, img)
	return buff.Bytes()
}
