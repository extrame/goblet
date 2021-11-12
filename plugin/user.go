package plugin

import (
	"net/http"

	"github.com/extrame/goblet"
)

type User struct {
	//用于启动或者不启动Create方法，当网站不需要新建用户时，
	CreateOnlyByName    string
	CreateOnlyByPermits []string
}

type UserModule struct {
	Id      int64
	Name    string `xorm:"unique" goblet:"name"`
	Pwd     string `goblet:"pwd,md5"`
	Permits map[string]bool
}

func (u *UserModule) TableName() string {
	return "user"
}

func (u *User) UpdateMany(cx *goblet.Context) {
	rec := new(UserModule)
	cx.Fill(rec)
	var err error
	if rec.Name != "" && rec.Pwd != "" {
		var has bool
		if has, err = goblet.DB.Where("name = ? and pwd = ?", rec.Name, rec.Pwd).Get(rec); err == nil && has {
			cx.AddLoginId(rec.Id)
			cx.RespondOK()
		} else {
			cx.RespondWithStatus("用户名或密码错误", http.StatusForbidden)
		}
	} else {
		cx.RespondWithStatus("用户名或密码为空", http.StatusForbidden)
	}

}

func (u *User) Create(cx *goblet.Context) {
	rec := new(UserModule)
	if rec.Name != "" && rec.Pwd != "" {
		if _, err := goblet.DB.Insert(rec); err != nil {
			cx.AddRespond("err", err)
			cx.RespondStatus(http.StatusBadRequest)
		} else {
			cx.AddRespond("user", rec)
		}
	}
}

func (u *User) New(cx *goblet.Context) {
	if len(u.CreateOnlyByPermits) > 0 {
		var user UserModule
		if id, has := cx.GetLoginId(); has {
			if has, err := goblet.DB.ID(id).Get(&user); err == nil && has {
				for _, permit := range u.CreateOnlyByPermits {
					if _, ok := user.Permits[permit]; ok {
						cx.RespondOK()
					}
				}
			}
		}
	}
	cx.RespondStatus(http.StatusMethodNotAllowed)
}
