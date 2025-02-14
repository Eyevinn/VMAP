package vmap

import "encoding/xml"

type VMAP struct {
	XMLName  xml.Name  `xml:"VMAP"`
	Text     string    `xml:",chardata"`
	Vmap     string    `xml:"vmap,attr"`
	Version  string    `xml:"version,attr"`
	AdBreaks []AdBreak `xml:"AdBreak"`
}

type AdBreak struct {
	AdSource       *AdSource        `xml:"AdSource"`
	TrackingEvents *[]TrackingEvent `xml:"TrackingEvents>Tracking"`
	Id             string           `xml:"breakId,attr"`
	BreakType      string           `xml:"breakType,attr"`
	TimeOffset     string           `xml:"timeOffset,attr"`
}

type AdSource struct {
	VASTData *VASTData `xml:"VASTAdData"`
}

type TrackingEvent struct {
	Event string `xml:"event,attr"`
	Text  string `xml:",chardata"`
}

type VASTData struct {
	VAST *VAST `xml:"VAST"`
}

type VAST struct {
	Text                      string `xml:",chardata"`
	Xsi                       string `xml:"xsi,attr"`
	NoNamespaceSchemaLocation string `xml:"noNamespaceSchemaLocation,attr"`
	Version                   string `xml:"version,attr"`
	Ad                        []Ad   `xml:"Ad"`
}

type Ad struct {
	Id       string  `xml:"id,attr"`
	Sequence int     `xml:"sequence,attr"`
	InLine   *InLine `xml:"InLine"`
}

type AdTagURI struct{}

type InLine struct {
	AdSystem   string      `xml:"AdSystem"`
	AdTitle    string      `xml:"AdTitle"`
	Impression *Impression `xml:"Impression"`
	Creatives  []Creative  `xml:"Creatives>Creative"`
}

type Impression struct {
	Id   string `xml:"id,attr"`
	Text string `xml:",chardata"`
}

type Creative struct {
	Id            string         `xml:"id,attr"`
	AdId          string         `xml:"adId,attr"`
	UniversalAdId *UniversalAdId `xml:"UniversalAdId"`
	Linear        *Linear        `xml:"Linear"`
}

type UniversalAdId struct{}

type Linear struct {
	Duration       string          `xml:"Duration"` // TODO: Make into duration object
	TrackingEvents []TrackingEvent `xml:"TrackingEvents>Tracking"`
	MediaFiles     []MediaFile     `xml:"MediaFiles>MediaFile"`
}

type VideoClicks struct {
	ClickThrough []ClickThrough `xml:"ClickThrough"`
}

type ClickThrough struct {
	Id   string `xml:"id,attr"`
	Text string `xml:",chardata"`
}

type MediaFile struct {
	Text      string `xml:",chardata"`
	Bitrate   int    `xml:"bitrate,attr"`
	Width     int    `xml:"width,attr"`
	Height    int    `xml:"height,attr"`
	Delivery  string `xml:"delivery,attr"`
	MediaType string `xml:"type,attr"`
}
