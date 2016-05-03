package link2

import (
	"net/http"
	"net/url"
	"testing"
)

import (
	"github.com/mikydna/x/redis"
	"golang.org/x/net/context"
)

var (
	google, _    = url.Parse("http://www.google.com/")
	yahoo, _     = url.Parse("https://www.yahoo.com/")
	altavista, _ = url.Parse("http://search.yahoo.com/?fr=altavista")

	testTableRedisExpand = []struct {
		url       string
		expected  *Result
		shouldErr bool
	}{
		{
			url: "http://google.com",
			expected: &Result{
				StatusCode:  200,
				ResolvedURL: google,
			},
			shouldErr: false,
		},
		{
			url: "http://yahoo.com",
			expected: &Result{
				StatusCode:  200,
				ResolvedURL: yahoo,
			},
			shouldErr: false,
		},
		{
			url: "http://altavista.com",
			expected: &Result{
				StatusCode:  200,
				ResolvedURL: altavista,
			},
			shouldErr: false,
		},
		{
			url:       "http://doesnotexist.really",
			expected:  nil,
			shouldErr: true,
		},
	}
)

var (
	testRedisConf = redis.Conf{
		Host:     "localhost:6379",
		Database: 12,
		Pool:     1,
	}
)

func TestRedisExpand(t *testing.T) {
	if testing.Short() {
		t.Skip("Requires redis database. Requires outbound http reqs")
	}

	client := http.DefaultClient
	expander, err := NewRedisExpander(testRedisConf, client, NoopContent)
	if err != nil {
		t.Fatal(err)
	}
	defer expander.FlushAll()

	{ // miss all

		for _, test := range testTableRedisExpand {
			ctx := context.TODO()
			result, err := expander.Expand(ctx, test.url)

			if result == nil {

				if test.expected != nil {
					t.Error("Unexpected non-nil expand result")
				}

				if test.shouldErr && (err == nil) {
					t.Error("Unexpected non error")
				}

			} else {

				if test.expected.ResolvedURL.String() != result.ResolvedURL.String() {
					t.Errorf("Unexpected expand resolved url: %v != %v", test.expected.ResolvedURL, result.ResolvedURL)
				}

				if test.expected.StatusCode != result.StatusCode {
					t.Errorf("Unexpected expand status code: %d != %d", test.expected.StatusCode, result.StatusCode)
				}

			}
		}

		if stats := expander.Stats(); stats["miss"] != 4 || stats["hit"] > 0 {
			t.Errorf("Unexpected hit/miss: miss %.0f != %.0f; hit %.0f != %.0f", stats["miss"], 4.0, stats["hit"], 0.0)
		}
	}

	{ // hit all

		for _, test := range testTableRedisExpand {
			ctx := context.TODO()
			result, err := expander.Expand(ctx, test.url)

			if result == nil {

				if test.expected != nil {
					t.Error("Unexpected non-nil expand result")
				}

				if test.shouldErr && (err == nil) {
					t.Error("Unexpected non error")
				}

			} else {

				if test.expected.ResolvedURL.String() != result.ResolvedURL.String() {
					t.Errorf("Unexpected expand resolved url: %v != %v", test.expected.ResolvedURL, result.ResolvedURL)
				}

				if test.expected.StatusCode != result.StatusCode {
					t.Errorf("Unexpected expand status code: %d != %d", test.expected.StatusCode, result.StatusCode)
				}

			}
		}

		if stats := expander.Stats(); stats["miss"] != 4 || stats["hit"] != 4 {
			t.Errorf("Unexpected hit/miss: miss%.0f != %.0f; hit %.0f != %.0f", stats["miss"], 4.0, stats["hit"], 4.0)
		}
	}

}
