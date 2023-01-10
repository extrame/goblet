package render

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

type RawRender int8

func (r *RawRender) PrepareInstance(c RenderContext) (RenderInstance, error) {
	return new(RawRenderInstance), nil
}

func (r *RawRender) Init(s RenderServer, funcs template.FuncMap) {
}

type RawRenderInstance int8

//interface to respond file with customerized name and size
type RawFile interface {
	io.Reader
	io.Seeker
	GetName() string
	GetSize() int64
}

func (r *RawRenderInstance) HeadRender(wr io.Writer, hwr HeadWriter, data interface{}, status int, funcs template.FuncMap) (err error) {
	var writen = int64(0)
	switch tdata := data.(type) {
	case RawFile:
		writen = tdata.GetSize()
	case *os.File:
		var info, _ = tdata.Stat()
		writen = info.Size()
	case io.Reader:
		writen, err = io.Copy(wr, tdata)
	case []byte:
		var int_writen = 0
		int_writen, err = wr.Write(tdata)
		writen = int64(int_writen)
	}
	hwr.Header().Set("Content-Length", strconv.FormatInt(writen, 10))
	return err
}

func (r *RawRenderInstance) Render(wr io.Writer, hwr HeadWriter, data interface{}, status int, funcs template.FuncMap) (err error) {
	var name string
	var seeker io.Seeker
	var reader io.Reader
	var size int64
	switch tdata := data.(type) {
	case RawFile:
		name = tdata.GetName()
		size = tdata.GetSize()
		seeker = tdata
		reader = tdata
		goto returnFile
	case *os.File:
		var info, _ = tdata.Stat()
		seeker = tdata
		reader = tdata
		name = info.Name()
		size = info.Size()
		goto returnFile
	case io.Reader:
		if bts, err := ioutil.ReadAll(tdata); err == nil {
			hwr.Header().Set("Content-Length", strconv.FormatInt(int64(len(bts)), 10))
			io.Copy(wr, bytes.NewBuffer(bts))
		}
	case string:
		hwr.Header().Set("Content-Length", strconv.FormatInt(int64(len(tdata)), 10))
		wr.Write([]byte(tdata))
	case []byte:
		hwr.Header().Set("Content-Length", strconv.FormatInt(int64(len(tdata)), 10))
		wr.Write(tdata)
	}
	// hwr.Header().Set("Content-Length", strconv.FormatInt(writen, 10))
	return err
returnFile:
	if hwr.Header().Get("Content-Type") == "" {
		hwr.Header().Set("Content-Type", "application/octet-stream")
		hwr.Header().Set("Content-Disposition", "attachment; filename="+name)
	}
	var ret int64
	if ret, err = seeker.Seek(0, 1); err == nil {
		hwr.Header().Set("Content-Length", strconv.FormatInt(size-ret, 10))
	}
	io.Copy(wr, reader)
	return err
}
