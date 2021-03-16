package goblet

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type File struct {
	Name string
	Path string
	rc   multipart.File `xorm:"-"`
}

func (f *File) Read(p []byte) (n int, err error) {
	return f.rc.Read(p)
}

func (f *File) ReadAt(p []byte, off int64) (n int, err error) {
	return f.rc.ReadAt(p, off)
}

func (f *File) Close() error {
	return f.rc.Close()
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	return f.rc.Seek(offset, whence)
}

func (f *File) GetName() string {
	return f.Name
}

func (f *File) GetSize() int64 {
	if fi, ok := f.rc.(*os.File); ok {
		stat, err := fi.Stat()
		if err == nil {
			return stat.Size()
		}
	}
	return 0
}

func (f *File) SaveInPublic(dir string, s *Server) error {
	f.Path = filepath.Join(s.PublicDir(), dir, f.Name)
	return s.saver.Save(filepath.Join(*s.wwwRoot, f.Path), f)
}

func (f *File) SaveInTemp(dir string, s *Server) error {
	f.Path = filepath.Join(os.TempDir(), dir, f.Name)
	return s.saver.Save(f.Path, f)
}

func (f *File) SaveInPrivate(dir string, s *Server) error {
	f.Path = filepath.Join(dir, f.Name)
	return s.saver.Save(filepath.Join(*s.wwwRoot, f.Path), f)
}

func (f *File) OpenInPrivate(s *Server) error {
	fi, err := os.Open(filepath.Join(*s.wwwRoot, f.Path))
	if err == nil {
		f.rc = fi
	}
	return err
}

//将文件保存在公开目录，可以使用http访问到
func (s *Server) SaveInPublic(path string, f io.Reader) error {
	fullPath := filepath.Join(*s.wwwRoot, s.PublicDir(), path)
	return s.saver.Save(fullPath, f)
}

func (s *Server) DelFileInPrivate(path string) error {
	fullPath := filepath.Join(*s.wwwRoot, path)
	return os.Remove(fullPath)
}

func (s *Server) DelFileInPublic(path string) error {
	fullPath := filepath.Join(*s.wwwRoot, s.PublicDir(), path)
	return s.saver.Delete(fullPath)
}

//将文件保存在私有目录，不可以使用http访问到
func (s *Server) SaveInPrivate(path string, f *File) error {
	fullPath := filepath.Join(*s.wwwRoot, path)
	return s.saver.Save(fullPath, f)
}

type Saver interface {
	Save(fullpath string, f io.Reader) error
	Delete(fullpath string) error
}

type LocalSaver struct {
}

func (l *LocalSaver) Save(path string, f io.Reader) error {
	var file *os.File
	var err error
	err = os.MkdirAll(filepath.Dir(path), 0777)
	if err == nil {
		if file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666); err == nil {
			io.Copy(file, f)
			file.Close()
		}
	}
	return err
}

func (l *LocalSaver) Delete(path string) error {
	return os.Remove(path)
}
