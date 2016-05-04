package link2

import (
	"io"
	"net/http"
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
// - servers can be temporarily down.. expbkoff retry ?

type ContentFunc func(io.Reader) Content

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

func (e *Expander) Expand(ctx context.Context, url string) (result *Result, err error) {
	startedAt := time.Now()

	resp, httpErr := ctxhttp.Get(ctx, e.client, url)

	// non-nil err still requires processing
	if httpErr != nil {
		err = httpErr
	}

	responseTime := time.Since(startedAt)

	if resp != nil {
		var content Content
		if resp.Body != nil {
			content = e.content(resp.Body)
			defer resp.Body.Close()
		}

		result = &Result{
			ResolvedURL:  resp.Request.URL,
			ResponseTime: responseTime,
			StatusCode:   resp.StatusCode,
			Content:      content,
		}
	}

	return
}
