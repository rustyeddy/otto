package data

import (
	"log/slog"
	"os"

	"github.com/rustyeddy/otto/messanger"
)

type Store struct {
	Filename string
	Source   map[string]map[string]float64
	StoreQ   chan *messanger.Msg

	f *os.File
}

var (
	store *Store
)

func GetStore() *Store {
	if store == nil {
		store = NewStore()
	}
	return store
}

func NewStore() *Store {
	m := make(map[string]map[string]float64)
	q := make(chan *messanger.Msg)
	store := &Store{
		Source: m,
		StoreQ: q,
	}

	go func() {
		for {
			select {
			case msg := <-store.StoreQ:
				store.Store(msg)
			}
		}
	}()

	return store
}

func (s *Store) Store(msg *messanger.Msg) error {
	slog.Info("Store: ", "message", msg)
	return nil
}
