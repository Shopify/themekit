package kit

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"

	"github.com/fatih/color"
)

// MessageSeparator ... TODO
const MessageSeparator string = "\n----------------------------------------------------------------\n"

// RedText ... TODO
var RedText = color.New(color.FgRed).SprintFunc()

// YellowText ... TODO
var YellowText = color.New(color.FgYellow).SprintFunc()

// BlueText ... TODO
var BlueText = color.New(color.FgBlue).SprintFunc()

// GreenText ... TODO
var GreenText = color.New(color.FgGreen).SprintFunc()

// TestFixture ... TODO
func TestFixture(name string) string {
	return string(RawTestFixture(name))
}

// RawTestFixture ... TODO
func RawTestFixture(name string) []byte {
	path := fmt.Sprintf("../fixtures/%s.json", name)
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

// BinaryTestData ... TODO
func BinaryTestData() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	buff := bytes.NewBuffer([]byte{})
	png.Encode(buff, img)
	return buff.Bytes()
}
