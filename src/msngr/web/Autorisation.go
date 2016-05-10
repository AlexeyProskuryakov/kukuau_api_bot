package web

import (
	"github.com/martini-contrib/render"
	"net/http"
	"fmt"
)

type AuthMap struct {
	//map for mapping default pages of chats or another services with users rights
	data map[string]string
}

func (a *AuthMap) AddAccessory(companyId, defaultUrl string) {
	a.data[companyId] = defaultUrl
}

func (a *AuthMap) GetDefaultUrl(companyId string) string {
	if dUrl, ok := a.data[companyId]; ok {
		return dUrl
	} else {
		return ""
	}
}

var DefaultUrlMap = &AuthMap{data:make(map[string]string)}

type authHandler struct {
	RedirectUrl string
}

var AutHandler = &authHandler{RedirectUrl:AUTH_URL}

const (
	NOT_BELONG_ROLES = "not_belong_role"
	CAN_NOT_READ = "can_not_read"
	CAN_NOT_WRITE = "can_not_write"
	MANAGER_ROLE = "manager"
)

func (ah *authHandler) CheckIncludeAnyRole(roles ...string) func(r render.Render, user User, req *http.Request) {
	return func(r render.Render, user User, req *http.Request) {
		if user.RoleName() == MANAGER_ROLE {
			return
		}
		for _, role := range roles {
			if user.RoleName() == role {
				return
			}
		}
		path := fmt.Sprintf("%s?%s=%s", ah.RedirectUrl, NOT_BELONG_ROLES, req.URL.Path)
		r.Redirect(path, 302)
	}
}

func (ah *authHandler) CheckReadRights(rights ...string) func(r render.Render, user User, req *http.Request) {
	return func(r render.Render, user User, req *http.Request) {
		if user.RoleName() == MANAGER_ROLE {
			return
		}
		for _, right := range rights {
			if !user.CanRead(right) {
				path := fmt.Sprintf("%s?%s=%s", ah.RedirectUrl, CAN_NOT_READ, req.URL.Path)
				r.Redirect(path, 302)
			}
		}
	}
}

func (ah *authHandler) CheckWriteRights(rights ...string) func(r render.Render, user User, req *http.Request) {
	return func(r render.Render, user User, req *http.Request) {
		if user.RoleName() == MANAGER_ROLE {
			return
		}
		for _, right := range rights {
			if !user.CanWrite(right) {
				path := fmt.Sprintf("%s?%s=%s", ah.RedirectUrl, CAN_NOT_WRITE, req.URL.Path)
				r.Redirect(path, 302)
			}
		}
	}
}