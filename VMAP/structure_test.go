package vmap

import (
	"encoding/xml"
	"io"
	"os"
	"testing"

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
	is.Equal(firstBreak.TimeOffset, "00:00:00")
	is.True(firstBreak.AdSource.VASTData.VAST != nil)
	is.Equal(len(*firstBreak.TrackingEvents), 1)

	secondBreak := vmap.AdBreaks[1]
	is.Equal(secondBreak.Id, "midroll.ad-2")
	is.Equal(secondBreak.BreakType, "linear")
	is.Equal(secondBreak.TimeOffset, "00:05:00")
	is.True(firstBreak.AdSource.VASTData.VAST != nil)
	is.Equal(len(*secondBreak.TrackingEvents), 1)

	thirdBreak := vmap.AdBreaks[2]
	is.Equal(thirdBreak.Id, "midroll.ad-3")
	is.Equal(thirdBreak.BreakType, "linear")
	is.Equal(thirdBreak.TimeOffset, "00:07:00")
	is.True(thirdBreak.AdSource.VASTData.VAST != nil)
	is.Equal(len(*thirdBreak.TrackingEvents), 1)
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
	firstAdImpression := firstAdInLine.Impression
	is.Equal(firstAdImpression.Id, "IMPRESSION-ID_001")
	firstAdCreatives := firstAdInLine.Creatives
	is.Equal(len(firstAdCreatives), 1)
	firstCreative := firstAdCreatives[0]
	is.Equal(firstCreative.Id, "CRETIVE-ID_001")
	is.Equal(firstCreative.AdId, "alvedon-10s")
	is.Equal(len(firstCreative.Linear.TrackingEvents), 5)
	is.Equal(firstCreative.Linear.Duration, "00:00:10")
	is.Equal(len(firstCreative.Linear.MediaFiles), 1)

	mediaFile := firstCreative.Linear.MediaFiles[0]
	is.Equal(mediaFile.Width, 718)
	is.Equal(mediaFile.Height, 404)
	is.Equal(mediaFile.MediaType, "video/mp4")
	is.Equal(mediaFile.Delivery, "progressive")
	is.Equal(mediaFile.Bitrate, 1300)
	is.Equal(mediaFile.Codec, "H.264")
}
