package vmap

import (
	"bytes"
	"errors"
	"io"
	"strconv"

	"github.com/CarlLindqvist/xmltokenizer"
)

func DecodeVast(input []byte) (VAST, error) {
	var vast VAST
	found := false
	f := bytes.NewReader([]byte(input))

	tok := xmltokenizer.New(f, xmltokenizer.WithAttrBufferSize(5))

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
			if token.SelfClosing {
				break
			}
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
		return vast, errors.New("no VAST token found in document")
	}
	return vast, nil
}

func DecodeVmap(input []byte) (VMAP, error) {
	var vmap VMAP
	found := false

	f := bytes.NewReader([]byte(input))

	tok := xmltokenizer.New(f, xmltokenizer.WithAttrBufferSize(5))

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
					vmap.XMLName.Space = string(attr.Value)
				}
				vmap.XMLName.Local = "VMAP"
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
		return vmap, errors.New("no VMAP token found in document")
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
			if token.SelfClosing {
				adBreak.AdSource.VASTData.VAST = &VAST{}
				break
			}
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
				adBreak.TrackingEvents = []TrackingEvent{}
			}
			var t TrackingEvent
			for i := range token.Attrs {
				attr := &token.Attrs[i]
				switch string(attr.Name.Local) {
				case "event":
					t.Event = string(attr.Value)
				}
			}
			if token.WasCDATA {
				t.Text = string(token.Data)
			} else {
				t.Text = string(xmlStringToString(token.Data))
			}
			adBreak.TrackingEvents = append(adBreak.TrackingEvents, t)
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
			if token.WasCDATA {
				imp.Text = string(token.Data)
			} else {
				imp.Text = string(xmlStringToString(token.Data))
			}
			inline.Impression = append(inline.Impression, imp)
		case "AdSystem":
			if token.WasCDATA {
				inline.AdSystem = string(token.Data)
			} else {
				inline.AdSystem = string(xmlStringToString(token.Data))
			}
		case "AdTitle":
			if token.WasCDATA {
				inline.AdTitle = string(token.Data)
			} else {
				inline.AdTitle = string(xmlStringToString(token.Data))
			}
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
			if token.WasCDATA {
				er.Value = string(token.Data)
			} else {
				er.Value = string(xmlStringToString(token.Data))
			}
			inline.Error = &er
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
			if token.WasCDATA {
				uaid.Id = string(token.Data)
			} else {
				uaid.Id = string(xmlStringToString(token.Data))
			}
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
			if token.WasCDATA {
				t.Text = string(token.Data)
			} else {
				t.Text = string(xmlStringToString(token.Data))
			}
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
			if token.WasCDATA {
				c.Linear.ClickThrough.Text = string(token.Data)
			} else {
				c.Linear.ClickThrough.Text = string(xmlStringToString(token.Data))
			}
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
			if token.WasCDATA {
				ct.Text = string(token.Data)
			} else {
				ct.Text = string(xmlStringToString(token.Data))
			}
			c.Linear.ClickTracking = append(c.Linear.ClickTracking, ct)
		case "Duration":
			if c.Linear == nil {
				c.Linear = &Linear{}
			}
			if token.WasCDATA {
				err = c.Linear.Duration.UnmarshalText(token.Data)
			} else {
				err = c.Linear.Duration.UnmarshalText(xmlStringToString(token.Data))
			}

			if err != nil {
				return err
			}
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
			if token.WasCDATA {
				m.Text = string(token.Data)
			} else {
				m.Text = string(xmlStringToString(token.Data))
			}
			c.Linear.MediaFiles = append(c.Linear.MediaFiles, m)
		}
	}
}

func (ext *Extension) UnmarshalToken(tok *xmltokenizer.Tokenizer, se *xmltokenizer.Token) error {
	for i := range se.Attrs {
		attr := &se.Attrs[i]
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
			if token.WasCDATA {
				par.Value = string(token.Data)
			} else {
				par.Value = string(xmlStringToString(token.Data))
			}
			ext.CreativeParameters = append(ext.CreativeParameters, par)
		}
	}
}

func xmlStringToString(input []byte) []byte {
	o := 0
	for i := 0; i < len(input); i++ {
		b := input[i]

		switch b {
		//If we see a '&' we have a special character that needs decoding
		case '&':
			cb := make([]byte, 0, 4)
		specialCharLoop:
			for {
				i++
				if i >= len(input) {
					break
				}

				c := input[i]
				switch c {
				case '#', 'x':
				case ';':
					break specialCharLoop
				default:
					cb = append(cb, c)
				}
			}
			ch := decodeSpecialCharacterFromHexCode(cb)
			for _, l := range []byte(string(ch)) {
				input[o] = l
				o++
			}
		//This is just a normal byte, just output it
		default:
			input[o] = b
			o++
		}
	}
	return input[0:o]
}

func decodeSpecialCharacterFromHexCode(input []byte) rune {
	// Handle &amp; &lt; &gt; &apos; &quot;
	switch string(input) {
	case "amp":
		return '&'
	case "lt":
		return '<'
	case "gt":
		return '>'
	case "apos":
		return '\''
	case "quot":
		return '"'
	}
	codePoint, _ := strconv.ParseInt(string(input), 16, 32)
	return rune(codePoint)
}
