package httpx

type Unmarshaler interface {
	UnmarshalJSONX([]byte) error
}
