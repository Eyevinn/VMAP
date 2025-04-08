package vmap

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type VMAP struct {
	XMLName  xml.Name  `xml:"VMAP" json:"xmlName"`
	Text     string    `xml:",chardata" json:"text"`
	Vmap     string    `xml:"vmap,attr" json:"vmap"`
	Version  string    `xml:"version,attr" json:"version"`
	AdBreaks []AdBreak `xml:"AdBreak" json:"adBreaks"`
}

type AdBreak struct {
	AdSource       *AdSource        `xml:"AdSource" json:"adSource"`
	TrackingEvents *[]TrackingEvent `xml:"TrackingEvents>Tracking" json:"trackingEvents"`
	Id             string           `xml:"breakId,attr" json:"id"`
	BreakType      string           `xml:"breakType,attr" json:"breakType"`
	TimeOffset     TimeOffset       `xml:"timeOffset,attr" json:"timeOffset"`
}

type AdSource struct {
	VASTData *VASTData `xml:"VASTAdData"`
}

type TrackingEvent struct {
	Event string `xml:"event,attr" json:"event"`
	Text  string `xml:",chardata" json:"text"`
}

type VASTData struct {
	VAST *VAST `xml:"VAST" json:"vast"`
}

type VAST struct {
	Text                      string `xml:",chardata" json:"text"`
	Xsi                       string `xml:"xsi,attr" json:"xsi"`
	NoNamespaceSchemaLocation string `xml:"noNamespaceSchemaLocation,attr" json:"noNamespaceSchemaLocation"`
	Version                   string `xml:"version,attr" json:"version"`
	Ad                        []Ad   `xml:"Ad" json:"ad"`
}

type Ad struct {
	Id       string  `xml:"id,attr" json:"id"`
	Sequence int     `xml:"sequence,attr" json:"sequence"`
	InLine   *InLine `xml:"InLine" json:"inLine"`
}

type AdTagURI struct{}

type InLine struct {
	AdSystem   string       `xml:"AdSystem" json:"adSystem"`
	AdTitle    string       `xml:"AdTitle" json:"adTitle"`
	Impression []Impression `xml:"Impression" json:"impression"`
	Creatives  []Creative   `xml:"Creatives>Creative" json:"creatives"`
}

type Impression struct {
	Id   string `xml:"id,attr" json:"id"`
	Text string `xml:",chardata" json:"text"`
}

type Creative struct {
	Id            string         `xml:"id,attr" json:"id"`
	AdId          string         `xml:"adId,attr" json:"adId"`
	UniversalAdId *UniversalAdId `xml:"UniversalAdId" json:"universalAdId"`
	Linear        *Linear        `xml:"Linear" json:"linear"`
}

type UniversalAdId struct{}

type Linear struct {
	Duration       Duration        `xml:"Duration" json:"duration"`
	TrackingEvents []TrackingEvent `xml:"TrackingEvents>Tracking" json:"trackingEvents"`
	MediaFiles     []MediaFile     `xml:"MediaFiles>MediaFile" json:"mediaFiles"`
	ClickThroughs  []ClickThrough  `xml:"VideoClicks>ClickThrough" json:"clickThroughs"`
	ClickTracking  []ClickTracking `xml:"VideoClicks>ClickTracking" json:"clickTracking"`
	CustomClick    []CustomClick   `xml:"VideoClicks>CustomClick" json:"customClick"`
}

type ClickThrough struct {
	Id   string `xml:"id,attr" json:"id"`
	Text string `xml:",chardata" json:"text"`
}

type ClickTracking struct {
	Id   string `xml:"id,attr" json:"id"`
	Text string `xml:",chardata" json:"text"`
}

type CustomClick struct {
	Id   string `xml:"id,attr" json:"id"`
	Text string `xml:",chardata" json:"text"`
}

type MediaFile struct {
	Text      string `xml:",chardata" json:"text"`
	Bitrate   int    `xml:"bitrate,attr" json:"bitrate"`
	Width     int    `xml:"width,attr" json:"width"`
	Height    int    `xml:"height,attr" json:"height"`
	Delivery  string `xml:"delivery,attr" json:"delivery"`
	MediaType string `xml:"type,attr" json:"mediaType"`
	Codec     string `xml:"codec,attr" json:"codec"`
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

func (d Duration) MarshalText() ([]byte, error) {
	if d.Duration == 0 {
		return []byte(""), nil
	}
	hours := int(d.Duration.Hours())
	minutes := int(d.Duration.Minutes()) % 60
	seconds := int(d.Duration.Seconds()) % 60
	milliseconds := int(d.Duration.Milliseconds()) % 1000

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds))
	if milliseconds > 0 {
		sb.WriteString(fmt.Sprintf(".%03d", milliseconds))
	}
	return []byte(sb.String()), nil
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

func (to TimeOffset) MarshalText() ([]byte, error) {
	if to.Duration != nil {
		return to.Duration.MarshalText()
	}
	if to.Position != 0 {
		return []byte(fmt.Sprintf("#%d", to.Position)), nil
	}
	if to.Percent != 0 {
		return []byte(fmt.Sprintf("%f%%", to.Percent*100)), nil
	}
	return []byte(""), nil
}
