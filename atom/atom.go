package atom

import (
	"encoding/xml"
	"io"
)

// Feed ... TODO
type Feed struct {
	XMLName xml.Name `xml:"http://www.w3.org/2005/Atom feed"`
	Title   string   `xml:"title"`
	ID      string   `xml:"id"`
	Link    Link     `xml:"link"`
	Updated string   `xml:"updated"`
	Entries []Entry  `xml:"entry"`
}

// Entry ... TODO
type Entry struct {
	Title   string `xml:"title"`
	ID      string `xml:"id"`
	Updated string `xml:"updated"`
	Author  Person `xml:"author"`
	Link    Link   `xml:"link"`
	Summary Text   `xml:"content"`
}

// Link ... TODO
type Link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

// Person ... TODO
type Person struct {
	Name string `xml:"name"`
}

// Text ... TODO
type Text struct {
	Type string `xml:"type,attr"`
	Body string `xml:"chardata"`
}

// LoadFeed ... TODO
func LoadFeed(r io.Reader) (Feed, error) {
	var atom Feed
	decoder := xml.NewDecoder(r)
	err := decoder.Decode(&atom)
	return atom, err
}

// LatestEntry ... TODO
func (f Feed) LatestEntry() Entry {
	return f.Entries[0]
}
