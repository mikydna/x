package link

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

type Result struct {
	StatusCode   int
	ResponseTime time.Duration
	ResolvedURL  *url.URL
	Content      Content
}

func (r *Result) MarshalStringMap() (map[string]string, error) {
	strmap := make(map[string]string)
	strmap["statusCode"] = fmt.Sprintf("%d", r.StatusCode)
	strmap["responseTime"] = fmt.Sprintf("%d", r.ResponseTime.Nanoseconds())
	strmap["resolvedURL"] = r.ResolvedURL.String()

	for key, val := range r.Content {
		ckey := fmt.Sprintf("c_%d", key)
		strmap[ckey] = val
	}

	return strmap, nil
}

func (r *Result) UnmarshalStringMap(strmap map[string]string) error {
	var result Result

	if str, exists := strmap["statusCode"]; exists {
		statusCode, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			return err
		}

		result.StatusCode = int(statusCode)
	}

	if str, exists := strmap["responseTime"]; exists {
		responseTime, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}

		result.ResponseTime = time.Duration(responseTime)
	}

	if str, exists := strmap["resolvedURL"]; exists {
		resolvedURL, err := url.Parse(str)
		if err != nil {
			return err
		}

		result.ResolvedURL = resolvedURL
	}

	result.Content = make(Content)
	for _, key := range []ContentType{Title, Description, FavIcon} {
		ckey := fmt.Sprintf("c_%d", key)
		if str, exists := strmap[ckey]; exists {
			result.Content[key] = str
		}
	}

	*r = result

	return nil
}
