package main

import (
	"os"
	"testing"
)

func TestCsvInit(t *testing.T) {
	file, err := os.OpenFile("./test_data/simple.csv", os.O_RDONLY, 0666)
	if err != nil {
		t.Fail()
	}
	csvIng, err := initCsvIngestor(file, 120)
	defer csvIng.destroy()

	if csvIng.fieldsByName["date"] != 3 {
		t.Errorf("wanted date to be 3 but is %d", csvIng.fieldsByName["date"])
	}
	if csvIng.fieldsByName["request"] != 4 {
		t.Errorf("wanted request to be 4 but is %d", csvIng.fieldsByName["date"])
	}
	if csvIng.fieldsByName["status"] != 5 {
		t.Errorf("wanted status to be 5 but is %d", csvIng.fieldsByName["date"])
	}
	e, err := csvIng.readEvent()
	expected := event{timestamp: 1549574332, section: "/api", statusCode: 200}
	if e != expected {
		t.Error("event is not as expected")
	}
}

func TestExtractPath(t *testing.T) {
	testData := []struct {
		input string
		want  string
	}{
		{"GET / HTTP/1.0", "/"},
		{"GET /datadog HTTP/1.0", "/datadog"},
		{"GET //// HTTP/1.0", "/"},
		{"GET /data/dog HTTP/1.0", "/data"},
		{"/data/dog HTTP/1.0", ""},
	}

	for _, v := range testData {
		t.Run(v.input, func(t *testing.T) {
			if result, _ := extractSection(v.input); result != v.want {
				t.Errorf("wanted %s but got %s", v.want, result)
			}
		})
	}
}

func TestSendEvents(t *testing.T) {
	testData := []struct {
		name  string
		input []event
		want  []event
	}{
		{
			name: "send",
			input: []event{
				{timestamp: 0, statusCode: 200, section: "/api"},
				{timestamp: 0, statusCode: 200, section: "/api"},
				{timestamp: 0, statusCode: 200, section: "/api"},
			},
			want: []event{
				{timestamp: 0, statusCode: 200, section: "/api"},
				{timestamp: 0, statusCode: 200, section: "/api"},
				{timestamp: 0, statusCode: 200, section: "/api"},
			},
		},
		{
			name: "reorder",
			input: []event{
				{timestamp: 0, statusCode: 200, section: "/api"},
				{timestamp: 12, statusCode: 200, section: "/api"},
				{timestamp: 0, statusCode: 200, section: "/api"},
			},
			want: []event{
				{timestamp: 0, statusCode: 200, section: "/api"},
				{timestamp: 0, statusCode: 200, section: "/api"},
				{timestamp: 12, statusCode: 200, section: "/api"},
			},
		},
		{
			name: "reorder send expired",
			input: []event{
				{timestamp: 1, statusCode: 200, section: "/api"},
				{timestamp: 0, statusCode: 200, section: "/api"},
				{timestamp: 130, statusCode: 200, section: "/api"},
				{timestamp: 129, statusCode: 200, section: "/api"},
			},
			want: []event{
				{timestamp: 0, statusCode: 200, section: "/api"},
				{timestamp: 1, statusCode: 200, section: "/api"},
				{timestamp: 129, statusCode: 200, section: "/api"},
				{timestamp: 130, statusCode: 200, section: "/api"},
			},
		},
	}

	for _, v := range testData {
		t.Run(v.name, func(t *testing.T) {
			//buffered so we don't have to use routines
			outChan := make(chan event, 100)
			defer close(outChan)
			file, err := os.OpenFile("./test_data/simple.csv", os.O_RDONLY, 0666)
			if err != nil {
				t.Error(err)
			}
			ingestor, err := initCsvIngestor(file, 120)
			ingestor.currentTimestamp = 0
			if err != nil {
				t.Error(err)
			}
			for _, v := range v.input {
				ingestor.sendEvent(v, outChan)
			}
			ingestor.sendAllEvents(outChan)
			for i := 0; i < len(v.want); i++ {
				res := <-outChan
				if res != v.want[i] {
					t.Errorf("wanted %v but got %v", v.want[i], res)
				}
			}
		})
	}
}
