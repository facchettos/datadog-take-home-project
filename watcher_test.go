package main

import (
	"fmt"
	"testing"
)

func TestUpdateCount(t *testing.T) {
	data1 := make([]int, 120)
	data1[0] = 1000

	data2 := make([]int, 120)
	data2[0] = 1000

	data3 := make([]int, 120)
	data3[0] = 1000

	data4 := make([]int, 120)
	data4[0] = 1000
	data4[60] = 1000

	testData := []struct {
		input    []int
		reqs     int
		previous int
		current  int
		want     int
	}{
		{data1, 1000, 0, 1, 1001},
		{data2, 1000, 0, 0, 1001},
		{data3, 1000, 0, 120, 1},
		{data4, 2000, 61, 125, 1001},
	}

	for k, v := range testData {
		t.Run(fmt.Sprintf("%d", k), func(t *testing.T) {
			watcher := watcher{
				ring:          v.input,
				reqs:          v.reqs,
				previousIndex: v.previous,
			}
			watcher.updateCount(v.current)
			if watcher.reqs != v.want {
				t.Errorf("wanted %d but got %d", v.want, watcher.reqs)
			}
		})
	}
}

func TestShouldAlertHigh(t *testing.T) {

	testData := []struct {
		reqs    int
		max     int
		current bool
		want    bool
	}{
		{10, 1000, false, false},
		{10000, 1000, false, true},
		{10000, 1000, true, false},
	}

	for k, v := range testData {
		t.Run(fmt.Sprintf("%d", k), func(t *testing.T) {
			watcher := watcher{
				maxReqPerWindow: v.max,
				reqs:            v.reqs,
				triggered:       v.current,
			}
			if watcher.shouldAlertHigh() != v.want {
				t.Errorf("wanted %v but got %v", v.want, watcher.shouldAlertHigh())
			}
		})
	}
}

func TestShouldAlertLow(t *testing.T) {

	testData := []struct {
		reqs    int
		max     int
		current bool
		want    bool
	}{
		{10, 1000, false, false},
		{10000, 1000, true, false},
		{1, 1000, true, true},
	}

	for k, v := range testData {
		t.Run(fmt.Sprintf("%d", k), func(t *testing.T) {
			watcher := watcher{
				maxReqPerWindow: v.max,
				reqs:            v.reqs,
				triggered:       v.current,
			}
			if watcher.shouldAlertLow() != v.want {
				t.Errorf("wanted %v but got %v", v.want, watcher.shouldAlertLow())
			}
		})
	}
}

//here I would use gomock instead on a real project, but it's overkill here
type alerterMock struct {
	alertHighCalls int
	alertLowCalls  int
	sendStatsCalls int
}

func (a *alerterMock) alertHigh(hits, max, timestamp, windowSize int) error {
	a.alertHighCalls++
	return nil
}
func (a *alerterMock) alertLow(timestamp int) error {
	a.alertLowCalls++
	return nil
}
func (a *alerterMock) sendStats(string, int) error {
	a.sendStatsCalls++
	return nil
}

func TestWatcherAlerts(t *testing.T) {
	testData := []struct {
		events           int
		wantAlertHigh    int
		wantAlertLow     int
		wantSendStats    int
		eventTimeStamp   int
		currentTimeStamp int
		isTriggered      bool
	}{
		{events: 1000, wantAlertHigh: 1, wantAlertLow: 0, wantSendStats: 0, eventTimeStamp: 1},
		{events: 1, wantAlertHigh: 0, wantAlertLow: 1, wantSendStats: 1, isTriggered: true, currentTimeStamp: 0, eventTimeStamp: 200},
	}

	for k, v := range testData {
		t.Run(fmt.Sprintf("%d", k), func(t *testing.T) {
			alertmock := alerterMock{}
			w := watcher{
				alerter:         &alertmock,
				maxReqPerWindow: 10,
				ring:            make([]int, 120),
				pathHits:        make(map[string]int),
				triggered:       v.isTriggered,
				previousIndex:   v.currentTimeStamp,
			}

			input := make(chan event, 2000)
			input <- event{timestamp: 0}
			for i := 0; i < v.events; i++ {
				input <- event{timestamp: v.eventTimeStamp}
			}
			close(input)
			w.processEvents(input)
			if alertmock.alertHighCalls != v.wantAlertHigh {
				t.Errorf("wanted %d but got %d for alerthigh", v.wantAlertHigh, alertmock.alertHighCalls)
			}
			if alertmock.alertLowCalls != v.wantAlertLow {
				t.Errorf("wanted %d but got %d for alertlow", v.wantAlertLow, alertmock.alertLowCalls)
			}
			if alertmock.sendStatsCalls != v.wantSendStats {
				t.Errorf("wanted %d but got %d for sendstats", v.wantSendStats, alertmock.sendStatsCalls)
			}
		})
	}

}
