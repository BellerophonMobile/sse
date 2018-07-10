package sse

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReader_good(t *testing.T) {
	suite := []struct {
		name      string
		input     string
		events    []*Event
		lastID    string
		retryTime time.Duration
	}{
		{
			name:  "single event",
			input: "data: Hello\n\n",
			events: []*Event{
				{Type: "message", Data: "Hello"},
			},
		},
		{
			name:  "single event without space",
			input: "data:Hello\n\n",
			events: []*Event{
				{Type: "message", Data: "Hello"},
			},
		},
		{
			name: "event with type",
			input: `data: World
event: foobar

`,
			events: []*Event{
				{Type: "foobar", Data: "World"},
			},
		},
		{
			name: "event with type without space",
			input: `data:World
event:foobar

`,
			events: []*Event{
				{Type: "foobar", Data: "World"},
			},
		},
		{
			name: "single multiline event",
			input: `data: YHOO
data: +2
data: 10

`,
			events: []*Event{
				{Type: "message", Data: "YHOO\n+2\n10"},
			},
		},
		{
			name: "events with comments",
			input: `: test stream

data: first event
id: 1

data:second event
id

data: third event

`,
			events: []*Event{
				{Type: "message", ID: "1", Data: "first event"},
				{Type: "message", Data: "second event"},
				{Type: "message", Data: "third event"},
			},
		},
		{
			name: "events with multiple IDs",
			input: `data: first event
id: 1

data:second event
id: 2

data: third event
id: 3

`,
			events: []*Event{
				{Type: "message", ID: "1", Data: "first event"},
				{Type: "message", ID: "2", Data: "second event"},
				{Type: "message", ID: "3", Data: "third event"},
			},
			lastID: "3",
		},
		{
			name: "events with multiple IDs and types",
			input: `data: first event
id: 1
event: first

id:2
data:second event
event: second

event:third
data: third event
id: 3

`,
			events: []*Event{
				{Type: "first", ID: "1", Data: "first event"},
				{Type: "second", ID: "2", Data: "second event"},
				{Type: "third", ID: "3", Data: "third event"},
			},
			lastID: "3",
		},
		{
			name: "blank events",
			input: `data

data
data

data:
`,
			events: []*Event{
				{Type: "message", Data: ""},
				{Type: "message", Data: "\n"},
			},
		},
		{
			name: "ignore space after colon",
			input: `data:test

data: test

`,
			events: []*Event{
				{Type: "message", Data: "test"},
				{Type: "message", Data: "test"},
			},
		},
		{
			name: "data with separate retry",
			input: `data: foo

retry: 10000

data: bar

`,
			events: []*Event{
				{Type: "message", Data: "foo"},
				{Type: "message", Data: "bar"},
			},
			retryTime: 10 * time.Second,
		},
		{
			name: "data with separate retry without spaces",
			input: `data:foo

retry:10000

data:bar

`,
			events: []*Event{
				{Type: "message", Data: "foo"},
				{Type: "message", Data: "bar"},
			},
			retryTime: 10 * time.Second,
		},
		{
			name: "data with joined retry",
			input: `data: foo
retry: 10000

`,
			events: []*Event{
				{Type: "message", Data: "foo"},
			},
			retryTime: 10 * time.Second,
		},
		{
			name: "data with joined retry without spaces",
			input: `data:foo
retry:10000

`,
			events: []*Event{
				{Type: "message", Data: "foo"},
			},
			retryTime: 10 * time.Second,
		},
	}

	for _, trial := range suite {
		t.Run(trial.name, func(t *testing.T) {
			r := NewReader(strings.NewReader(trial.input))

			for i, expected := range trial.events {
				actual, err := r.NextEvent()
				if err != nil {
					t.Errorf("failed to read event: %v", err)
					return
				}

				assert.Equal(t, expected, actual, "Event %d not equal", i)
			}

			assert.Equal(t, trial.retryTime, r.RetryTime)
			assert.Equal(t, trial.lastID, r.LastEventID)

			// Make sure there are no errornous events remaining.
			evt, err := r.NextEvent()
			if evt != nil {
				t.Errorf("unexpected event: %v", evt)
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
