package apistats

import (
	"fmt"
	"net/http"
)

type Counter interface {
	Size() int
}

type StatHandler struct {
	DoneCounter    Counter
	InQueueCounter Counter
	BrokenCounter  Counter
}

func NewStatHandler(doneCounter, inQueueCounter, brokenCounter Counter) *StatHandler {
	return &StatHandler{
		DoneCounter:    doneCounter,
		InQueueCounter: inQueueCounter,
		BrokenCounter:  brokenCounter,
	}
}

func (sh *StatHandler) Handler(w http.ResponseWriter, r *http.Request) {
	resp := fmt.Sprintf("\tDone: %d\n \tIn the queue: %d\n \tBlacklisted: %d", sh.DoneCounter.Size(), sh.InQueueCounter.Size(), sh.BrokenCounter.Size())
	w.Write([]byte(resp))
}
