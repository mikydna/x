package twitter

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestPost_TwitterRateLimitConv(t *testing.T) {
	b, _ := ioutil.ReadFile("./_testdata/twitter-ratelimit.json")
	buffer := bytes.NewReader(b)
	decoder := json.NewDecoder(buffer)

	twitterRateLimit := TwitterRateLimit{}
	if err := decoder.Decode(&twitterRateLimit); err != nil {
		t.Error(err)
	}

	rateLimit := FromTwitterRateLimit(twitterRateLimit)
	userProfileBannerKey := TwitterResourceKey{"users", "/users/profile_banner"}
	userProfileBanner := rateLimit.Resources[userProfileBannerKey]

	if userProfileBanner.Limit != 180 {
		t.Errorf("Unexpected resource value %d != %d", 180, userProfileBanner.Limit)
	}

	if userProfileBanner.Remaining != 180 {
		t.Errorf("Unexpected resource value %d != %d", 180, userProfileBanner.Remaining)
	}

	if userProfileBanner.Reset.Unix() != 1403602426 {
		t.Errorf("Unexpected resource value %d != %d", 1403602426, userProfileBanner.Reset.Unix())
	}

}
