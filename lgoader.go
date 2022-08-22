package lgoader

import (
	"fmt"
	"net/http"
	"time"

	"github.com/alitto/pond"
	"github.com/go-resty/resty/v2"
)

type RequestSequence []*http.Request

type SequenceGenerator func() RequestSequence

type Stage struct {
	name                                        string
	generator                                   SequenceGenerator
	Results                                     chan LoadResult
	Done                                        chan bool
	loadTime, requestInterval, sequenceInterval time.Duration
	pool                                        *pond.WorkerPool
	signal                                      chan struct{}
	sequencesChan                               chan RequestSequence
	timer                                       *time.Timer
}

type StageCfg struct {
	Name                                        string
	SequenceGenerator                           SequenceGenerator
	Workers                                     int
	Capacity                                    int
	RequestInterval, SequenceInterval, LoadTime time.Duration
	Results                                     chan LoadResult
	Done                                        chan bool
}

func NewStage(cfg *StageCfg) *Stage {
	var (
		capacity int
		loadTime time.Duration
	)

	if cfg.Capacity == 0 {
		capacity = 1
	} else {
		capacity = cfg.Capacity
	}

	if cfg.LoadTime == 0 {
		loadTime = (cfg.RequestInterval + cfg.SequenceInterval) * 100
	} else {
		loadTime = cfg.LoadTime
	}

	if loadTime < cfg.SequenceInterval || loadTime < cfg.RequestInterval {
		panic("intervals cannot be larger than total load time")
	}

	stage := &Stage{
		name:             cfg.Name,
		generator:        cfg.SequenceGenerator,
		pool:             pond.New(cfg.Workers, capacity),
		Results:          cfg.Results,
		Done:             make(chan bool),
		requestInterval:  cfg.RequestInterval,
		sequenceInterval: cfg.SequenceInterval,
		signal:           make(chan struct{}),
		sequencesChan:    make(chan RequestSequence, capacity),
		loadTime:         loadTime,
	}

	if cfg.Results == nil {
		stage.Results = make(chan LoadResult)
	}

	if cfg.Done == nil {
		stage.Done = make(chan bool)
	}

	return stage
}

func (s *Stage) GetResults() chan LoadResult {
	return s.Results
}

func (s *Stage) Ready() <-chan bool {
	return s.Done
}

type LoadResult struct {
	IsErr bool
	StatusCode,
	DNSLookup,
	ConnectionTime,
	TPCConnectionTime,
	TLSHandshake,
	ServerTime,
	ResponseTime,
	TotalTime,
	IsConnectionReused,
	StageName string
}

type LoadTest struct {
	client        *resty.Client
	Stages        []*Stage
	StageInterval time.Duration
	Results       chan LoadResult
	Done          chan bool
}

func NewLoadTest(interval time.Duration, stages ...*Stage) *LoadTest {
	return &LoadTest{
		client:        resty.New(),
		Stages:        stages,
		StageInterval: interval,
		Results:       make(chan LoadResult),
		Done:          make(chan bool),
	}
}

func (lt *LoadTest) Load() chan LoadResult {
	go lt.load()
	return lt.Results
}

func (lt *LoadTest) load() {
	for _, stage := range lt.Stages {
		go stage.genReq()

		timer := time.NewTimer(stage.loadTime)

	stageLoop:
		for {
			select {
			case rs, ok := <-stage.sequencesChan:
				if ok {
					stage.pool.Submit(func() {
						lt.runSequence(rs, stage)
					})
				}
			case <-timer.C:
				close(stage.signal)
				close(stage.Done)
				break stageLoop
			}
		}
		stage.pool.Stop()
	}
	for _, stage := range lt.Stages {
		close(stage.Results)
	}
	close(lt.Results)
	close(lt.Done)
}

func (lt *LoadTest) runSequence(rs RequestSequence, stg *Stage) {
	for _, r := range rs {
		res := lt.doRequest(r)
		res.StageName = stg.name

		lt.sendResult(stg, *res)
		time.Sleep(stg.requestInterval)
	}
}

func (lt *LoadTest) doRequest(r *http.Request) *LoadResult {
	rr := lt.client.R().EnableTrace()

	migrateHttpReqToResty(r, rr)

	resp, err := rr.Send()
	if err != nil {
		return &LoadResult{IsErr: true}
	}

	ti := resp.Request.TraceInfo()

	return &LoadResult{
		IsErr:              false,
		StatusCode:         fmt.Sprint(resp.StatusCode()),
		DNSLookup:          ti.DNSLookup.String(),
		ConnectionTime:     ti.ConnTime.String(),
		TPCConnectionTime:  ti.TCPConnTime.String(),
		TLSHandshake:       ti.TLSHandshake.String(),
		ServerTime:         ti.ServerTime.String(),
		ResponseTime:       ti.ResponseTime.String(),
		TotalTime:          ti.TotalTime.String(),
		IsConnectionReused: fmt.Sprint(ti.IsConnReused),
	}

}

func (lt *LoadTest) sendResult(stg *Stage, lr LoadResult) {
	select {
	case stg.Results <- lr:
	case lt.Results <- lr:
	default:
	}
}

func (s *Stage) genReq() {
loop:
	for {
		select {
		case <-s.signal:
			break loop
		default:
			time.Sleep(s.sequenceInterval)
			s.sequencesChan <- s.generator()
		}
	}
	close(s.sequencesChan)
}

func migrateHttpReqToResty(strd *http.Request, resty *resty.Request) {
	resty.URL = strd.URL.String()
	resty.Method = strd.Method
	resty.QueryParam = strd.URL.Query()
	resty.FormData = strd.Form
	resty.Header = strd.Header
	resty.Body = strd.Body
	resty.RawRequest = strd
	resty.Cookies = strd.Cookies()
}
