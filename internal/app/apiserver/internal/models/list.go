package models

type List struct {
	ID    string
	Name  string
	Items []Item
}

type Item struct {
	ID     string
	ListID string
	Name   string
	Done   bool
}
