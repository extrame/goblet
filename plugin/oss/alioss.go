package oss

import (
	"errors"
	"github.com/denverdino/aliyungo/oss"
	toml "github.com/extrame/go-toml-config"
	"github.com/extrame/goblet"
	"io"
)

type Alioss struct {
	region          *string
	accessKeyId     *string
	accessKeySecret *string
}

var ossclient *oss.Client

func (r *Alioss) Init(server *goblet.Server) error {
	ossclient = newSaver(r.region, r.accessKeyId, r.accessKeySecret)
	return nil
}

func (r *Alioss) ParseConfig(prefix string) error {
	r.region = toml.String(prefix+".region", "beijing")
	r.accessKeyId = toml.String(prefix+".access_id", "")
	r.accessKeySecret = toml.String(prefix+".access_secret", "")
	return nil
}

func newSaver(region, accessKeyId, accessKeySecret *string) (l *oss.Client) {
	reg := oss.Beijing
	switch *region {
	case "beijing":
		reg = oss.Beijing
	case "qingdao":
		reg = oss.Qingdao
	case "shanghai":
		reg = oss.Shanghai
	case "shenzheng":
		reg = oss.Shenzhen
	case "hangzhou":
		reg = oss.Hangzhou
	case "hongkong":
		reg = oss.Hongkong
	}

	return oss.NewOSSClient(reg, false, *accessKeyId, *accessKeySecret, false)
}

func InitSaver(region, accessKeyId, accessKeySecret string) {
	ossclient = newSaver(&region, &accessKeyId, &accessKeySecret)
}

func PutFile(bucket, path string, file io.Reader, size int64) error {
	if ossclient == nil {
		errors.New("oss init failed,please check your config file or call InitSaver munal!")
	}
	bkt := ossclient.Bucket(bucket)
	err := bkt.PutReader(path, file, size, "application/octet-stream", oss.Private, oss.Options{})
	return err
}

func GetFile(bucket, path string) (io.ReadCloser, error) {
	if ossclient == nil {
		errors.New("oss init failed,please check your config file or call InitSaver munal!")
	}
	bkt := ossclient.Bucket(bucket)
	res, err := bkt.GetReader(path)
	return res, err
}
