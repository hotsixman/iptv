package types

type Device struct {
	Name string
}

type Format struct {
	Codec  string
	Width  int
	Height int
	Fps    float64
}
