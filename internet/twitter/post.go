package twitter

import (
	"encoding/json"
	"net/url"
	"time"
)

import (
	"github.com/mikydna/z/x/stats"
)

type Post struct {
	Id        int
	UserId    int
	CreatedAt time.Time
	Text      string
	Urls      []url.URL
	Hashtags  []string
	Mentions  []int
	Retweet   *Post
	Stat      stats.Stat
}

func (p *Post) MarshalJSON() ([]byte, error) {
	type simplified struct {
		Id        int        `json:"id"`
		UserId    int        `json:"user"`
		CreatedAt int64      `json:"created_at"`
		Text      string     `json:"text"`
		Urls      []string   `json:"urls"`
		Hashtags  []string   `json:"hashtags"`
		Mentions  []int      `json:"mentions"`
		Retweet   *Post      `json:"retweet"`
		Stat      stats.Stat `json:"stat"`
	}

	urlStrs := make([]string, len(p.Urls))
	for i, url := range p.Urls {
		urlStrs[i] = url.String()
	}

	simple := simplified{
		Id:        p.Id,
		UserId:    p.UserId,
		CreatedAt: p.CreatedAt.UnixNano(),
		Text:      p.Text,
		Urls:      urlStrs,
		Hashtags:  p.Hashtags,
		Mentions:  p.Mentions,
		Retweet:   p.Retweet,
		Stat:      p.Stat,
	}

	return json.Marshal(&simple)
}

func (p *Post) UnmarshalJSON(b []byte) error {
	type simplified struct {
		Id        int        `json:"id"`
		UserId    int        `json:"user"`
		CreatedAt int64      `json:"created_at"`
		Text      string     `json:"text"`
		Urls      []string   `json:"urls"`
		Hashtags  []string   `json:"hashtags"`
		Mentions  []int      `json:"mentions"`
		Retweet   *Post      `json:"retweet"`
		Stat      stats.Stat `json:"stat"`
	}

	var intermediate simplified
	if err := json.Unmarshal(b, &intermediate); err != nil {
		return err
	}

	urls := make([]url.URL, len(intermediate.Urls))
	for i, urlStr := range intermediate.Urls {
		if parsed, err := url.Parse(urlStr); err == nil {
			urls[i] = *parsed
		} else {
			return err
		}
	}

	*p = Post{
		Id:        intermediate.Id,
		UserId:    intermediate.UserId,
		CreatedAt: time.Unix(0, intermediate.CreatedAt),
		Text:      intermediate.Text,
		Urls:      urls,
		Hashtags:  intermediate.Hashtags,
		Mentions:  intermediate.Mentions,
		Retweet:   intermediate.Retweet,
		Stat:      intermediate.Stat,
	}

	return nil
}

func FromTwitterPost(tp TwitterPost) Post {
	urls := []url.URL{}
	for _, entity := range tp.Entities.Urls {
		urlstr := entity.ExpandedUrl
		url, err := url.Parse(urlstr)
		if err != nil {
			continue
		}

		urls = append(urls, *url)
	}

	post := Post{
		Id:        tp.Id,
		UserId:    tp.User.Id,
		CreatedAt: tp.CreatedAt.Add(0),
		Text:      tp.Text,
		Urls:      urls,
		Stat:      *stats.NewStat(time.Now()),
	}

	post.Stat.Set("retweet", float64(tp.RetweetCount))
	post.Stat.Set("favorite", float64(tp.FavoriteCount))

	if tp.Retweet != nil {
		retweet := FromTwitterPost(*tp.Retweet)
		post.Retweet = &retweet
	}

	return post
}
