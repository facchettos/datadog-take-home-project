package main

import (
	"log"
	"os"
)

type event struct {
	timestamp  int
	statusCode int
	section    string
}

type watcher struct {
	alerter  alerter
	ingestor ingestor

	maxReqPerWindow int
	reqs            int
	previousIndex   int
	windowSize      int

	triggered bool

	//either stdin or an actual file
	input *os.File

	ring []int

	pathHits        map[string]int
	pathMaxHits     string
	lastStatsUpdate int
}

func (w *watcher) ingestAndProcess() {
	//using a buffered channel to avoid unnecessary blocking
	//size may be subject to change
	ch := make(chan event, 100)
	go w.ingestor.readEvents(ch)
	w.processEvents(ch)
}

func (w *watcher) processEvent(e event) error {
	w.updateCount(e.timestamp)
	w.previousIndex = e.timestamp

	if e.timestamp > w.lastStatsUpdate+10 {
		w.alerter.sendStats(w.pathMaxHits, w.lastStatsUpdate)
		w.resetStats(e.timestamp)
	}
	w.updateStats(e)

	if w.shouldAlertHigh() {
		w.triggered = true
		return w.alerter.alertHigh(w.reqs, w.maxReqPerWindow, e.timestamp, w.windowSize)
	}
	if w.shouldAlertLow() {
		w.triggered = false
		return w.alerter.alertLow(e.timestamp)
	}
	return nil
}

func (w *watcher) processEvents(input chan event) {
	//we need the first event to instantiate the timestamp
	shouldInit := true
	var e event
	for e := range input {
		if shouldInit {
			w.lastStatsUpdate = e.timestamp
			shouldInit = false
		}

		err := w.processEvent(e)
		if err != nil {
			log.Printf("error while alerting: %s", err.Error())
		}
	}
	//if not triggered by the last event
	if e.timestamp > w.lastStatsUpdate {
		w.alerter.sendStats(w.pathMaxHits, w.lastStatsUpdate)
	}
}

func (w *watcher) updateCount(current int) {
	if current-w.previousIndex >= len(w.ring) {
		//we can simplify if it's been more than the window
		//since last event
		for i := 0; i < len(w.ring); i++ {
			w.ring[i] = 0
		}
		w.reqs = 0
	} else {
		//this implements a sliding window
		for i := w.previousIndex + 1; i <= current; i++ {
			//removes what was added between now-<thewindow> and
			//lastevent if it was longer than <thewindow> since it
			//happened
			w.reqs -= w.ring[i%len(w.ring)]
			w.ring[i%len(w.ring)] = 0
		}
	}
	w.reqs++
	w.ring[current%len(w.ring)]++
	w.previousIndex = current
}

func (w *watcher) shouldAlertHigh() bool {
	return !w.triggered && w.reqs > w.maxReqPerWindow
}

func (w *watcher) shouldAlertLow() bool {
	return w.triggered && w.reqs <= w.maxReqPerWindow
}

func (w *watcher) updateStats(e event) {
	w.pathHits[e.section]++
	if w.pathHits[e.section] > w.pathHits[w.pathMaxHits] {
		w.pathMaxHits = e.section
	}
}

func (w *watcher) resetStats(nextEvent int) {
	for nextEvent > w.lastStatsUpdate+10 {
		w.lastStatsUpdate += 10
	}
	w.pathMaxHits = ""
	w.pathHits = make(map[string]int)
}
