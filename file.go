package goblet

import (
	"io"
	"os"
	"path/filepath"
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

//将文件保存在公开目录，可以使用http访问到
func (s *Server) SaveInPublic(path string, f *File) error {
	fullPath := filepath.Join(*s.wwwRoot, s.PublicDir(), path)
	return s.saver.Save(fullPath, f)
}

//将文件保存在私有目录，不可以使用http访问到
func (s *Server) SaveInPrivate(path string, f *File) error {
	fullPath := filepath.Join(*s.wwwRoot, path)
	return s.saver.Save(fullPath, f)
}

type Saver interface {
	Save(fullpath string, f io.Reader) error
}

type LocalSaver struct {
}

func (l *LocalSaver) Save(path string, f io.Reader) error {
	var file *os.File
	var err error
	if file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666); err == nil {
		io.Copy(file, f)
		file.Close()
	}
	return err
}
