package models

import "encoding/json"

type Subscription struct {
	UserID string
	ListID string
	Email  string
}

type ListEvent struct {
	ListID string
	Event  json.RawMessage
}
