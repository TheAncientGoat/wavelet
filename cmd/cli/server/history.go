package server

import (
	"sync"
	"time"
)

type HistoryEntry struct {
	ID   string
	Type HistoryEntryType
	Time time.Time
}

func (e HistoryEntry) String() string {
	var s = e.ID

	switch e.Type {
	case HistoryEntrySelf:
		s += " - Self"
	case HistoryEntryRoot:
		s += " - Root"
	case HistoryEntryAccount:
		s += " - Account"
	case HistoryEntryTransaction:
		s += " - Transaction"
	}

	s += " at " + e.Time.Format("15:04:05")

	return s
}

type HistoryEntryType int

const (
	HistoryEntrySelf HistoryEntryType = iota
	HistoryEntryRoot
	HistoryEntryAccount
	HistoryEntryTransaction
)

type HistoryStore struct {
	Store []HistoryEntry
	mu    sync.Mutex
}

// Add adds an entry into History. If an entry duplicates, the Time and Type
// will be renewed.
func (s *HistoryStore) add(ID string, t HistoryEntryType) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := 0; i < len(s.Store); i++ {
		if s.Store[i].ID == ID {
			s.Store[i].Time = time.Now()
			s.Store[i].Type = t

			return
		}
	}

	s.Store = append(s.Store, HistoryEntry{
		ID:   ID,
		Type: t,
		Time: time.Now(),
	})
}
