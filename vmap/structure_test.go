package vmap

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestUnmarshalVMAP(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-vmap/testVmap.xml")
	is.NoErr(err)
	defer f.Close()

	var vmap VMAP
	xmlBytes, err := io.ReadAll(f)
	is.NoErr(err)
	err = xml.Unmarshal(xmlBytes, &vmap)
	is.NoErr(err)

	is.Equal(len(vmap.AdBreaks), 3)
	firstBreak := vmap.AdBreaks[0]
	is.Equal(firstBreak.Id, "midroll.ad-1")
	is.Equal(firstBreak.BreakType, "linear")
	is.True(firstBreak.TimeOffset.Duration == nil)
	is.Equal(firstBreak.TimeOffset.Position, OffsetStart)
	is.True(firstBreak.AdSource.VASTData.VAST != nil)
	is.Equal(len(firstBreak.TrackingEvents), 1)

	secondBreak := vmap.AdBreaks[1]
	is.Equal(secondBreak.Id, "midroll.ad-2")
	is.Equal(secondBreak.BreakType, "linear")
	is.Equal(*secondBreak.TimeOffset.Duration, Duration{5 * time.Minute})
	is.True(firstBreak.AdSource.VASTData.VAST != nil)
	is.Equal(len(secondBreak.TrackingEvents), 1)

	thirdBreak := vmap.AdBreaks[2]
	is.Equal(thirdBreak.Id, "midroll.ad-3")
	is.Equal(thirdBreak.BreakType, "linear")
	is.Equal(*thirdBreak.TimeOffset.Duration, Duration{7 * time.Minute})
	is.True(thirdBreak.AdSource.VASTData.VAST != nil)
	is.Equal(len(thirdBreak.TrackingEvents), 1)
}

func TestUnmarshalVast(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-vmap/testVast.xml")
	is.NoErr(err)
	defer f.Close()

	var vast VAST
	xmlBytes, err := io.ReadAll(f)
	is.NoErr(err)
	err = xml.Unmarshal(xmlBytes, &vast)
	is.NoErr(err)

	is.Equal(len(vast.Ad), 2)
	firstAd := vast.Ad[0]
	is.Equal(firstAd.Id, "POD_AD-ID_001")
	firstAdInLine := firstAd.InLine
	is.Equal(firstAdInLine.AdSystem, "Test Adserver")
	is.Equal(firstAdInLine.AdTitle, "Ad That Test-Adserver Wants Player To See #1")

	// Error validation
	firstAdError := firstAdInLine.Error
	is.True(firstAdError != nil)
	is.Equal(firstAdError.Value, "https://error-url/code")
	// Extension validation
	firstAdExtensions := firstAdInLine.Extensions
	is.Equal(len(firstAdExtensions), 1)
	firstAdExtension := firstAdExtensions[0]
	is.Equal(firstAdExtension.ExtensionType, "FreeWheel")
	firstAdExtensionCParams := firstAdExtension.CreativeParameters[0]
	is.Equal(firstAdExtensionCParams.CreativeId, "132285420")
	is.Equal(firstAdExtensionCParams.Name, "AdType")
	is.Equal(firstAdExtensionCParams.Value, "bumper")
	is.Equal(firstAdExtensionCParams.CreativeParameterType, "Linear")
	// Impression validation
	firstAdImpression := firstAdInLine.Impression
	is.Equal(len(firstAdImpression), 1)
	// Creatives validation
	firstAdCreatives := firstAdInLine.Creatives
	is.Equal(len(firstAdCreatives), 1)
	firstCreative := firstAdCreatives[0]
	is.Equal(firstCreative.Id, "CRETIVE-ID_001")
	is.Equal(firstCreative.AdId, "alvedon-10s")
	is.Equal(len(firstCreative.Linear.TrackingEvents), 5)
	is.Equal(firstCreative.Linear.Duration, Duration{10 * time.Second})
	is.Equal(len(firstCreative.Linear.MediaFiles), 1)
	is.True(firstCreative.Linear.ClickThrough != nil)
	is.Equal(len(firstCreative.Linear.ClickTracking), 0)
	is.Equal(len(firstCreative.Linear.CustomClick), 0)
	// MediaFile validation
	mediaFile := firstCreative.Linear.MediaFiles[0]
	is.Equal(mediaFile.Width, 718)
	is.Equal(mediaFile.Height, 404)
	is.Equal(mediaFile.MediaType, "video/mp4")
	is.Equal(mediaFile.Delivery, "progressive")
	is.Equal(mediaFile.Bitrate, 1300)
	is.Equal(mediaFile.Codec, "H.264")
}

func TestUnmarshalDuration(t *testing.T) {
	is := is.New(t)
	d := Duration{}
	err := d.UnmarshalText([]byte("00:00:10"))

	is.NoErr(err)
	is.Equal(d.Duration, 10*time.Second)
	err = d.UnmarshalText([]byte("00:01:00"))
	is.Equal(d.Duration, 1*time.Minute)
	is.NoErr(err)

	err = d.UnmarshalText([]byte("00:01:00.300"))
	is.NoErr(err)
	is.Equal(d.Duration, 1*time.Minute+300*time.Millisecond)

	err = d.UnmarshalText([]byte("04:01:12.345"))
	is.NoErr(err)
	is.Equal(d.Duration, 4*time.Hour+1*time.Minute+12*time.Second+345*time.Millisecond)
}

func TestMarshalJson(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-vmap/testVmap.xml")
	is.NoErr(err)
	defer f.Close()

	var vmap VMAP
	xmlBytes, err := io.ReadAll(f)
	is.NoErr(err)
	err = xml.Unmarshal(xmlBytes, &vmap)
	is.NoErr(err)
	jsonBytes, err := json.Marshal(vmap)
	is.NoErr(err)
	is.True(json.Valid(jsonBytes))

	var vmap2 VMAP
	err = json.Unmarshal(jsonBytes, &vmap2)
	is.NoErr(err)
	is.Equal(vmap, vmap2)
}
