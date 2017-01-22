package goblet

import (
	"io"
)

type File struct {
	Name string
	rc   io.ReadCloser
}

func (f *File) Read(p []byte) (n int, err error) {
	return f.rc.Read(p)
}

func (f *File) Close() error {
	return f.rc.Close()
}
