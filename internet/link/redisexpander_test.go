package link

import (
	"testing"
)

import (
	"github.com/mikydna/x/redis"
)

var (
	TestRedisConf = redis.Conf{
		Host:     "localhost:6379",
		Database: 12,
		Pool:     10,
	}

	TestTableRedisExpander = []TestCase{
		TestCase{
			URL:      "http://buff.ly/1VzMRMR",
			Expected: "http://www.fastcodesign.com/3059211/how-to-design-a-wearable-for-lebron-james",
		},
		TestCase{
			URL:      "http://53eig.ht/1NPi5aD",
			Expected: "http://fivethirtyeight.com/features/today-is-clintons-chance-to-end-the-groundhog-day-campaign",
		},
		TestCase{
			URL:      "http://www.nytimes.com/2016/04/26/books/review-in-don-delillos-zero-k-daring-to-outwit-death.html",
			Expected: "http://www.nytimes.com/2016/04/26/books/review-in-don-delillos-zero-k-daring-to-outwit-death.html?_r=4",
		},
		TestCase{
			URL:         "http://i.dont.exist/?z=1&a=0",
			Expected:    "",
			ExpectedErr: true,
		},
	}
)

func TestRedisExpander(t *testing.T) {
	cached, err := NewRedisExpander(TestRedisConf, DefaultExpander)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer cached.FlushAll()

	for _, test := range TestTableRedisExpander {
		result := cached.Expand(test.URL)

		if hasErr := (result.Err != nil); hasErr != test.ExpectedErr {
			t.Error(result.Err)
		}

		if resolved := result.URL; resolved != nil && resolved.String() != test.Expected {
			t.Error("Unexpected expand result")
		}
	}

	if stats := cached.Stats(); stats.Values["miss"] != 4 {
		t.Error("Unexpected cache state")
	}

	for _, test := range TestTableRedisExpander {
		result := cached.Expand(test.URL)

		if hasErr := (result.Err != nil); hasErr != test.ExpectedErr {
			t.Error(result.Err)
		}

		if resolved := result.URL; resolved != nil && resolved.String() != test.Expected {
			t.Error("Unexpected expand result")
		}

	}

	if stats := cached.Stats(); stats.Values["miss"] != 4 && stats.Values["hit"] != 4 {
		t.Error("Unexpected cache state")
	}

}
