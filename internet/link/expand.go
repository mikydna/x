package link

import (
	"net/http"
	"net/url"
	"time"
)

import (
	"golang.org/x/net/publicsuffix"
)

// var DefaultClient *http.Client = &http.Client{
// 	// Timeout: 2500 * time.Millisecond,
// }

var DefaultClient = http.DefaultClient

var DefaultExpander Expander = NewLinkExpander(
	DefaultClient,
	[]Format{RemoveUTMQueryParams, Normalize},
)

type Result struct {
	URL          *url.URL
	Domain       string
	Title        string
	Err          error
	StatusCode   int
	ResponseTime time.Duration
}

type Format func(url *url.URL) *url.URL

type Expander interface {
	Expand(string) *Result
}

type LinkExpander struct {
	client     *http.Client
	formatters []Format
}

func NewLinkExpander(client *http.Client, formatters []Format) *LinkExpander {
	expander := &LinkExpander{
		client:     client,
		formatters: formatters,
	}

	return expander
}

func (e *LinkExpander) Expand(rawurl string) (result *Result) {
	start := time.Now()
	resp, respErr := e.client.Get(rawurl)
	if resp == nil {
		result = &Result{Err: respErr}
		return
	}

	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()

	elapsed := time.Since(start)
	statusCode := resp.StatusCode

	var url *url.URL
	var err error

	switch statusCode {
	case http.StatusOK:
		url = resp.Request.URL
		err = nil

	case http.StatusFound:
		url, _ = resp.Location()
		err = nil

	case http.StatusSeeOther:
		url = resp.Request.URL
		err = nil

	case http.StatusNotFound:
		url = resp.Request.URL
		err = respErr

	default:
		if locUrl, _ := resp.Location(); locUrl != nil {
			url = locUrl
		} else if reqUrl := resp.Request.URL; reqUrl != nil {
			url = reqUrl
		}
	}

	if url != nil {
		for _, format := range e.formatters {
			url = format(url)
		}
	}

	var domain string
	if url != nil {
		domain, _ = publicsuffix.EffectiveTLDPlusOne(url.Host)
	}

	var title string

	if statusCode == 200 && resp.Body != nil {
		title = ExtractTitle(resp.Body)
	}

	result = &Result{
		URL:          url,
		Domain:       domain,
		Title:        title,
		Err:          err,
		StatusCode:   statusCode,
		ResponseTime: elapsed,
	}

	return
}
