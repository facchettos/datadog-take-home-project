package main

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type ingestor interface {
	readEvents(chan event)
}

type csvIngestor struct {
	fieldsByName     map[string]int
	csvReader        *csv.Reader
	input            io.ReadCloser
	requestTimeout   int
	currentTimestamp int
	overdueRequests  map[int][]event
}

func (csvIngestor csvIngestor) destroy() {
	if csvIngestor.input != os.Stdin {
		csvIngestor.input.Close()
	}
}

func initCsvIngestor(input io.ReadCloser, requestTimeout int) (csvIngestor, error) {
	ingestor := csvIngestor{}
	ingestor.input = input
	csvReader := csv.NewReader(input)
	if csvReader == nil {
		return ingestor, errors.New("could not create the csv reader")
	}
	ingestor.csvReader = csvReader

	fields, err := ingestor.csvReader.Read()
	if err != nil {
		return ingestor, err
	}
	indexMap := make(map[string]int)
	for k, v := range fields {
		indexMap[v] = k
	}
	ingestor.fieldsByName = indexMap

	ingestor.requestTimeout = requestTimeout
	ingestor.overdueRequests = make(map[int][]event, requestTimeout)

	return ingestor, nil
}

func (ingestor csvIngestor) readEvent() (event, error) {
	fields, err := ingestor.csvReader.Read()
	if err != nil {
		return event{}, err
	}
	timestamp, err := strconv.Atoi(fields[ingestor.fieldsByName["date"]])
	if err != nil {
		return event{}, errors.New("corrupted timestamp")
	}

	statusCode, err := strconv.Atoi(fields[ingestor.fieldsByName["status"]])
	if err != nil {
		return event{}, errors.New("corrupted status code")
	}

	section, err := extractSection(fields[ingestor.fieldsByName["request"]])
	if err != nil {
		return event{}, errors.New("corrupted request field")
	}
	result := event{
		timestamp:  timestamp,
		statusCode: statusCode,
		section:    section,
	}

	return result, nil
}

func (ingestor *csvIngestor) readEvents(output chan event) {
	defer close(output)
	defer ingestor.destroy()
	shouldInit := true
	for event, err := ingestor.readEvent(); err != io.EOF; event, err = ingestor.readEvent() {
		if shouldInit {
			ingestor.currentTimestamp = event.timestamp
			shouldInit = false
		}
		if err != nil {
			log.Printf("error while reading the csv file %s", err)
			continue
		}
		ingestor.sendEvent(event, output)
	}
	//sending all the remaining ones
	ingestor.sendAllEvents(output)
}

func (ingestor *csvIngestor) sendEvent(event event, output chan event) {
	if event.timestamp == ingestor.currentTimestamp {
		output <- event
	} else if event.timestamp < ingestor.currentTimestamp {
		//discard, should not happen if timeout is set properly
		return
	} else {
		//send to the channel all event from all timestamps that have expired
		for ingestor.currentTimestamp < event.timestamp-ingestor.requestTimeout {
			for _, v := range ingestor.overdueRequests[ingestor.currentTimestamp] {
				output <- v
			}
			delete(ingestor.overdueRequests, ingestor.currentTimestamp)
			ingestor.currentTimestamp++
		}
		//add to the list of request to process later
		ingestor.overdueRequests[event.timestamp] = append(ingestor.overdueRequests[event.timestamp], event)
	}
}

func (ingestor *csvIngestor) sendAllEvents(output chan event) {
	//we need to sort because because ranging over a map is not deterministic nor ordered
	//we could maybe use a slice instead but given the small size of the sort we
	//need to do, I decided that it was ok using the map for simplicity instead of
	//using lots of modulos(moduli?)
	indexes := make([]int, 0, len(ingestor.overdueRequests))
	for k := range ingestor.overdueRequests {
		indexes = append(indexes, k)
	}
	sort.Ints(indexes)
	for _, index := range indexes {
		for _, v := range ingestor.overdueRequests[index] {
			output <- v
		}
	}
}

//extracts the sections for the field that we have in the csv file
func extractSection(field string) (string, error) {
	words := strings.Split(field, " ")
	if len(words) < 3 {
		return "", errors.New("field is not valid")
	}
	path := words[1]
	indexSecondSlash := strings.Index(path[1:], "/") + 1
	section := path
	if indexSecondSlash <= 0 {
		return section, nil
	}
	return section[:indexSecondSlash], nil
}
