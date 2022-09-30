package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	watcher := readArgsCreateWatcher()

	watcher.ingestAndProcess()
}

func readArgsCreateWatcher() watcher {
	//interval := flag.Duration("interval", time.Minute*2, "set the interval")
	maxRps := flag.Int("maxRps", 10, "set the number of requests needed to trigger a warning ")
	windowSize := flag.Int("window", 120, "set the size of the window to watch for (in seconds)")
	requestTimeout := flag.Int("requestTimeout", 120, "define the delay after which a request is considered timed out")
	file := flag.String("f", "", "input file")
	flag.Parse()

	//	fmt.Println(*file)
	input := os.Stdin
	var err error
	if *file != "" {
		input, err = os.Open(*file)
		if err != nil {
			log.Fatalf("could not open the file %s", *file)
		}
	}

	csvIngestor, err := initCsvIngestor(input, *requestTimeout)
	if err != nil {
		log.Fatal("could not create the csv ingestor")
	}
	watcher := watcher{
		maxReqPerWindow: *maxRps * *windowSize,
		windowSize:      *windowSize,
		input:           input,
		ingestor:        &csvIngestor,
		ring:            make([]int, *windowSize),
		pathHits:        make(map[string]int),
		alerter:         consoleAlerter{},
	}

	return watcher
}
