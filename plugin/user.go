package plugin

import (
	"net/http"

	"github.com/extrame/goblet"
	"gorm.io/gorm"
)

type User struct {
	//用于启动或者不启动Create方法，当网站不需要新建用户时，
	CreateOnlyByName    string
	CreateOnlyByPermits []string
}

type UserModule struct {
	Id      int64           `gorm:"primaryKey"`
	Name    string          `gorm:"uniqueIndex" goblet:"name"`
	Pwd     string          `goblet:"pwd,md5"`
	Permits map[string]bool `gorm:"serializer:json"`
}

func (u *UserModule) TableName() string {
	return "user"
}

func (u *User) UpdateMany(cx *goblet.Context) {
	rec := new(UserModule)
	cx.Fill(rec)
	if rec.Name != "" && rec.Pwd != "" {
		result := cx.DB().Where("name = ? AND pwd = ?", rec.Name, rec.Pwd).First(rec)
		if result.Error == nil {
			cx.AddLoginId(rec.Id)
			cx.RespondOK()
		} else if result.Error == gorm.ErrRecordNotFound {
			cx.RespondWithStatus("用户名或密码错误", http.StatusForbidden)
		} else {
			cx.RespondWithStatus("数据库错误", http.StatusInternalServerError)
		}
	} else {
		cx.RespondWithStatus("用户名或密码为空", http.StatusForbidden)
	}
}

func (u *User) Create(cx *goblet.Context) {
	rec := new(UserModule)
	cx.Fill(rec)
	if rec.Name != "" && rec.Pwd != "" {
		result := cx.DB().Create(rec)
		if result.Error != nil {
			cx.AddRespond("err", result.Error)
			cx.RespondStatus(http.StatusBadRequest)
		} else {
			cx.AddRespond("user", rec)
		}
	} else {
		cx.RespondWithStatus("用户名或密码为空", http.StatusBadRequest)
	}
}

func (u *User) New(cx *goblet.Context) {
	if len(u.CreateOnlyByPermits) > 0 {
		var user UserModule
		if id, has := cx.GetLoginId(); has {
			result := cx.DB().First(&user, id)
			if result.Error == nil {
				for _, permit := range u.CreateOnlyByPermits {
					if _, ok := user.Permits[permit]; ok {
						cx.RespondOK()
						return
					}
				}
			}
		}
	}
	cx.RespondStatus(http.StatusMethodNotAllowed)
}
