package link2

import (
	"net/http"
	"testing"
	"time"
)

import (
	"golang.org/x/net/context"
)

func TestExpand(t *testing.T) {
	bg := context.Background()
	threaded := context.WithValue(bg, "request-id", 1)
	withTimeout, cancel := context.WithTimeout(threaded, 5*time.Second)

	start := time.Now()
	go func() {
		result, err := Expand(withTimeout, http.DefaultClient, "http://google.com")
		defer cancel()

		t.Log(time.Since(start), result, err)
	}()

	select {
	case <-withTimeout.Done():
		t.Log(withTimeout.Value("request-id"))
	case <-time.After(6 * time.Second):
		t.Log(withTimeout.Err())
		t.Error("Unexpected")
	}

}
