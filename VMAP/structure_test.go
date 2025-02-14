package vmap

import (
	"encoding/xml"
	"io"
	"os"
	"testing"

	"github.com/matryer/is"
)

func TestMarshalVMAP(t *testing.T) {
	is := is.New(t)
	f, err := os.Open("sample-vmap/testVmap.xml")
	is.NoErr(err)
	defer f.Close()

	var vmap VMAP
	xmlBytes, err := io.ReadAll(f)
	is.NoErr(err)
	err = xml.Unmarshal(xmlBytes, &vmap)
	is.NoErr(err)
}
