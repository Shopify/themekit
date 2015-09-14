package themekit

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"
)

const MessageSeparator string = "\n----------------------------------------------------------------\n"

var RedText = color.New(color.FgRed).SprintFunc()
var YellowText = color.New(color.FgYellow).SprintFunc()
var BlueText = color.New(color.FgBlue).SprintFunc()
var GreenText = color.New(color.FgGreen).SprintFunc()

func TestFixture(name string) string {
	return string(RawTestFixture(name))
}

func RawTestFixture(name string) []byte {
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
	return bytes
}

func BinaryTestData() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	buff := bytes.NewBuffer([]byte{})
	png.Encode(buff, img)
	return buff.Bytes()
}
