package event

import "time"

type Event struct {
	Id       string    `json:"Id"`
	UserId   string    `json:"UserId"`
	Name     string    `json:"Name"`
	DateTime time.Time `json:"DateTime"`
}
