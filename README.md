# lgoader
Load Testing Package written in Go

## Quick Start
```go
package main

import (
	"github.com/jcarugati/lgoader"
)

func main() {
	// Create stages
	stage0 := NewStage(&StageCfg{
		Name:              "test0",
		SequenceGenerator: makeSequence,
		Workers:           1,
		Capacity:          1000,
		RequestInterval:   0,
		SequenceInterval:  0,
		LoadTime:          1 * time.Second,
		Results:           nil,
		Done:              nil,
	})

	stage1 := NewStage(&StageCfg{
		Name:              "test1",
		SequenceGenerator: makeSequence,
		Workers:           4,
		Capacity:          1000,
		RequestInterval:   0,
		SequenceInterval:  0,
		LoadTime:          10 * time.Second,
		Results:           nil,
		Done:              nil,
	})

	stage2 := NewStage(&StageCfg{
		Name:              "test2",
		SequenceGenerator: makeSequence,
		Workers:           8,
		Capacity:          1000,
		RequestInterval:   0,
		SequenceInterval:  0,
		LoadTime:          1200 * time.Second,
		Results:           nil,
		Done:              nil,
	})

	// Generate Load Test and assign the stages that'll run
	lt := NewLoadTest(5*time.Second, stage0, stage1, stage2)

	resultChan := lt.Load()

	// Listen to results either using the individual stages of the main LoadTest channel
loop:
	for {
		select {
		case r := <-resultChan:
			// This is how you read results from the main results channel.
			fmt.Println("LT: ", r)
		case rs := <-stage0.Results:
			// This is how you read results from individual stages
			fmt.Println("STAGE 0: ", rs)
		case rs1 := <-stage1.Results:
			fmt.Println("STAGE 1: ", rs1)
		case rs2 := <-stage2.Results:
			fmt.Println("STAGE 2: ", rs2)
		case <-lt.Done:
			// Listen to Done channel to break out of loop
			break loop
		}
	}
}
// makeSequence
// Create a generator function that'll be used in each stage
func makeSequence() RequestSequence {
	const (
		testUrl = "api.test.com/test"
    )
	
	generated := RequestSequence{}

	request, _ := http.NewRequest(http.MethodPost, testUrl, bytes.NewBuffer([]byte{}))

	generated = append(generated, request)

	return generated
}
```
### Request Sequence
Is meant to be an array of `*http.Request`'s. Each stage need an RequestSequence generator function that'll be used to make the requests for load testing.
The idea behind it is to allow the developer to add custom logic used to create a sequence of request, therefore adding flexibility.