package link2

import (
	"bytes"
	"io/ioutil"
	"testing"
)

var testTableExtractBasic = []struct {
	file     string
	expected Content
}{
	{
		file: "./_testdata/google.resp",
		expected: map[ContentType]string{
			Title:       "Google",
			Description: "Search the world's information, including webpages, images, videos and more. Google has many special features to help you find exactly what you're looking for.",
			FavIcon:     "/images/branding/product/ico/googleg_lodp.ico",
		},
	},
	{
		file: "./_testdata/medium.resp",
		expected: map[ContentType]string{
			Title:       "Medium",
			Description: "", // loaded via js
			FavIcon:     "https://cdn-static-1.medium.com/_/fp/icons/favicon-medium.TAS6uQ-Y7kcKgi0xjcYHXw.ico",
		},
	},
}

func TestExtractBasic(t *testing.T) {
	for _, test := range testTableExtractBasic {
		b, err := ioutil.ReadFile(test.file)
		if err != nil {
			t.Fatalf("Can not read test file: %s", test.file)
		}

		body := bytes.NewReader(b)
		extracted := ExtractBasic(body)

		for _, ct := range []ContentType{Title, Description, FavIcon} {
			if extracted[ct] != test.expected[ct] {
				t.Errorf("Unexpected content for '%d': %s != %s", ct, extracted[ct], test.expected[ct])
			}
		}
	}
}
