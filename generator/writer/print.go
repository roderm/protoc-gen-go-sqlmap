package writer

type Printer interface {
	P(str ...interface{})
	Write(p []byte) (n int, err error)
}
