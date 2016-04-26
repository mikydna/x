package link

import (
	"net/url"
	"sort"
	"strings"
)

func RemoveUTMQueryParams(target *url.URL) *url.URL {
	copy, _ := url.Parse(target.String())

	params := copy.Query()
	for key, _ := range params {
		if strings.HasPrefix(key, "utm_") {
			params.Del(key)
		}
	}

	copy.RawQuery = params.Encode()

	return copy
}

func Normalize(target *url.URL) *url.URL {
	copy, _ := url.Parse(target.String())

	params := copy.Query()
	keys := make([]string, len(params))
	i := 0
	for key, _ := range params {
		keys[i] = key
		i++
	}

	orderedKeys := sort.StringSlice(keys)

	orderedParams := make(url.Values)
	for _, key := range orderedKeys {
		orderedParams.Add(key, params.Get(key))
	}

	if copy.Scheme == "https" {
		copy.Scheme = "http"
	}

	copy.Host = strings.ToLower(copy.Host)
	copy.Path = strings.TrimRight(copy.Path, "/")
	copy.RawQuery = orderedParams.Encode()

	return copy
}
