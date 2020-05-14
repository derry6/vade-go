package source

import (
    "encoding/json"
)

type Action string

const (
    Created Action = "Created"
    Updated Action = "Updated"
    Deleted Action = "Deleted"
)

type Event struct {
    Action    Action      `json:"action,omitempty"`
    Source    string      `json:"source,omitempty"`
    Path      string      `json:"path,omitempty"`
    Key       string      `json:"key,omitempty"`
    ValueFrom interface{} `json:"vFrom,omitempty"`
    ValueTo   interface{} `json:"vTo,omitempty"`
}

func (e *Event) String() string {
    if e == nil {
        return "{}"
    }
    d, _ := json.Marshal(e)
    return string(d)
}

func NewEvent(action Action, key string) *Event {
    return &Event{Action: action, Key: key}
}
