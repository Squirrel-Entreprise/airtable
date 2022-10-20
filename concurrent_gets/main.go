// Exemplify making concurrent requests of the Airtable API.
//
// I've pushed 50 requests through in just 7 seconds, and the API didn't rate-limit
// me.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"
)

// Make this many concurrent API requests
var numReqs = 50
var app1URL = "https://api.airtable.com/v0/app3n84SwUnqPsvgC/Table_1?view=Just_1"

// Get API key secret out of a JSON file
type jsonConfig struct {
	APIKey string `json:"APIKey"`
}

var config = jsonConfig{}

func init() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("usage: go run main.go CONFIG_JSON")
		os.Exit(1)
	}
	dat, err := os.ReadFile(args[1])
	checkErr(err)
	json.Unmarshal(dat, &config)
}

type requestResult struct {
	start time.Time     // request start time
	code  int           // response status code
	dur   time.Duration // how long the response took
}

func (r requestResult) String() string {
	return fmt.Sprintf("started request %s, took %5.2fs, returned %d",
		r.start.Format("2006-01-02T15:04:05.000000"),
		r.dur.Seconds(),
		r.code)
}

func main() {
	// Get back individual requestResults from the concurrently running threads
	chResult := make(chan requestResult, numReqs)
	// Accumulate the results
	results := make([]requestResult, numReqs)

	// Start concurrent requests and wait for all responses
	beg := time.Now()
	for i := 1; i < numReqs+1; i++ {
		go func() {
			makeRequest(app1URL, chResult)
		}()
	}
	for i := 0; i < numReqs; i++ {
		results[i] = <-chResult
	}
	elapsed := time.Since(beg)

	// Sort by when the request was actually made
	sort.Slice(results, func(i, j int) bool {
		return results[i].start.Before(results[j].start)
	})

	bads := 0
	for _, result := range results {
		fmt.Println(result)
		if result.code != 200 {
			bads += 1
		}
	}

	fmt.Printf("Got %d bad responses out of %d requests in %.2fs\n", bads, numReqs, elapsed.Seconds())
}

func makeRequest(url string, codeResp chan<- requestResult) {
	req, err := http.NewRequest("GET", url, nil)
	checkErr(err)
	req.Header.Add("Authorization", "Bearer "+config.APIKey)

	client := &http.Client{}

	beg := time.Now()
	resp, err := client.Do(req)
	end := time.Now()

	checkErr(err)

	codeResp <- requestResult{start: beg, code: resp.StatusCode, dur: end.Sub(beg)}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

/* Submitting 50 requests "at once"

All 50 are sent in the span of one second, but each subsequent request takes progressively longer to respond:

started request 42 at Jul 21 11:23:27.565001, ended in 1.50s
started request 17 at Jul 21 11:23:27.565066, ended in 1.53s
started request 13 at Jul 21 11:23:27.565545, ended in 1.55s
started request  6 at Jul 21 11:23:27.565309, ended in 1.59s
started request 36 at Jul 21 11:23:27.564799, ended in 1.61s
started request 14 at Jul 21 11:23:27.565381, ended in 1.62s
started request 27 at Jul 21 11:23:27.564980, ended in 1.64s
started request 39 at Jul 21 11:23:27.564951, ended in 1.65s
started request 34 at Jul 21 11:23:27.564746, ended in 1.67s
started request 18 at Jul 21 11:23:27.565072, ended in 1.68s
started request 45 at Jul 21 11:23:27.565177, ended in 1.70s
started request 22 at Jul 21 11:23:27.565197, ended in 1.71s
started request 32 at Jul 21 11:23:27.564678, ended in 1.73s
started request 49 at Jul 21 11:23:27.565287, ended in 1.75s
started request  9 at Jul 21 11:23:27.565290, ended in 1.76s
started request 38 at Jul 21 11:23:27.564612, ended in 1.78s
started request 29 at Jul 21 11:23:27.564628, ended in 1.79s
started request 20 at Jul 21 11:23:27.565144, ended in 1.81s
started request  4 at Jul 21 11:23:27.564470, ended in 1.83s
started request 48 at Jul 21 11:23:27.565283, ended in 1.84s
started request  3 at Jul 21 11:23:27.564935, ended in 1.86s
started request  1 at Jul 21 11:23:27.564494, ended in 1.87s
started request 47 at Jul 21 11:23:27.565156, ended in 1.89s
started request 46 at Jul 21 11:23:27.565184, ended in 3.74s
started request 33 at Jul 21 11:23:27.564731, ended in 3.76s
started request 15 at Jul 21 11:23:27.565349, ended in 3.78s
started request 30 at Jul 21 11:23:27.564635, ended in 3.80s
started request 23 at Jul 21 11:23:27.565204, ended in 3.81s
started request  7 at Jul 21 11:23:27.565303, ended in 3.83s
started request 31 at Jul 21 11:23:27.564641, ended in 3.85s
started request 11 at Jul 21 11:23:27.565303, ended in 3.88s
started request 12 at Jul 21 11:23:27.565291, ended in 3.88s
started request 43 at Jul 21 11:23:27.565046, ended in 3.90s
started request 44 at Jul 21 11:23:27.564943, ended in 3.96s
started request 41 at Jul 21 11:23:27.564993, ended in 3.96s
started request 40 at Jul 21 11:23:27.564982, ended in 3.98s
started request 21 at Jul 21 11:23:27.565054, ended in 3.99s
started request 35 at Jul 21 11:23:27.564753, ended in 4.02s
started request 19 at Jul 21 11:23:27.565085, ended in 4.04s
started request  2 at Jul 21 11:23:27.564474, ended in 4.05s
started request 24 at Jul 21 11:23:27.565191, ended in 4.07s
started request  8 at Jul 21 11:23:27.565304, ended in 4.08s
started request 26 at Jul 21 11:23:27.565276, ended in 4.10s
started request 25 at Jul 21 11:23:27.565257, ended in 4.11s
started request 10 at Jul 21 11:23:27.565295, ended in 4.13s
started request 28 at Jul 21 11:23:27.564622, ended in 7.16s
started request 16 at Jul 21 11:23:27.565060, ended in 7.18s
started request 37 at Jul 21 11:23:27.564810, ended in 7.20s
started request  5 at Jul 21 11:23:27.565298, ended in 7.22s
started request 50 at Jul 21 11:23:27.564445, ended in 7.35s
Got 0 bad responses out of 50 requests in 7.3542835s
*/
