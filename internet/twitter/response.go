package twitter

import (
	"strconv"
	"strings"
	"time"
)

type RubyTimestamp struct {
	time.Time
}

func (t *RubyTimestamp) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), "\"")
	ts, err := time.Parse(time.RubyDate, str)
	if err != nil {
		return err
	}

	*t = RubyTimestamp{ts}

	return nil
}

type UnixTimestamp struct {
	time.Time
}

func (t *UnixTimestamp) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), "\"")
	unix, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return err
	}

	ts := time.Unix(unix, 0)

	*t = UnixTimestamp{ts}

	return nil
}

/* twitter "ratelimit" response struct */
type TwitterRateLimit struct {
	Context struct {
		Application string `json:"application"`
		AccessToken string `json:"access_token"`
	} `json:"rate_limit_context"`

	Resources map[string]map[string]struct {
		Limit     int           `json:"limit"`
		Remaining int           `json"remaining"`
		Reset     UnixTimestamp `json:"reset"`
	} `json:"resources"`
}

/* twitter "timeline" response struct */
type TwitterPost struct {
	Id        int           `json:"id"`
	CreatedAt RubyTimestamp `json:"created_at"`
	User      struct {
		Id int `json:"id"`
	} `json:"user"`

	Text     string `json:"text"`
	Entities struct {
		Urls []struct {
			Url         string `json:"url"`
			ExpandedUrl string `json:"expanded_url"`
			DisplayUrl  string `json:"display_url"`
		} `json:"urls"`
	} `json:"entities"`

	RetweetCount  int          `json:"retweet_count"`
	FavoriteCount int          `json:"favorite_count"`
	Retweet       *TwitterPost `json:"retweeted_status"`
}
