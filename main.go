package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/rishabht08/template-miner/pkg/miner"
)

var methods = []string{"GET", "POST", "PUT", "DELETE"}
var statuses = []int{200, 201, 301, 400, 401, 403, 404, 500, 502}
var paths = []string{"/login", "/home", "/user", "/cart", "/api/data", "/products", "/logout"}
var userAgents = []string{
	"Mozilla/5.0", "curl/7.68.0", "PostmanRuntime/7.29.0", "Go-http-client/1.1",
}

func randomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255))
}

func randomTimestamp() string {
	return time.Now().Add(time.Duration(-rand.Intn(10000)) * time.Second).Format("02/Jan/2006:15:04:05 -0700")
}

func generateLog() string {
	return fmt.Sprintf(`%s - - [%s] "%s %s HTTP/1.1" %d %d "%s" "%s"`,
		randomIP(),
		randomTimestamp(),
		methods[rand.Intn(len(methods))],
		paths[rand.Intn(len(paths))],
		statuses[rand.Intn(len(statuses))],
		rand.Intn(5000)+100,
		"http://example.com",
		userAgents[rand.Intn(len(userAgents))],
	)
}

func main() {
	numLogs := 150 // generate between 100–200
	logs := make([]string, 0)
	for i := 0; i < numLogs; i++ {
		logs = append(logs, generateLog())
	}

	patternMiner, err := miner.NewMiner(nil, "rishab-key", miner.Config{
		RegexMap: map[string]string{
			`(\b\d{1,3}(?:\.\d{1,3}){3}:\d+\s\[\d{2}/[A-Za-z]{3}/\d{4}:\d{2}:\d{2}:\d{2}\.\d{3}\])`: "Ts[T]",
			`(\d{4}-\d{2}-\d{2})`: "date",
			`(\d{2}:\d{2}:\d{2})`: "time",
		},
	})
	if err != nil {
		panic(err)
	}
	defer patternMiner.Close()

	results := patternMiner.Train(logs)
	for _, res := range results {
		fmt.Printf("\nOriginal:    %s\nTemplate:    %s\nTemplate ID: %s\nParams:      %s\nTokens:    %s\n",
			res.OriginalLog, res.Template, res.TemplateID, res.Parameters, res.Tokens)
	}

}
