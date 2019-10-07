package main

import (
	"bytes"
	"sync"
)

type ConvertManager struct {
	ResultsQueue map[string]*ConvertQueue
	ResultsLock  sync.Mutex
}

type ConvertQueue struct {
	Results chan ConvertResult
	Waiting int32
}

type ConvertResult struct {
	Data  *bytes.Buffer
	Error error
}

func NewConvertManager() *ConvertManager {
	return &ConvertManager{
		ResultsQueue: map[string]*ConvertQueue{},
	}
}

var convertQueue = NewConvertManager()
