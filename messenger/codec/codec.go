package codec

type Codec[T any] interface {
	Marshal(v T) ([]byte, error)
	Unmarshal(b []byte) (T, error)
}
