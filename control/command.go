package control

import (
	"encoding/json"
)

type Action uint8

const (
	Noop Action = iota
	Start
	Stop
)

type Command struct {
	Action Action `json:"action"`
	Target string `json:"target"`
}

func (c *Command) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}
