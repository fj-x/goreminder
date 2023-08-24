package reminder

import "github.com/fj-x/goreminder/event"

type Reminder struct {
	Event  event.Event
	ChatID string
}
