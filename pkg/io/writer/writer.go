package writer

import "io"

type OutputWriter interface {
	io.Writer
	io.Closer
}
