package link

import (
	// "log"
	"net/http"
	"net/url"
	"testing"
	"time"
)

import (
	"golang.org/x/net/context"
)

func panicURLParse(rawurl string) *url.URL {
	parsed, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}

	return parsed
}

var expandTests = []struct {
	url       string
	expected  *Result
	shouldErr bool
}{

	// happy
	{
		url:       "http://google.com",
		shouldErr: false,
		expected: &Result{
			StatusCode:  200,
			ResolvedURL: panicURLParse("http://www.google.com/"),
		},
	},
	{
		url:       "http://yahoo.com",
		shouldErr: false,
		expected: &Result{
			StatusCode:  200,
			ResolvedURL: panicURLParse("https://www.yahoo.com/"),
		},
	},
	{
		url:       "http://altavista.com",
		shouldErr: false,
		expected: &Result{
			StatusCode:  200,
			ResolvedURL: panicURLParse("http://search.yahoo.com/?fr=altavista"),
		},
	},

	// worried
	{
		url:       "http://nyti.ms/1YR7S3w",
		shouldErr: true,
		expected: &Result{
			StatusCode:  303,
			ResolvedURL: panicURLParse("http://www.nytimes.com/2016/04/20/business/economy/liberal-biases-too-may-block-progress-on-climate-change.html?_r=3"),
		},
	},

	// sad
	{
		url:       "http://doesnotexist.really",
		shouldErr: true,
		expected:  nil,
	},
}

func TestExpand(t *testing.T) {
	if testing.Short() {
		t.Skip("Requires outbound http reqs")
	}

	expander := NewExpander(http.DefaultClient, ExtractBasic)

	for _, test := range expandTests {
		ctx := context.TODO()
		result, err := expander.Expand(ctx, test.url)

		if test.shouldErr && (err == nil) {
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

}

func TestExpand_WithContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Requires outbound http reqs")
	}

	var (
		result *Result
		err    error
	)

	expander := NewExpander(http.DefaultClient, ExtractBasic)

	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)

	go func() {
		result, err = expander.Expand(ctx, "http://google.com")
		defer cancel()
	}()

	select {
	case <-ctx.Done():

		if result == nil {
			t.Fatal("Unexpected nil result")
		}

		if err != nil {
			t.Fatalf("Unexpected err: %v", err)
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
