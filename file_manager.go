package goblet

import (
	"github.com/valyala/fasthttp"
	"mime/multipart"
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
	return &filerSaver{filepath.Join(path...), setName, cx.ctx, "", nil}
}

type filerSaver struct {
	path       string
	nameSetter func(string) (string, error)
	ctx        *fasthttp.RequestCtx
	key        string
	header     *multipart.FileHeader
}

func (f *filerSaver) From(key string) *filerSaver {
	f.key = key
	return f
}

//用于文件保存中的重命名，
func (f *filerSaver) NameBy(fn func(string) (string, error)) *filerSaver {
	f.nameSetter = fn
	return f
}

//Execute the file save process and return the result
func (f *filerSaver) Exec() (path string, status int, err error) {
	var header *multipart.FileHeader
	if header, err = f.ctx.FormFile(f.key); err != nil {
		status = SAVEFILE_FORMFILE_ERROR
	} else {
		if fname, er := f.nameSetter(f.header.Filename); er == nil {
			path = filepath.Join(f.path, fname)
			if err = fasthttp.SaveMultipartFile(header, path); err != nil {
				status = SAVEFILE_COPY_ERROR
			}
		} else {
			status = SAVEFILE_RENAME_ERROR_BY_USER
			err = er
		}
	}
	status = SAVEFILE_SUCCESS
	// if _, err := os.Stat(f.path); err == nil {
	// 	if err := os.MkdirAll(f.path, 0755); err == nil {
	// 		var file multipart.File
	// 		f.header, err = f.ctx.FormFile(f.key)
	// 		f.header.
	// 		if file != nil {
	// 			defer file.Close()
	// 		}
	// 		if err == nil {
	// 			fname := f.nameSetter(f.header.Filename)
	// 			var fwriter *os.File
	// 			path = filepath.Join(f.path, fname)
	// 			fwriter, err = os.Create(path)
	// 			if err == nil {
	// 				defer fwriter.Close()
	// 				_, err = io.Copy(fwriter, file)
	// 			} else {
	// 				status = SAVEFILE_COPY_ERROR
	// 			}
	// 		} else {
	// 			status = SAVEFILE_FORMFILE_ERROR
	// 		}
	// 	} else {
	// 		status = SAVEFILE_CREATE_DIR_ERROR
	// 	}
	// } else {
	// 	status = SAVEFILE_STATE_DIR_ERROR
	// }
	return
}

func setName(fname string) (string, error) {
	return fname, nil
}
