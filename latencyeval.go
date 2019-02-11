package main

import (
	"fmt"
	"time"
)

func getLatency(dtStart time.Time, dtEnd time.Time, protocolTypeID int) (*int, *int, error) {
	isUE := false
	isReceived := false
	enbSentTmp, err := serv.apiConn.GetPacketsStatus(&isUE, &isReceived, &dtStart, &dtEnd, &protocolTypeID, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	isUE = true
	isReceived = true
	ueRecvTmp, err := serv.apiConn.GetPacketsStatus(&isUE, &isReceived, &dtStart, &dtEnd, &protocolTypeID, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println(len(*enbSentTmp), len(*ueRecvTmp))

	enbSent := serv.apiConn.ExcludeNullPacketID(enbSentTmp)
	ueRecv := serv.apiConn.ExcludeNullPacketID(ueRecvTmp)
	cntOk, cntLost, _, _ := serv.apiConn.CalculateLatency(&enbSent, &ueRecv)

	// First report dtStart
	// Last report dtEnd
	// totalCount
	// lostCount
	// latencyMoments []
	// statsInfo {min, max, avg, std}
	// isUl
	// ProtocolTypeID
	// totalSize

	return &cntOk, &cntLost, err
}
