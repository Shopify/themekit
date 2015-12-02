package atom

import (
	"encoding/xml"
	"io"
)

type Feed struct {
	XMLName xml.Name "http://www.w3.org/2005/Atom"
	Title   string   `xml:"title"`
	Id      string   `xml:"id"`
	Link    Link     `xml:"link"`
	Updated string   `xml:"updated"`
	Entries []Entry  `xml:"entry"`
}

type Entry struct {
	Title   string `xml:"title"`
	Id      string `xml:"id"`
	Updated string `xml:"updated"`
	Author  Person `xml:"author"`
	Link    Link   `xml:"link"`
	Summary Text   `xml:"content"`
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

type Person struct {
	Name string `xml:"name"`
}

type Text struct {
	Type string `xml:"type,attr"`
	Body string "chardata"
}

func LoadFeed(r io.Reader) (Feed, error) {
	var atom Feed
	decoder := xml.NewDecoder(r)
	err := decoder.Decode(&atom)
	return atom, err
}

func (f Feed) LatestEntry() Entry {
	return f.Entries[0]
}
