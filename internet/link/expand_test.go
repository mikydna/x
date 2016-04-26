package link

import (
	"testing"
)

type TestCase struct {
	URL         string
	Expected    string
	ExpectedErr bool
}

var TestTableExpander = []TestCase{
	TestCase{
		URL:      "http://buff.ly/1VzMRMR",
		Expected: "http://www.fastcodesign.com/3059211/how-to-design-a-wearable-for-lebron-james",
	},
	TestCase{
		URL:      "http://53eig.ht/1NPi5aD",
		Expected: "http://fivethirtyeight.com/features/today-is-clintons-chance-to-end-the-groundhog-day-campaign",
	},
	TestCase{
		URL:      "http://theatln.tc/1ZFJFhB",
		Expected: "http://www.theatlantic.com/business/archive/2016/03/how-trackers-make-leisure-like-work/471864",
	},
	TestCase{
		URL:      "http://nyti.ms/1YR7S3w",
		Expected: "http://www.nytimes.com/2016/04/20/business/economy/liberal-biases-too-may-block-progress-on-climate-change.html?_r=3",
	},
	TestCase{
		URL:      "http://bit.ly/26qsQM1",
		Expected: "http://vimeo.com/133127360",
	},
	TestCase{
		URL:      "http://www.nytimes.com/2016/04/26/books/review-in-don-delillos-zero-k-daring-to-outwit-death.html",
		Expected: "http://www.nytimes.com/2016/04/26/books/review-in-don-delillos-zero-k-daring-to-outwit-death.html?_r=4",
	},
	TestCase{
		URL:      "http://google.com/?z=1&a=0",
		Expected: "http://www.google.com?a=0&z=1",
	},
	TestCase{
		URL:         "http://i.dont.exist/?z=1&a=0",
		Expected:    "",
		ExpectedErr: true,
	},
}

func TestExpand(t *testing.T) {
	expander := NewLinkExpander(DefaultClient, []Format{RemoveUTMQueryParams, Normalize})
	for _, test := range TestTableExpander {
		result := expander.Expand(test.URL)

		if hasErr := result.Err != nil; hasErr != test.ExpectedErr {
			t.Error(result.Err)
		}

		if resolved := result.URL; resolved != nil && resolved.String() != test.Expected {
			t.Error("Unexpected expand result")
		}
	}
}
