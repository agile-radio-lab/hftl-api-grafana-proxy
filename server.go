package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"bitbucket.org/hftloai/hftlapiconnector/models"

	"bitbucket.org/hftloai/hftlapiconnector"
)

type server struct {
	sync.RWMutex

	ctx    context.Context
	i      int
	events []AnnotationResponse

	apiConn *hftlapiconnector.APIClient
}

func newServer() *server {
	return &server{ctx: context.Background()}
}

func (s *server) seed(max int) {
	s.RLock()
	defer s.RUnlock()

	expansion := 20 * time.Minute
	n := time.Now().Add(-(expansion * time.Duration(max)))
	for i := 0; i < max; i++ {
		t := n.Add(time.Duration(i+1) * expansion)
		s.events = append(s.events, annResp(t, i))
		s.i++
	}
}

func (s *server) generate(period time.Duration) {
	t := time.NewTicker(period)
	for {
		select {
		case <-t.C:
			n := time.Now()
			s.RLock()
			s.events = append(s.events, annResp(n, s.i))
			s.i++
			s.RUnlock()
		case <-s.ctx.Done():
			return
		}
	}
}

// root exists so that jsonds can be successfully added as a Grafana Data Source.
//
// If this exists then Grafana emits this when adding the datasource:
//
//		Success
// 		Data source is working
//
// otherwise it emits "Unknown error"
func (s *server) root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok\n")
}

func (s *server) latencyLoop() {
	for {
		s.apiConn.SessionID = "srsLTE"
		dtStart, err := time.Parse(time.RFC3339Nano, "2019-01-15T13:19:40.000Z")
		if err != nil {
			fmt.Println(err.Error())
		}
		dtEnd := dtStart.Add(1 * time.Second)
		cnt, cntLost, err := getLatency(dtStart, dtEnd, 17)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println(*cnt, *cntLost)
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func (s *server) annotations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
	case http.MethodPost:
		ar := AnnotationsReq{}
		if err := json.NewDecoder(r.Body).Decode(&ar); err != nil {
			http.Error(w, fmt.Sprintf("json decode failure: %v", err), http.StatusBadRequest)
			return
		}

		evs := s.filterEvents(ar.Annotation, ar.Range.From, ar.Range.To)
		if err := json.NewEncoder(w).Encode(evs); err != nil {
			log.Printf("json enc: %+v", err)
		}
	default:
		http.Error(w, "bad method; supported OPTIONS, POST", http.StatusBadRequest)
		return
	}
}

func (s *server) getProcessingTimeThresholdResult(q *QueryRequest, t *string, args ...string) *QueryResponseTable {
	qResp := QueryResponseTable{Target: *t}
	qResp.Columns = append(qResp.Columns, TableColumn{Text: "TMean", Type: "number"})
	qResp.Columns = append(qResp.Columns, TableColumn{Text: "TStd", Type: "number"})

	dataPoint := make([]interface{}, 2)
	dataPoint[0] = 60
	dataPoint[1] = 0
	qResp.Rows = append(qResp.Rows, dataPoint)

	dataPoint = make([]interface{}, 2)
	dataPoint[0] = 0
	dataPoint[1] = 20
	qResp.Rows = append(qResp.Rows, dataPoint)

	return &qResp
}

func (s *server) getProcessingTimeStatsResult(q *QueryRequest, t *string, args ...string) *QueryResponseTable {
	var res *[]models.APIProcessingState
	res, err := s.apiConn.GetFunctionProcessingStates(args[1], &q.Range.From, &q.Range.To)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	qResp := QueryResponseTable{Target: *t}
	qResp.Columns = append(qResp.Columns, TableColumn{Text: "Mean", Type: "number"})
	qResp.Columns = append(qResp.Columns, TableColumn{Text: "Std", Type: "number"})
	qResp.Columns = append(qResp.Columns, TableColumn{Text: "Time", Type: "number"})
	for _, p := range *res {
		dataPoint := make([]interface{}, 3)
		dataPoint[0] = p.Result.ProcessingTimeMoments[0]
		dataPoint[1] = math.Sqrt(float64(p.Result.ProcessingTimeMoments[1]) - math.Pow(float64(p.Result.ProcessingTimeMoments[0]), 2))
		dataPoint[2] = p.LastReportAt.Unix()
		qResp.Rows = append(qResp.Rows, dataPoint)
	}
	return &qResp
}

func (s *server) getProcessingTimeResult(q *QueryRequest, t *string, args ...string) *QueryResponseTimeserie {
	var res *[]models.APIProcessingState
	res, err := s.apiConn.GetFunctionProcessingStates(args[1], &q.Range.From, &q.Range.To)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	qResp := QueryResponseTimeserie{Target: *t}
	for _, p := range *res {
		dataPoint := make([]interface{}, 2)
		dataPoint[0] = p.Result.ProcessingTimeMoments[0]
		dataPoint[1] = p.FirstReportAt.UnixNano() / 1e6
		qResp.DataPoints = append(qResp.DataPoints, dataPoint)
	}
	return &qResp
}

func (s *server) getUEResult(q *QueryRequest, t *string, args ...string) *QueryResponseTimeserie {

	var res *[]models.APIUeState

	isUe := false
	if args[0] == "ueself" {
		isUe = true
	}

	res, err := s.apiConn.GetUEStates(args[1], &q.Range.From, &q.Range.To, &isUe)
	if err != nil {
		fmt.Println(args[1], q.Range.From, q.Range.To, isUe)
		log.Println(err.Error())
		return nil
	}

	qResp := QueryResponseTimeserie{Target: *t}
	for _, p := range *res {
		dataPoint := make([]interface{}, 2)
		var report *models.APIUeMacPhyReport
		var rfReport *models.APIUeRfReport
		if args[2] == "dl" {
			report = p.Result.MacPhyReportDl
			rfReport = p.Result.RfReportDl
		} else if args[2] == "ul" {
			report = p.Result.MacPhyReportUl
			rfReport = p.Result.RfReportUl
		} else {
			return nil
		}

		switch args[3] {
		case "mcs":
			dataPoint[0] = report.Mcs
			break
		case "snr":
			dataPoint[0] = rfReport.Snr
			break
		case "tp":
			dataPoint[0] = (report.MacTp / 1e6) * 8
			break
		case "nbrb":
			dataPoint[0] = report.NbRb
			break
		case "wbcqi":
			dataPoint[0] = report.WidebandCqi
			break
		}

		dataPoint[1] = p.FirstReportAt.UnixNano() / 1e6
		qResp.DataPoints = append(qResp.DataPoints, dataPoint)
	}
	return &qResp
}

func (s *server) getQueryResult(q QueryRequest) *[]interface{} {
	if val, ok := q.ScopedVars["SessionID"]; ok {
		s.apiConn.SessionID = val.Text.(string)
	} else {
		return nil
	}

	resp := make([]interface{}, len(q.Targets))
	for ti, target := range q.Targets {
		if target.Type == "timeseries" {
			args := strings.Split(target.Target, "_")

			if len(args) < 2 {
				continue
			}
			if args[0] == "ptime" {
				resp[ti] = s.getProcessingTimeResult(&q, &target.Target, args...)
			}
			if args[0] == "ue" || args[0] == "ueself" {
				resp[ti] = s.getUEResult(&q, &target.Target, args...)
			}
		} else if target.Type == "table" {
			args := strings.Split(target.Target, "_")

			if len(args) < 2 {
				continue
			}
			if args[0] == "statsptime" {
				resp[ti] = s.getProcessingTimeStatsResult(&q, &target.Target, args...)
			}
			if args[0] == "threshold" {
				resp[ti] = s.getProcessingTimeThresholdResult(&q, &target.Target, args...)
			}
		}
	}
	return &resp
}

func (s *server) queries(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
	case http.MethodPost:
		qr := QueryRequest{}
		if err := json.NewDecoder(r.Body).Decode(&qr); err != nil {
			http.Error(w, fmt.Sprintf("json decode failure: %v", err), http.StatusBadRequest)
			return
		}

		resp := s.getQueryResult(qr)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("json enc: %+v", err)
		}
	default:
		http.Error(w, "bad method; supported OPTIONS, POST", http.StatusBadRequest)
		return
	}
}

func (s *server) searches(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
	case http.MethodPost:
		resp := []string{}
		resp = append(resp, "threshold_DL encoding")
		resp = append(resp, "statsptime_DL encoding")
		resp = append(resp, "ptime_DL encoding")
		resp = append(resp, "ptime_UL decoding")
		resp = append(resp, "ue_0_dl_mcs")
		resp = append(resp, "ue_0_dl_tp")
		resp = append(resp, "ue_0_dl_nbrb")
		resp = append(resp, "ue_0_dl_wbcqi")
		resp = append(resp, "ueself_0_dl_snr")
		resp = append(resp, "ue_0_ul_mcs")
		resp = append(resp, "ue_0_ul_tp")
		resp = append(resp, "ue_0_ul_nbrb")
		resp = append(resp, "ue_0_ul_wbcqi")
		resp = append(resp, "ue_0_ul_snr")

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("json enc: %+v", err)
		}
	default:
		http.Error(w, "bad method; supported OPTIONS, POST", http.StatusBadRequest)
		return
	}
}

func (s *server) filterEvents(a Annotation, from, to time.Time) []AnnotationResponse {
	events := []AnnotationResponse{}
	for _, event := range s.events {
		event.Annotation = a
		event.Annotation.ShowLine = true
		if event.Time > from.Unix()*1000 && event.Time < to.Unix()*1000 {
			events = append(events, event)
		}
	}
	return events
}

// annResp isn't required; it just codifies a standard AnnotationResponse
// between the seed and generate funcs.
func annResp(t time.Time, i int) AnnotationResponse {
	return AnnotationResponse{
		// Grafana expects unix milliseconds:
		// https://github.com/grafana/simple-json-datasource#annotation-api
		Time: t.Unix() * 1000,

		Title: fmt.Sprintf("event %04d", i),
		Text:  fmt.Sprintf("text about the event %04d", i),
		Tags:  "atag btag ctag",
	}
}
