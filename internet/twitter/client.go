package twitter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

import (
	"github.com/mikydna/x/concurrent"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type OAuthClient struct {
	*http.Client
}

func NewOAuth2Client(key, secret, tokenUrl string) *OAuthClient {
	creds := &clientcredentials.Config{
		ClientID:     key,
		ClientSecret: secret,
		TokenURL:     tokenUrl,
	}

	return &OAuthClient{creds.Client(oauth2.NoContext)}
}

type TwitterResource struct {
	Host         string
	Family       string
	Version      string
	Path         string
	Query        string
	ResponseType string
}

type TwitterResourceKey struct {
	Family string
	Path   string
}

func (r TwitterResource) Key() TwitterResourceKey {
	return TwitterResourceKey{
		Family: r.Family,
		Path:   r.Path,
	}
}

func (r TwitterResource) Url(query map[string]string) *url.URL {
	resourceUrl, _ := url.Parse(r.Host)
	resourceUrl.Path = fmt.Sprintf("/%s/%s.%s", r.Version, r.Path, r.ResponseType)

	newQuery, _ := url.ParseQuery(r.Query)
	for key, value := range query {
		newQuery.Add(key, value)
	}

	resourceUrl.RawQuery = newQuery.Encode()

	return resourceUrl
}

var (
	TwitterRequestWait     = 5 * time.Second
	TwitterRefreshInterval = 5 * time.Second

	ErrRequestTimeout = errors.New("Request timeout")
	ErrCouldNotStart  = errors.New("Cound not start")
)

var (
	TwitterTokenUrl = "https://api.twitter.com/oauth2/token"

	// https://api.twitter.com/1.1/statuses/user_timeline.json
	TimelineResource TwitterResource = TwitterResource{
		Host:         "https://api.twitter.com",
		Version:      "1.1",
		Family:       "statuses",
		Path:         "/statuses/user_timeline",
		Query:        "count=200&trim_user=true",
		ResponseType: "json",
	}

	// https://api.twitter.com/1.1/application/rate_limit_status.json
	RateLimitResource TwitterResource = TwitterResource{
		Host:         "https://api.twitter.com",
		Version:      "1.1",
		Family:       "application",
		Path:         "/application/rate_limit_status",
		Query:        "",
		ResponseType: "json",
	}
)

type APIClient interface {
	Timeline(int) ([]byte, error)
}

type TwitterClient struct {
	permits map[TwitterResourceKey]*concurrent.Semaphore

	*OAuthClient
}

func Non200Error(statusCode int, req string) error {
	message := fmt.Sprintf("Recieved non-200: statusCode=%d url=%s", statusCode, req)
	return errors.New(message)
}

func NewClient(key, secret string) (*TwitterClient, error) {
	oauthClient := NewOAuth2Client(key, secret, TwitterTokenUrl)
	twitterClient := &TwitterClient{
		make(map[TwitterResourceKey]*concurrent.Semaphore),
		oauthClient,
	}

	// ? fix
	go func() {
		interval := time.NewTicker(30 * time.Second)
		for _ = range interval.C {
			if data, err := twitterClient.RateLimit(false); err == nil {
				twitterClient.update(data)
				log.Println("timeline", twitterClient.permits[TimelineResource.Key()].Available())

			} else {
				log.Println(err)
			}
		}
	}()

	// make initial call to ratelimit resource to set permits
	if data, err := twitterClient.RateLimit(true); err == nil {
		twitterClient.update(data)
		return twitterClient, nil

	} else {
		return nil, err

	}
}

func (c *TwitterClient) update(data *RateLimit) {
	for key, stats := range data.Resources {
		if _, exists := c.permits[key]; !exists {
			c.permits[key] = concurrent.NewSemaphore(stats.Limit)
		} else {
			c.permits[key].Drain()
		}

		c.permits[key].Release(stats.Remaining)
	}
}

func (c *TwitterClient) Acquire(key TwitterResourceKey) <-chan concurrent.Permit {
	permits, exists := c.permits[key]
	if !exists {
		empty := make(chan concurrent.Permit)
		defer close(empty)
		return empty
	}

	return permits.Acquire(1)
}

func (c *TwitterClient) RateLimit(skipAcquire bool) (*RateLimit, error) {
	params := map[string]string{}
	b, err := getResource(c, RateLimitResource, params, skipAcquire)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewReader(b)
	decoder := json.NewDecoder(buffer)
	ratelimit := TwitterRateLimit{}
	if err := decoder.Decode(&ratelimit); err != nil {
		return nil, err
	}

	return FromTwitterRateLimit(ratelimit), nil
}

func (c *TwitterClient) Posts(userId int) ([]Post, error) {
	params := map[string]string{
		"user_id": fmt.Sprintf("%d", userId),
	}

	b, err := getResource(c, TimelineResource, params, false)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewReader(b)
	decoder := json.NewDecoder(buffer)
	timeline := []TwitterPost{}
	if err := decoder.Decode(&timeline); err != nil {
		return nil, err
	}

	posts := []Post{}
	for _, post := range timeline {
		posts = append(posts, FromTwitterPost(post))
	}

	return posts, nil
}

func getResource(c *TwitterClient, r TwitterResource, q map[string]string, skipAcquire bool) ([]byte, error) {
	acquired := make(chan bool, 1)
	defer close(acquired)

	if skipAcquire {
		acquired <- true
	} else {
		<-c.Acquire(r.Key())
		acquired <- true
	}

	select {
	case <-acquired:
		resp, err := c.Get(r.Url(q).String())
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, Non200Error(resp.StatusCode, r.Url(q).String())
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return b, nil

	case <-time.After(TwitterRequestWait):
		return nil, ErrRequestTimeout
	}

}
