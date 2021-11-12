package oss

import (
	"errors"
	"io"

	"github.com/denverdino/aliyungo/oss"
	"github.com/extrame/goblet"
)

type Alioss struct {
	Region          string `goblet:"region,beijing"`
	AccessKeyId     string `goblet:"access_id"`
	AccessKeySecret string `goblet:"access_secret"`
}

var ossclient *oss.Client

func (r *Alioss) Init(server *goblet.Server) error {
	ossclient = newSaver(r.Region, r.AccessKeyId, r.AccessKeySecret)
	return nil
}

func (s *Alioss) AddCfgAndInit(server *goblet.Server) error {
	server.AddConfig("alioss", &s)
	return s.Init(server)
}

func newSaver(region, accessKeyId, accessKeySecret string) (l *oss.Client) {
	reg := oss.Beijing
	switch region {
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

	return oss.NewOSSClient(reg, false, accessKeyId, accessKeySecret, false)
}

func InitSaver(region, accessKeyId, accessKeySecret string) {
	ossclient = newSaver(region, accessKeyId, accessKeySecret)
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
