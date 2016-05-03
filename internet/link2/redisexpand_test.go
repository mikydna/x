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

type testcase struct {
	URL       string
	Expected  *Result
	ShouldErr bool
}

var (
	google, _    = url.Parse("http://www.google.com/")
	yahoo, _     = url.Parse("https://www.yahoo.com/")
	altavista, _ = url.Parse("http://search.yahoo.com/?fr=altavista")

	TestTable = []testcase{
		testcase{
			URL: "http://google.com",
			Expected: &Result{
				StatusCode:  200,
				ResolvedURL: google,
			},
			ShouldErr: false,
		},
		testcase{
			URL: "http://yahoo.com",
			Expected: &Result{
				StatusCode:  200,
				ResolvedURL: yahoo,
			},
			ShouldErr: false,
		},
		testcase{
			URL: "http://altavista.com",
			Expected: &Result{
				StatusCode:  200,
				ResolvedURL: altavista,
			},
			ShouldErr: false,
		},
		testcase{
			URL:       "http://doesnotexist.really",
			Expected:  nil,
			ShouldErr: true,
		},
	}
)

var (
	TestRedisConf = redis.Conf{
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
	expander, err := NewRedisExpander(TestRedisConf, client, NoopContent)
	if err != nil {
		t.Fatal(err)
	}
	defer expander.FlushAll()

	{ // miss test

		for _, test := range TestTable {
			expected := test.Expected
			shouldErr := test.ShouldErr

			ctx := context.TODO()
			result, err := expander.Expand(ctx, test.URL)

			if result == nil {

				if expected != nil {
					t.Error("Unexpected non-nil expand result")
				}

				if shouldErr && (err == nil) {
					t.Error("Unexpected non error")
				}

			} else {

				if expected.ResolvedURL.String() != result.ResolvedURL.String() {
					t.Errorf("Unexpected expand resolved url: %v != %v", expected.ResolvedURL, result.ResolvedURL)
				}

				if expected.StatusCode != result.StatusCode {
					t.Errorf("Unexpected expand status code: %d != %d", expected.StatusCode, result.StatusCode)
				}

			}
		}

		if stats := expander.Stats(); stats["miss"] != 4 || stats["hit"] > 0 {
			t.Errorf("Unexpected hit/miss: miss %.0f != %.0f; hit %.0f != %.0f", stats["miss"], 4.0, stats["hit"], 0.0)
		}
	}

	{ // hit test

		for _, test := range TestTable {
			expected := test.Expected
			shouldErr := test.ShouldErr

			ctx := context.TODO()
			result, err := expander.Expand(ctx, test.URL)

			if result == nil {

				if expected != nil {
					t.Error("Unexpected non-nil expand result")
				}

				if shouldErr && (err == nil) {
					t.Error("Unexpected non error")
				}

			} else {

				if expected.ResolvedURL.String() != result.ResolvedURL.String() {
					t.Errorf("Unexpected expand resolved url: %v != %v", expected.ResolvedURL, result.ResolvedURL)
				}

				if expected.StatusCode != result.StatusCode {
					t.Errorf("Unexpected expand status code: %d != %d", expected.StatusCode, result.StatusCode)
				}

			}
		}

		if stats := expander.Stats(); stats["miss"] != 4 || stats["hit"] != 4 {
			t.Errorf("Unexpected hit/miss: miss%.0f != %.0f; hit %.0f != %.0f", stats["miss"], 4.0, stats["hit"], 4.0)
		}
	}
}
