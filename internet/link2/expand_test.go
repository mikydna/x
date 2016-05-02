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
	expander := NewExpander(http.DefaultClient, ExtractBasic)

	todo := context.TODO()
	threaded := context.WithValue(todo, "request-id", 1)
	withTimeout, cancel := context.WithTimeout(threaded, 5*time.Second)

	go func() {
		result, err := expander.Expand(withTimeout, "http://google.com")
		defer cancel()

		t.Log(result, err)
	}()

	select {
	case <-withTimeout.Done():
		t.Log(withTimeout.Value("request-id"))
	case <-time.After(6 * time.Second):
		t.Log(withTimeout.Err())
		t.Error("Unexpected")
	}

}
