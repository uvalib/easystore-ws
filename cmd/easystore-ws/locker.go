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

var locksAccess sync.Mutex
var locks = make(map[string]*sync.Mutex)

func accessLock(id string) {

	_, ok := locks[id]
	if !ok {
		// lock the lock map
		locksAccess.Lock()

		// another check now that we have exclusive write access to the lock map
		_, ok := locks[id]
		if !ok {
			locks[id] = &sync.Mutex{}
		}

		// release the lock map
		locksAccess.Unlock()
	}
	locks[id].Lock()
}

func accessUnlock(id string) {

	_, ok := locks[id]
	if ok {
		locks[id].Unlock()
	}
}

//
// end of file
//
