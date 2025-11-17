package main

import (
	"sync"
)

//
// I dont like this much but we need to serialize access to operations and doing it at the
// web service is the most reliable way to do it.
//
// The nature of easystore updates make them multi-step and non-tractional so access concurrency
// needs to be managed :(

var locks = sync.Map{}

func accessLock(id string) {

	var m *sync.Mutex
	var ok bool

	a, ok := locks.Load(id)
	if ok {
		m = a.(*sync.Mutex)
	} else {
		m = &sync.Mutex{}
		locks.Store(id, m)
	}
	m.Lock()
}

func accessUnlock(id string) {

	a, ok := locks.Load(id)
	if ok {
		m := a.(*sync.Mutex)
		m.Unlock()
	}
}

//
// end of file
//
