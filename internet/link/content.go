package link

import (
	"io"
	// "io/ioutil"
	"log"
	"strings"
)

import (
	"golang.org/x/net/html"
)

// var flatten func(root *html.Node) []*html.Node
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

func ExtractTitle(body io.Reader) string {
	doc, err := html.Parse(body)
	if err != nil {
		return ""
	}

	content := flatten(doc)

	// collect data
	extracted := make(map[string][]string)
	for _, node := range content {
		tag := node.Data
		switch tag {
		case "title":

		case "meta":
			var name, content string
			for _, attr := range node.Attr {
				switch attr.Key {
				case "name", "property":
					split := strings.Split(attr.Val, ":")
					name = split[len(split)-1]
				case "content":
					content = attr.Val
				}
			}

			if name != "" && content != "" {
				extracted[name] = append(extracted[name], content)
				break
			}
		}
	}

	log.Println("title", extracted["title"])
	log.Println("description", extracted["description"])

	return ""
}
