package rssfilter

import "encoding/xml"

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Atom    string   `xml:"xmlns:atom,attr,omitempty"`
	DC      string   `xml:"xmlns:dc,attr,omitempty"`
	Content string   `xml:"xmlns:content,attr,omitempty"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	AtomLink      *AtomLink `xml:"atom:link,omitempty"`
	Title         string    `xml:"title,omitempty"`
	Description   string    `xml:"description,omitempty"`
	Link          string    `xml:"link,omitempty"`
	Language      string    `xml:"language,omitempty"`
	LastBuildDate string    `xml:"lastBuildDate,omitempty"`
	Items         []Item    `xml:"item"`
}

type AtomLink struct {
	Href string `xml:"href,attr,omitempty"`
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
}

type Item struct {
	Title       string     `xml:"title,omitempty"`
	Description string     `xml:"description,omitempty"`
	Link        string     `xml:"link,omitempty"`
	GUID        string     `xml:"guid,omitempty"`
	PubDate     string     `xml:"pubDate,omitempty"`
	Enclosure   *Enclosure `xml:"enclosure,omitempty"`
}

type Enclosure struct {
	URL    string `xml:"url,attr,omitempty"`
	Type   string `xml:"type,attr,omitempty"`
	Length string `xml:"length,attr,omitempty"`
}
