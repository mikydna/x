package twitter

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestTwitterRateLimit_Decode(t *testing.T) {
	b, err := ioutil.ReadFile("./_testdata/twitter-ratelimit.json")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	buffer := bytes.NewReader(b)
	decoder := json.NewDecoder(buffer)

	ratelimit := TwitterRateLimit{}
	if err := decoder.Decode(&ratelimit); err != nil {
		t.Log(err)
	}

	// rate limit context
	expectedAccessToken := "786491-24zE39NUezJ8UTmOGOtLhgyLgCkPyY4dAcx6NA6sDKw"
	if ratelimit.Context.AccessToken != expectedAccessToken {
		t.Errorf("Unexpected access token: %s != %s", expectedAccessToken, ratelimit.Context.AccessToken)
	}

	var expectedApplication string
	if ratelimit.Context.Application != expectedApplication {
		t.Error("Unexpected application: %s != %s", expectedApplication, ratelimit.Context.Application)
	}

	// rate limit resources
	families := []string{}
	for family, _ := range ratelimit.Resources {
		families = append(families, family)
	}

	if len(families) != 4 {
		t.Errorf("Unexpected resource families len: %d != %d", 4, len(families))
	}

	// rate limit resource
	helpPrivacy := ratelimit.Resources["help"]["/help/privacy"]
	if helpPrivacy.Limit != 15 && helpPrivacy.Remaining != 15 && helpPrivacy.Reset.Unix() != 1403602426 {
		t.Errorf("Unexpected resource stat: %v (check test json)", helpPrivacy)
	}

}

func TestTwitterTimeline_Decode(t *testing.T) {
	// t.Skip("TODO")
	b, err := ioutil.ReadFile("./_testdata/twitter-post.json")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	buffer := bytes.NewReader(b)
	decoder := json.NewDecoder(buffer)

	post := TwitterPost{}
	if err := decoder.Decode(&post); err != nil {
		t.Log(err)
	}
}

func TestTwitterTimeline_Decode_PostsFull(t *testing.T) {
	b, err := ioutil.ReadFile("./_testdata/twitter-posts.json")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	buffer := bytes.NewReader(b)
	decoder := json.NewDecoder(buffer)

	posts := []TwitterPost{}
	if err := decoder.Decode(&posts); err != nil {
		t.Log(err)
	}

	if posts[0].Retweet != nil {
		t.Errorf("Unexpected RetweetOf")
	}

	if posts[1].Retweet.Id != 714454338081570817 {
		t.Errorf("Unexpected RetweetOf")
	}

	if posts[1].Retweet.User.Id != 13745182 {
		t.Errorf("Unexpected RetweetOf")
	}
}
