package link2

import (
	"net/http"
	"testing"
	"time"
)

import (
	"golang.org/x/net/context"
)

func TestExpand(t *testing.T) {
	if testing.Short() {
		t.Skip("Requires outbound http calls")
	}

	var (
		result    *Result
		expandErr error
	)

	expander := NewExpander(http.DefaultClient, ExtractBasic)

	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)

	go func() {
		result, expandErr = expander.Expand(ctx, "http://google.com")
		defer cancel()
	}()

	select {
	case <-ctx.Done():

		if result == nil {
			t.Fatal("Unexpected nil result")
		}

		if result.StatusCode != 200 {
			t.Errorf("Unexpected statusCode: %d != %d", 200, result.StatusCode)
		}

		if result.ResolvedURL == nil || result.ResolvedURL.String() != "http://www.google.com/" {
			t.Errorf("Unexpected resolved url: %s != %s", "http://www.google.com/", result.ResolvedURL)
		}

	case <-time.After(3 * time.Second):
		t.Fatal("Unexpected timeout")
	}

}
