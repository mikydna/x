package link2

import (
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
// - servers can be temporarily down.. expbkoff retry

type doHandlerFunc func(*http.Response, error) error

func doAsync(ctx context.Context, req *http.Request, f doHandlerFunc) error {
	// ? create a new client for each req, need to access cancel
	t := &http.Transport{}
	c := &http.Client{
		Transport: t,
	}

	errs := make(chan error, 1)
	go func() {
		errs <- f(c.Do(req))
	}()

	select {
	case <-ctx.Done():
		t.CancelRequest(req)
		<-errs
		return ctx.Err()
	case err := <-errs:
		return err
	}
}

func Expand(ctx context.Context, client *http.Client, rawurl string) (*Result, error) {
	startedAt := time.Now()
	resp, err := ctxhttp.Get(ctx, client, rawurl)
	if err != nil {
		return nil, err
	}

	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	result := &Result{
		ResolvedURL:  resp.Request.URL,
		ResponseTime: time.Since(startedAt),
		StatusCode:   resp.StatusCode,
	}

	// req, err := http.NewRequest("GET", rawurl, nil)
	// if err != nil {
	// 	return nil, err
	// }

	// var result Result
	// err = doAsync(ctx, req, func(resp *http.Response, err error) error {
	// 	if resp != nil && resp.Body != nil {
	// 		defer resp.Body.Close()
	// 	}

	// 	result.StatusCode = resp.StatusCode
	// 	result.ResolvedURL = resp.Request.URL

	// 	return err
	// })

	return result, nil
}
