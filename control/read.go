package control

import (
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"
)

func Read(commands chan Command) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		str := req.URL.Query().Get("command")

		var action Action
		switch strings.ToLower(str) {
		case "start":
			action = Start
		case "stop":
			action = Stop
		default:
			action = Start
		}

		fmt.Fprintf(w, "Hello, %q %s", html.EscapeString(req.URL.Path), time.Now())

		commands <- Command{action, "foo"}
	}
}
