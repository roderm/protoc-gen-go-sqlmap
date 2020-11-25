package generator

type Printer interface {
	StoreName() string
	P(str ...interface{})
	Write(p []byte) (n int, err error)
}
