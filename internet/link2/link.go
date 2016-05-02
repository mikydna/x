package link2

import (
	"net/url"
	"time"
)

import (
	"golang.org/x/net/context"
)

type ContentType uint16

const (
	Title ContentType = iota
	Desciption
)

type Content struct {
	Values map[ContentType]string
}

type Result struct {
	StatusCode   int
	ResponseTime time.Duration
	ResolvedURL  *url.URL
	Content      *Content
}

type Expander interface {
	Expand(context.Context, string) (*Result, error)
}
