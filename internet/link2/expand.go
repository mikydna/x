package link2

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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

func (r *Result) MarshalStringMap() (map[string]string, error) {
	strmap := make(map[string]string)
	strmap["statusCode"] = fmt.Sprintf("%d", r.StatusCode)
	strmap["responseTime"] = fmt.Sprintf("%d", r.ResponseTime.Nanoseconds())
	strmap["resolvedURL"] = r.ResolvedURL.String()

	for key, val := range r.Content {
		rkey := fmt.Sprintf("c_%d", key)
		strmap[rkey] = val
	}

	return strmap, nil
}

func (r *Result) UnmarshalStringMap(strmap map[string]string) error {
	var result Result

	if str, exists := strmap["statusCode"]; exists {
		statusCode, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			return err
		}

		result.StatusCode = int(statusCode)
	}

	if str, exists := strmap["responseTime"]; exists {
		responseTime, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}

		result.ResponseTime = time.Duration(responseTime)
	}

	if str, exists := strmap["resolvedURL"]; exists {
		resolvedURL, err := url.Parse(str)
		if err != nil {
			return err
		}

		result.ResolvedURL = resolvedURL
	}

	result.Content = make(Content)
	for _, key := range []ContentType{Title, Description} {
		rhkey := fmt.Sprintf("c_%d", key)
		if str, exists := strmap[rhkey]; exists {
			result.Content[key] = str
		}
	}

	*r = result

	return nil
}
