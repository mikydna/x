package link2

import (
	"io"
	"strings"
)

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

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
			text := element.FirstChild.Data
			collect["title"] = append(collect["title"], text)

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
	}

	if descs := collect["description"]; len(descs) > 0 {
		content[Description] = descs[0]
	}

	return
}
