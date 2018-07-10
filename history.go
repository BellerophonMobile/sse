package sse

import (
	"container/list"
)

type history struct {
	items list.List
	limit int
}

func (h *history) append(event *Event) {
	h.items.PushBack(event)

	for h.items.Len() > h.limit {
		h.items.Remove(h.items.Front())
	}
}

func (h *history) find(id string) *list.Element {
	if id == "" {
		return h.items.Front()
	}

	for el := h.items.Front(); el != nil; el = el.Next() {
		if el.Value.(*Event).ID == id {
			return el.Next()
		}
	}

	return h.items.Front()
}
