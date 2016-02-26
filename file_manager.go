package goblet

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const (
	SAVEFILE_SUCCESS              = iota
	SAVEFILE_STATE_DIR_ERROR      = iota
	SAVEFILE_CREATE_DIR_ERROR     = iota
	SAVEFILE_FORMFILE_ERROR       = iota
	SAVEFILE_RENAME_ERROR_BY_USER = iota
	SAVEFILE_COPY_ERROR           = iota
)

func (cx *Context) SaveFileAt(path ...string) *filerSaver {
	path = append([]string{*cx.Server.UploadsDir}, path...)
	return &filerSaver{filepath.Join(path...), "", setName, cx.request, "", nil}
}

type filerSaver struct {
	path string
	//文件保存的完整路径
	FullPath   string
	nameSetter func(string) (string, error)
	request    *http.Request
	key        string
	header     *multipart.FileHeader
}

func (f *filerSaver) From(key string) *filerSaver {
	f.key = key
	return f
}

func (f *filerSaver) NameBy(fn func(string) (string, error)) *filerSaver {
	f.nameSetter = fn
	return f
}

//Execute the file save process and return the result
func (f *filerSaver) Exec() (path string, status int, err error) {
	if _, err := os.Stat(f.path); err == nil {
		if err := os.MkdirAll(f.path, 0755); err == nil {
			var file multipart.File
			file, f.header, err = f.request.FormFile(f.key)
			if file != nil {
				defer file.Close()
			}
			if err == nil {
				if fname, er := f.nameSetter(f.header.Filename); er == nil {
					var fwriter *os.File
					f.FullPath = filepath.Join(f.path, fname)
					fwriter, err = os.Create(f.FullPath)
					if err == nil {
						defer fwriter.Close()
						_, err = io.Copy(fwriter, file)
					} else {
						status = SAVEFILE_COPY_ERROR
					}
				} else {
					err = er
					status = SAVEFILE_RENAME_ERROR_BY_USER
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

func setName(fname string) (string, error) {
	return fname, nil
}
