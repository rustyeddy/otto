package codec

import "encoding/json"

type JSON[T any] struct{}

func (JSON[T]) Marshal(v T) ([]byte, error)   { return json.Marshal(v) }
func (JSON[T]) Unmarshal(b []byte) (T, error) { var v T; return v, json.Unmarshal(b, &v) }
