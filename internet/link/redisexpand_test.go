package link

import (
	"net/http"
	"testing"
)

import (
	"github.com/mikydna/x/redis"
	"golang.org/x/net/context"
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

	{ // misses

		for _, test := range expandTests {
			ctx := context.TODO()
			result, err := expander.Expand(ctx, test.url)

			if (err == nil) && test.shouldErr {
				t.Error("Unexpected non error")
			}

			if result == nil {

				if test.expected != nil {
					t.Error("Unexpected non-nil expand result")
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

		if stats := expander.Stats(); stats["miss"] != 5 || stats["hit"] > 0 {
			t.Errorf("Unexpected hit/miss: miss %.0f != %d; hit %.0f != %d", stats["miss"], 5, stats["hit"], 0.0)
		}
	}

	{ // hits

		for _, test := range expandTests {
			ctx := context.TODO()
			result, err := expander.Expand(ctx, test.url)

			if (err == nil) && test.shouldErr {
				t.Errorf("Unexpected non error: %s", test.url)
			}

			if result == nil {

				if test.expected != nil {
					t.Errorf("Unexpected non-nil expand result: %s", test.url)
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

		if stats := expander.Stats(); stats["miss"] != 5 || stats["hit"] != 5 {
			t.Errorf("Unexpected hit/miss: miss %.0f != %d; hit %.0f != %d", stats["miss"], 5, stats["hit"], 5)
		}
	}

}
