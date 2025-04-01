package vmap

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

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
	TimeOffset     TimeOffset       `xml:"timeOffset,attr"`
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
	AdSystem   string       `xml:"AdSystem"`
	AdTitle    string       `xml:"AdTitle"`
	Impression []Impression `xml:"Impression"`
	Creatives  []Creative   `xml:"Creatives>Creative"`
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
	Duration       Duration        `xml:"Duration"` // TODO: Make into duration object
	TrackingEvents []TrackingEvent `xml:"TrackingEvents>Tracking"`
	MediaFiles     []MediaFile     `xml:"MediaFiles>MediaFile"`
	ClickThroughs  []ClickThrough  `xml:"VideoClicks>ClickThrough"`
	ClickTracking  []ClickTracking `xml:"VideoClicks>ClickTracking"`
	CustomClick    []CustomClick   `xml:"VideoClicks>CustomClick"`
}

type ClickThrough struct {
	Id   string `xml:"id,attr"`
	Text string `xml:",chardata"`
}

type ClickTracking struct {
	Id   string `xml:"id,attr"`
	Text string `xml:",chardata"`
}

type CustomClick struct {
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
	Codec     string `xml:"codec,attr"`
}

type Duration struct{ time.Duration }

func (d *Duration) UnmarshalText(data []byte) error {
	s := string(data)
	s = strings.TrimSpace(s)
	if s == "" {
		*d = Duration{}
		return nil
	}
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return fmt.Errorf("invalid duration format: %s", s)
	}
	// TODO: Figure this part out
	hours, minutes, seconds := parts[0], parts[1], parts[2]
	var sb strings.Builder
	dur := time.Duration(0)
	sb.WriteString(hours)
	sb.WriteString("h")
	sb.WriteString(minutes)
	sb.WriteString("m")
	// TODO: Handle seconds with decimal
	if strings.Contains(seconds, ".") {
		parts := strings.Split(seconds, ".")
		sb.WriteString(parts[0])
		sb.WriteString("s")
		sb.WriteString(parts[1])
		sb.WriteString("ms")
	} else {
		sb.WriteString(seconds)
		sb.WriteString("s")
	}
	dur, err := time.ParseDuration(sb.String())
	if err != nil {
		return fmt.Errorf("error parsing duration: %w", err)
	}
	*d = Duration{dur}
	return nil
}

// TimeOffset represents the time offset for an ad break in the VMAP document.
type TimeOffset struct {
	// If this is not nil, we're dealing with a duration offset.
	Duration *Duration

	//If not zero and duration is nil, it's a position number.
	// -1 is reserved for start, -2 for end.
	Position int

	// If duration is nil and position is zero, this is a percentage offset.
	Percent float32
}

const (
	OffsetStart = -1
	OffsetEnd   = -2
)

func (to *TimeOffset) UnmarshalText(data []byte) error {
	switch string(data) {
	case "start":
		to.Position = OffsetStart
	case "end":
		to.Position = OffsetEnd

	}
	if strings.HasSuffix(string(data), "%") {
		p, err := strconv.ParseInt(strings.TrimSuffix(string(data), "%"), 10, 8)
		if err != nil {
			return fmt.Errorf("error parsing percentage offset: %w", err)
		}
		to.Percent = float32(p) / 100
		return nil
	}
	if strings.HasPrefix(string(data), "#") {
		p, err := strconv.ParseInt(strings.TrimPrefix(string(data), "#"), 10, 8)
		if err != nil {
			return fmt.Errorf("error parsing position offset: %w", err)
		}
		to.Position = int(p)
		return nil
	}
	var d Duration
	to.Duration = &d
	return to.Duration.UnmarshalText(data)
}
