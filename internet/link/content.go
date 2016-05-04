package link

import (
	"io"
	"strings"
)

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type ContentType uint16

const (
	Title ContentType = iota
	Description
	FavIcon
)

type Content map[ContentType]string

func NoopContent(io.Reader) Content {
	return make(Content)
}

func ExtractBasic(body io.Reader) (content Content) {
	doc, err := html.Parse(body)
	if err != nil {
		return
	}

	elements := flatten(doc)

	collect := make(map[string][]string)
	for _, element := range elements {
		switch element.DataAtom {
		case atom.Title:
			if element.FirstChild != nil {
				text := element.FirstChild.Data
				collect["_title"] = append(collect["_title"], text)
			}

		case atom.Link:
			var rel, href string
			for _, attr := range element.Attr {
				switch attr.Key {
				case "rel":
					rel = attr.Val
				case "href":
					href = attr.Val
				}
			}

			if rel != "" && href != "" {
				// <link href="/images/branding/product/ico/googleg_lodp.ico" rel="shortcut icon">
				if strings.Contains(rel, "shortcut") && strings.Contains(rel, "icon") {
					collect["favicon"] = append(collect["favicon"], href)
				}
			}

		case atom.Meta:
			// special case: pair meta attrs to form key-val pairs
			// - twitter, og: uses a non-standard 'property' tag name ?
			var name, content string
			for _, attr := range element.Attr {
				switch attr.Key {
				case "name", "property":
					// strip pfx, i.e., 'og:', 'twitter:'
					split := strings.Split(attr.Val, ":")
					name = split[len(split)-1]

				case "content":
					content = attr.Val
				}
			}

			if name != "" && content != "" {
				collect[name] = append(collect[name], content)
			}
		}
	}

	// (fix later)
	// - this is a little more involved. meta/og tags are often the cleanest strs
	//   but they are not always avail.
	// - (nice) if multiple entries, per type, exists, there is a non-trival
	//   decision to be made which one to use (a/b conv?, length, compare for dup
	//   words?)
	// - (nice) if you are stuck with only title tags, ideally, titles from the
	//   same tldp1 domain should be analyzed and stemmed to remove common pfx/sfx

	content = make(Content)

	if titles := collect["title"]; len(titles) > 0 {
		content[Title] = titles[0]

	} else if backupTitles := collect["_title"]; len(backupTitles) > 0 {
		content[Title] = collect["_title"][0]

	}

	if descs := collect["description"]; len(descs) > 0 {
		content[Description] = descs[0]
	}

	if favicons := collect["favicon"]; len(favicons) > 0 {
		content[FavIcon] = favicons[0]
	}

	return
}

func flatten(root *html.Node) []*html.Node {
	result := []*html.Node{root}
	for curr := root.FirstChild; curr != nil; curr = curr.NextSibling {
		switch curr.Type {
		case html.ElementNode:
			result = append(result, flatten(curr)...)
		}
	}

	return result
}
