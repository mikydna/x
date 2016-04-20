package twitter

import (
	"os"
	"testing"
)

var (
	TwitterConsumerKey    = os.Getenv("TwitterConsumerKey")
	TwitterConsumerSecret = os.Getenv("TwitterConsumerSecret")
)

func TestAPIClient_RateLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Invokes outbound http calls")
	}

	if TwitterConsumerKey == "" {
		t.Errorf("Env var required: %s", "TwitterConsumerKey")
		t.FailNow()
	}

	if TwitterConsumerSecret == "" {
		t.Errorf("Env var required: %s", "TwitterConsumerSecret")
		t.FailNow()
	}

	client, _ := NewClient(TwitterConsumerKey, TwitterConsumerSecret)
	if ratelimit, err := client.RateLimit(true); err != nil {
		t.Error(err)
	} else {
		if resources := ratelimit.Resources; len(resources) != 45 {
			t.Errorf("Unexpected num of resources %d != %d", 45, len(resources))
		}
	}
}

func TestAPIClient_Timeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Invokes outbound http calls")
	}

	if TwitterConsumerKey == "" {
		t.Errorf("Env var required: %s", "TwitterConsumerKey")
		t.FailNow()
	}

	if TwitterConsumerSecret == "" {
		t.Errorf("Env var required: %s", "TwitterConsumerSecret")
		t.FailNow()
	}

	client, err := NewClient(TwitterConsumerKey, TwitterConsumerSecret)
	if err != nil {
		t.Error(err)
	}

	if timeline, err := client.Posts(14305066); err != nil {
		t.Error(err)
	} else {
		if timeline[0].UserId != 14305066 {
			t.Errorf("Unexpected user in response: %d != %d", 14305066, timeline[0].UserId)
		}
	}

}
