package main

import (
	"fmt"
	"time"
)

type alerter interface {
	//alertHigh(hits,max, timestamp, windowSize)
	alertHigh(int, int, int, int) error
	//alertLow(timestamp)
	alertLow(int) error
	//sendstats(path,timestamp)
	sendStats(string, int) error
}

type consoleAlerter struct {
}

func (exp consoleAlerter) alertHigh(hits, max, timestamp, windowSize int) error {
	//TODO get the text from document
	asDate := time.Unix(int64(timestamp), 0)
	_, err := fmt.Printf("%v alert: there has been more than %d request over the last %d seconds\n", asDate, max, windowSize)
	return err
}

func (exp consoleAlerter) alertLow(timestamp int) error {
	asDate := time.Unix(int64(timestamp), 0)
	_, err := fmt.Printf("%v requests have gone back down below the alert treshold\n", asDate)
	return err
}

func (exp consoleAlerter) sendStats(bestHit string, timestamp int) error {
	startAsDate := time.Unix(int64(timestamp), 0)
	_, err := fmt.Printf("%v the most popular was %s between during the last 10 seconds\n", startAsDate.Add(time.Second*10), bestHit)
	return err
}
