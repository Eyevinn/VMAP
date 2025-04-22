package vmap

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/muktihari/xmltokenizer"
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
	Text  string `xml:",chardata" json:"url"`
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

func DecodeVast(input []byte) (VAST, error) {
	var vast VAST
	found := false
	f := bytes.NewReader([]byte(input))

	tok := xmltokenizer.New(f)

	for {
		token, err := tok.Token() // Token is only valid until next tok.Token() invocation (short-lived object).
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		switch string(token.Name.Local) {
		case "VAST":
			found = true
			// Reuse Token object in the sync.Pool since we only use it temporarily.
			se := xmltokenizer.GetToken().Copy(token)
			err = vast.UnmarshalToken(tok, se)
			xmltokenizer.PutToken(se) // Put back to sync.Pool.
			if err != nil {
				return vast, err
			}
		}
	}

	if !found {
		return vast, errors.New("No VAST token found in document")
	}
	return vast, nil
}

func DecodeVmap(input []byte) (VMAP, error) {
	var vmap VMAP
	found := false

	f := bytes.NewReader([]byte(input))

	tok := xmltokenizer.New(f)

	for {
		token, err := tok.Token() // Token is only valid until next tok.Token() invocation (short-lived object).
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		switch string(token.Name.Local) {
		case "VMAP":
			found = true
			for i := range token.Attrs {
				attr := &token.Attrs[i]
				switch string(attr.Name.Local) {
				case "version":
					vmap.Version = string(attr.Value)
				case "vmap":
					vmap.Vmap = string(attr.Value)
				}
			}

		case "AdBreak":
			var adBreak AdBreak
			// Reuse Token object in the sync.Pool since we only use it temporarily.
			se := xmltokenizer.GetToken().Copy(token)
			err = adBreak.UnmarshalToken(tok, se)
			xmltokenizer.PutToken(se) // Put back to sync.Pool.
			if err != nil {
				return vmap, err
			}
			vmap.AdBreaks = append(vmap.AdBreaks, adBreak)
		}
	}

	if !found {
		return vmap, errors.New("No VMAP token found in document")
	}
	return vmap, nil
}

func (adBreak *AdBreak) UnmarshalToken(tok *xmltokenizer.Tokenizer, se *xmltokenizer.Token) error {
	adBreak.AdSource = &AdSource{
		VASTData: &VASTData{},
	}
	var err error
	for i := range se.Attrs {
		attr := &se.Attrs[i]
		switch string(attr.Name.Local) {
		case "breakId":
			adBreak.Id = string(attr.Value)
		case "breakType":
			adBreak.BreakType = string(attr.Value)
		case "timeOffset":
			err = adBreak.TimeOffset.UnmarshalText(attr.Value)
			if err != nil {
				return err
			}
		}
	}

	for {
		token, err := tok.Token()
		if err != nil {
			return err
		}
		if token.IsEndElementOf(se) { // Reach desired EndElement
			return nil
		}
		if token.IsEndElement { // Ignore child's EndElements
			continue
		}
		switch string(token.Name.Local) {
		case "VAST":
			var vast VAST
			// Reuse Token object in the sync.Pool since we only use it temporarily.
			se := xmltokenizer.GetToken().Copy(token)
			err = vast.UnmarshalToken(tok, se)
			xmltokenizer.PutToken(se) // Put back to sync.Pool.
			if err != nil {
				return err
			}
			adBreak.AdSource.VASTData.VAST = &vast
		case "Tracking":
			if adBreak.TrackingEvents == nil {
				adBreak.TrackingEvents = &[]TrackingEvent{}
			}
			var t TrackingEvent
			for i := range token.Attrs {
				attr := &token.Attrs[i]
				switch string(attr.Name.Local) {
				case "event":
					t.Event = string(attr.Value)
				}
			}
			t.Text = string(token.Data)
			*adBreak.TrackingEvents = append(*adBreak.TrackingEvents, t)
		}
	}
}

func (vast *VAST) UnmarshalToken(tok *xmltokenizer.Tokenizer, se *xmltokenizer.Token) error {
	for i := range se.Attrs {
		attr := &se.Attrs[i]
		switch string(attr.Name.Local) {
		case "version":
			vast.Version = string(attr.Value)
		}
	}

	for {
		token, err := tok.Token()
		if err != nil {
			return err
		}
		if token.IsEndElementOf(se) { // Reach desired EndElement
			return nil
		}
		if token.IsEndElement { // Ignore child's EndElements
			continue
		}
		switch string(token.Name.Local) {
		case "Ad":
			var ad Ad
			// Reuse Token object in the sync.Pool since we only use it temporarily.
			se := xmltokenizer.GetToken().Copy(token)
			err = ad.UnmarshalToken(tok, se)
			xmltokenizer.PutToken(se) // Put back to sync.Pool.
			if err != nil {
				return err
			}
			vast.Ad = append(vast.Ad, ad)
		}
	}
}

func (ad *Ad) UnmarshalToken(tok *xmltokenizer.Tokenizer, se *xmltokenizer.Token) error {
	for i := range se.Attrs {
		attr := &se.Attrs[i]
		switch string(attr.Name.Local) {
		case "sequence":
			seq, err := strconv.Atoi(string(attr.Value))
			if err != nil {
				return err
			}
			ad.Sequence = seq
		case "id":
			ad.Id = string(attr.Value)
		}
	}
	for {
		token, err := tok.Token()
		if err != nil {
			return err
		}
		if token.IsEndElementOf(se) { // Reach desired EndElement
			return nil
		}
		if token.IsEndElement { // Ignore child's EndElements
			continue
		}
		switch string(token.Name.Local) {
		case "InLine":
			var inline InLine
			// Reuse Token object in the sync.Pool since we only use it temporarily.
			se := xmltokenizer.GetToken().Copy(token)
			err = inline.UnmarshalToken(tok, se)
			xmltokenizer.PutToken(se) // Put back to sync.Pool.
			if err != nil {
				return err
			}
			ad.InLine = &inline
		}
	}
}

func (inline *InLine) UnmarshalToken(tok *xmltokenizer.Tokenizer, se *xmltokenizer.Token) error {
	for {
		token, err := tok.Token()
		if err != nil {
			return err
		}
		if token.IsEndElementOf(se) { // Reach desired EndElement
			return nil
		}
		if token.IsEndElement { // Ignore child's EndElements
			continue
		}
		//fmt.Println(string(token.Name.Local))
		switch string(token.Name.Local) {
		case "Creative":
			var c Creative
			se := xmltokenizer.GetToken().Copy(token)
			err = c.UnmarshalToken(tok, se)
			xmltokenizer.PutToken(se) // Put back to sync.Pool.
			if err != nil {
				return err
			}
			inline.Creatives = append(inline.Creatives, c)
		case "Impression":
			var imp Impression
			for i := range token.Attrs {
				attr := &token.Attrs[i]
				switch string(attr.Name.Local) {
				case "id":
					imp.Id = string(attr.Value)
				}
			}
			imp.Text = string(token.Data)
			inline.Impression = append(inline.Impression, imp)
		case "AdSystem":
			inline.AdSystem = string(token.Data)
		case "AdTitle":
			inline.AdTitle = string(token.Data)
		case "Extension":
			var e Extension
			// Reuse Token object in the sync.Pool since we only use it temporarily.
			se := xmltokenizer.GetToken().Copy(token)
			err = e.UnmarshalToken(tok, se)
			xmltokenizer.PutToken(se) // Put back to sync.Pool.
			if err != nil {
				return err
			}
			inline.Extensions = append(inline.Extensions, e)
		case "Error":
			var er Error
			er.Value = string(token.Data)
			inline.Error = &er

			//default:
			//fmt.Printf("%+v\n", x)

		}
	}
}

func (c *Creative) UnmarshalToken(tok *xmltokenizer.Tokenizer, se *xmltokenizer.Token) error {
	for i := range se.Attrs {
		attr := &se.Attrs[i]
		switch string(attr.Name.Local) {
		case "id":
			c.Id = string(attr.Value)
		case "adId":
			c.AdId = string(attr.Value)
		case "sequence":
			//TODO
		}
	}

	for {
		token, err := tok.Token()
		if err != nil {
			return err
		}
		if token.IsEndElementOf(se) { // Reach desired EndElement
			return nil
		}
		if token.IsEndElement { // Ignore child's EndElements
			continue
		}

		switch string(token.Name.Local) {
		case "UniversalAdId":
			var uaid UniversalAdId
			for i := range token.Attrs {
				attr := &token.Attrs[i]
				switch string(attr.Name.Local) {
				case "idRegistry":
					uaid.IdRegistry = string(attr.Value)
				}
			}
			uaid.Id = string(token.Data)
			c.UniversalAdId = &uaid
		case "Tracking":
			if c.Linear == nil {
				c.Linear = &Linear{}
			}
			var t TrackingEvent
			for i := range token.Attrs {
				attr := &token.Attrs[i]
				switch string(attr.Name.Local) {
				case "event":
					t.Event = string(attr.Value)
				}
			}
			t.Text = string(token.Data)
			c.Linear.TrackingEvents = append(c.Linear.TrackingEvents, t)
		case "ClickThrough":
			c.Linear.ClickThrough = &ClickThrough{}
			for i := range token.Attrs {
				attr := &token.Attrs[i]
				switch string(attr.Name.Local) {
				case "id":
					c.Linear.ClickThrough.Id = string(attr.Value)
				}
			}
			c.Linear.ClickThrough.Text = string(token.Data)
		case "ClickTracking":
			if c.Linear == nil {
				c.Linear = &Linear{}
			}
			var ct ClickTracking
			for i := range token.Attrs {
				attr := &token.Attrs[i]
				switch string(attr.Name.Local) {
				case "id":
					ct.Id = string(attr.Value)
				}
			}
			ct.Text = string(token.Data)
			c.Linear.ClickTracking = append(c.Linear.ClickTracking, ct)
		case "Duration":
			if c.Linear == nil {
				c.Linear = &Linear{}
			}
			c.Linear.Duration.UnmarshalText(token.Data)
		case "MediaFile":
			if c.Linear == nil {
				c.Linear = &Linear{}
			}
			var m MediaFile
			for i := range token.Attrs {
				attr := &token.Attrs[i]
				switch string(attr.Name.Local) {
				case "bitrate":
					m.Bitrate, err = strconv.Atoi(string(attr.Value))
					if err != nil {
						return err
					}
				case "height":
					m.Height, err = strconv.Atoi(string(attr.Value))
					if err != nil {
						return err
					}
				case "width":
					m.Width, err = strconv.Atoi(string(attr.Value))
					if err != nil {
						return err
					}
				case "delivery":
					m.Delivery = string(attr.Value)
				case "type":
					m.MediaType = string(attr.Value)
				case "codec":
					m.Codec = string(attr.Value)
				}
			}
			m.Text = string(token.Data)
			c.Linear.MediaFiles = append(c.Linear.MediaFiles, m)
		}
	}
}

func (ext *Extension) UnmarshalToken(tok *xmltokenizer.Tokenizer, se *xmltokenizer.Token) error {
	fmt.Println("Extension parsing", len(se.Attrs))
	for i := range se.Attrs {
		attr := &se.Attrs[i]
		fmt.Println("Reading extension attribs", string(attr.Name.Local))
		switch string(attr.Name.Local) {
		case "type":
			ext.ExtensionType = string(attr.Value)
		}
	}
	for {
		token, err := tok.Token()
		if err != nil {
			return err
		}
		if token.IsEndElementOf(se) { // Reach desired EndElement
			return nil
		}
		if token.IsEndElement { // Ignore child's EndElements
			continue
		}

		switch string(token.Name.Local) {
		case "CreativeParameter":
			fmt.Println("Parsing Creative Parameters", len(token.Attrs))
			var par CreativeParameter
			for i := range token.Attrs {
				attr := &token.Attrs[i]
				switch string(attr.Name.Local) {
				case "creativeId":
					par.CreativeId = string(attr.Value)
				case "name":
					par.Name = string(attr.Value)
				case "type":
					par.CreativeParameterType = string(attr.Value)
				}
			}
			par.Value = string(token.Data)
			fmt.Println(par.Value)
			par.CreativeParameterType = ext.ExtensionType
			ext.CreativeParameters = append(ext.CreativeParameters, par)
		}
	}
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
	Extensions []Extension  `xml:"Extensions>Extension" json:"extensions"`
	Error      *Error       `xml:"Error" json:"error"`
}

type Error struct {
	Value string `xml:",chardata" json:"value"`
}

type Impression struct {
	Id   string `xml:"id,attr" json:"id"`
	Text string `xml:",chardata" json:"url"`
}

type Creative struct {
	Id            string         `xml:"id,attr" json:"id"`
	AdId          string         `xml:"adId,attr" json:"adId"`
	UniversalAdId *UniversalAdId `xml:"UniversalAdId" json:"universalAdId"`
	Linear        *Linear        `xml:"Linear" json:"linear"`
}

type UniversalAdId struct {
	IdRegistry string `xml:"idRegistry,attr" json:"idRegistry"`
	Id         string `xml:",chardata" json:"id"`
}

type Linear struct {
	Duration       Duration        `xml:"Duration" json:"duration"`
	TrackingEvents []TrackingEvent `xml:"TrackingEvents>Tracking" json:"trackingEvents"`
	MediaFiles     []MediaFile     `xml:"MediaFiles>MediaFile" json:"mediaFiles"`
	ClickThrough   *ClickThrough   `xml:"VideoClicks>ClickThrough" json:"clickThrough"`
	ClickTracking  []ClickTracking `xml:"VideoClicks>ClickTracking" json:"clickTracking"`
	CustomClick    []CustomClick   `xml:"VideoClicks>CustomClick" json:"customClick"`
}

type ClickThrough struct {
	Id   string `xml:"id,attr" json:"id"`
	Text string `xml:",chardata" json:"url"`
}

type ClickTracking struct {
	Id   string `xml:"id,attr" json:"id"`
	Text string `xml:",chardata" json:"url"`
}

type CustomClick struct {
	Id   string `xml:"id,attr" json:"id"`
	Text string `xml:",chardata" json:"url"`
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

// NOTE: Specifically built for FreeWheel's CreativeParamer extension at the moment.
type Extension struct {
	ExtensionType      string              `xml:"type,attr" json:"type"`
	CreativeParameters []CreativeParameter `xml:"CreativeParameters>CreativeParameter" json:"creativeParameters"`
}

type CreativeParameter struct {
	CreativeId            string `xml:"creativeId,attr" json:"creativeId"`
	Name                  string `xml:"name,attr" json:"name"`
	Value                 string `xml:",chardata" json:"value"`
	CreativeParameterType string `xml:"type,attr" json:"creativeParameterType"`
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
		return []byte("00:00:00"), nil
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
		return nil
	case "end":
		to.Position = OffsetEnd
		return nil
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
