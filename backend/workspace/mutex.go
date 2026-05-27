// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package workspace

import (
	"sync"
)

// Credit: https://stackoverflow.com/questions/40931373/how-to-gc-a-map-of-mutexes-in-go/62562831#62562831
// Package mutex provides locking per-key.
// For example, you can acquire a lock for a specific user ID and all other requests for that user ID
// will block until that entry is unlocked (effectively your work load will be run serially per-user ID),
// and yet have work for separate user IDs happen concurrently.

// Mutex wraps a map of mutexes. Each key locks separately.
type Mutex struct {
	mapLock sync.Mutex             // lock for entry map
	ma      map[string]*mutexEntry // entry map
}

type mutexEntry struct {
	thisMutexMap *Mutex     // point back to Mutex, so we can synchronize removing this mutexEntry when cnt==0
	el           sync.Mutex // entry-specific lock
	cnt          int        // reference count
	key          string     // key in ma
}

// Unlocker provides an Unlock method to release the lock.
type Unlocker interface {
	Unlock()
}

// NewMutex returns an initialized Mutex.
func NewMutex() *Mutex {
	return &Mutex{ma: make(map[string]*mutexEntry)}
}

// Lock acquires a lock corresponding to this key.
// This method will never return nil and Unlock() must be called
// to release the lock when done.
func (m *Mutex) Lock(key string) Unlocker {
	// read or create entry for this key atomically
	m.mapLock.Lock()
	entry, ok := m.ma[key]
	if !ok {
		entry = &mutexEntry{thisMutexMap: m, key: key}
		m.ma[key] = entry
	}
	entry.cnt++ // ref count
	m.mapLock.Unlock()

	// acquire lock, will block here until entry.cnt==1
	entry.el.Lock()

	return entry
}

// Unlock releases the lock for this entry.
func (entry *mutexEntry) Unlock() {
	thisMutexMap := entry.thisMutexMap
	// decrement and if needed, remove the entry atomically
	thisMutexMap.mapLock.Lock()
	entry.cnt--        // ref count
	if entry.cnt < 1 { // if it hits zero then we own it and remove from map
		delete(thisMutexMap.ma, entry.key)
	}
	thisMutexMap.mapLock.Unlock()
	// now that map stuff is handled, we unlock and let
	// anything else waiting on this key through
	entry.el.Unlock()
}
