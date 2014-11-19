package goblet

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const (
	SAVEFILE_SUCCESS          = iota
	SAVEFILE_STATE_DIR_ERROR  = iota
	SAVEFILE_CREATE_DIR_ERROR = iota
	SAVEFILE_FORMFILE_ERROR   = iota
	SAVEFILE_COPY_ERROR       = iota
)

func (cx *Context) SaveFileAt(path ...string) *filerSaver {
	path = append([]string{*cx.Server.WwwRoot, *cx.Server.PublicDir}, path...)
	return &filerSaver{filepath.Join(path...), setName, cx.Request, "", nil}
}

type filerSaver struct {
	path       string
	nameSetter func(string) string
	request    *http.Request
	key        string
	header     *multipart.FileHeader
}

func (f *filerSaver) From(key string) *filerSaver {
	f.key = key
	return f
}

func (f *filerSaver) NameBy(fn func(string) string) *filerSaver {
	f.nameSetter = fn
	return f
}

func (f *filerSaver) Exec() (status int, err error) {
	if _, err := os.Stat(f.path); err == nil {
		if err := os.MkdirAll(f.path, 0755); err == nil {
			var file multipart.File
			file, f.header, err = f.request.FormFile(f.key)
			if file != nil {
				defer file.Close()
			}
			if err == nil {
				fname := f.nameSetter(f.header.Filename)
				var fwriter *os.File
				fwriter, err = os.Create(fname)
				if err == nil {
					defer fwriter.Close()
					_, err = io.Copy(fwriter, file)
				} else {
					status = SAVEFILE_COPY_ERROR
				}
			} else {
				status = SAVEFILE_FORMFILE_ERROR
			}
		} else {
			status = SAVEFILE_CREATE_DIR_ERROR
		}
	} else {
		status = SAVEFILE_STATE_DIR_ERROR
	}
	return
}

func setName(fname string) string {
	return fname
}
