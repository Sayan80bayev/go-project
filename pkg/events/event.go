package events

import "encoding/json"

type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}
