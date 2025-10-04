package data

import (
	"log/slog"
	"os"

	"github.com/rustyeddy/otto/messanger"
)

type Store struct {
	Filename string
	Datas    map[string]*Timeseries
	StoreQ   chan *messanger.Msg

	f *os.File
}

var (
	store *Store
)

func NewFileStore(fname string) *Store {
	m := make(map[string]*Timeseries)
	q := make(chan *messanger.Msg)
	store := &Store{
		Datas:  m,
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

func (s *Store) Save(label string, ts *Timeseries) error {
	s.Datas[label] = ts
	// now write to file
	return nil
}

func (s *Store) Load(label string) (*Timeseries, error) {
	return s.Datas[label], nil

}
