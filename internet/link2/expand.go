package link2

import (
	"io"
	"net/http"
	"net/url"
	"time"
)

import (
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

// expansion is expensive
// - network is a limited resource
// - it pretty slow (200-3000s)

// expansion "resolve" logic *can* to be domain specific
// - nytimes has 10+ redirects + pay-wall

// expansion results should expire/retry-later
// - servers can be temporarily down.. expbkoff retry

type ContentFunc func(io.Reader) Content

type Result struct {
	StatusCode   int
	ResponseTime time.Duration
	ResolvedURL  *url.URL
	Content      Content
}

type Expander struct {
	client  *http.Client
	content ContentFunc
}

func NewExpander(client *http.Client, processor ContentFunc) *Expander {
	expander := Expander{
		client:  client,
		content: processor,
	}

	return &expander
}

func (e *Expander) Expand(ctx context.Context, url string) (*Result, error) {
	startedAt := time.Now()
	resp, err := ctxhttp.Get(ctx, e.client, url)
	if err != nil {
		return nil, err
	}

	responseTime := time.Since(startedAt)

	var content Content
	if resp != nil && resp.Body != nil {
		content = e.content(resp.Body)
		defer resp.Body.Close()
	}

	result := &Result{
		ResolvedURL:  resp.Request.URL,
		ResponseTime: responseTime,
		StatusCode:   resp.StatusCode,
		Content:      content,
	}

	return result, nil
}
