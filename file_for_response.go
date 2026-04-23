package goblet

import "io"

type fileWithName struct {
	origin io.ReadSeeker
	name   string
}

// FileWithName make file can by download for another name
func FileWithName(file io.ReadSeeker, name string) *fileWithName {
	return &fileWithName{
		origin: file,
		name:   name,
	}
}

func (f *fileWithName) Read(p []byte) (n int, err error) {
	return f.origin.Read(p)
}

func (f *fileWithName) Seek(offset int64, whence int) (int64, error) {
	return f.origin.Seek(offset, whence)
}

func (f *fileWithName) GetName() string {
	return f.name
}

func (f *fileWithName) GetSize() int64 {
	var info, _ = f.origin.Stat()
	return info.Size()
}
