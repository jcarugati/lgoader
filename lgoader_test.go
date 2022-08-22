package lgoader

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jcarugati/lgoader/helpers"
	"github.com/stretchr/testify/assert"
)

const (
	url = "/test"
)

var srv = helpers.HttpMock(url, http.StatusOK, "")

func TestNewStage(t *testing.T) {
	a := assert.New(t)

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

loop:
	for {
		select {
		case r := <-resultChan:
			a.NotNil(r)
			fmt.Println("LT: ", r)
		case rs := <-stage0.Results:
			a.NotNil(rs)

			fmt.Println("STAGE 0: ", rs)
		case rs1 := <-stage1.Results:
			a.NotNil(rs1)

			fmt.Println("STAGE 1: ", rs1)
		case rs2 := <-stage2.Results:
			a.NotNil(rs2)

			fmt.Println("STAGE 2: ", rs2)
		case <-lt.Done:
			break loop
		}
	}
}

// Generate the sequence generator Logic
func makeSequence() RequestSequence {
	generated := RequestSequence{}

	domain := srv.URL
	path := url

	request, _ := http.NewRequest(http.MethodPost, domain+path, bytes.NewBuffer([]byte{}))

	generated = append(generated, request)

	return generated
}
