package web

import (
	"github.com/martini-contrib/render"
	"net/http"
	"fmt"
)

type AuthHandler struct {
	RedirectUrl string
}

const (
	NOT_BELONG_ROLES = "not_belong_role"
	CAN_NOT_READ = "can_not_read"
	CAN_NOT_WRITE = "can_not_write"
)

func (ah *AuthHandler) CheckIncludeAnyRole(roles ...string) func(r render.Render, user user, req *http.Request) {
	return func(r render.Render, user user, req *http.Request) {
		for _, role := range roles {
			if user.Role == role {
				return
			}
		}
		path := fmt.Sprintf("%s?%s=%s", ah.RedirectUrl, NOT_BELONG_ROLES, req.URL.Path)
		r.Redirect(path, 302)
	}
}

func (ah *AuthHandler) CheckReadRights(rights ...string) func(r render.Render, user user, req *http.Request) {
	return func(r render.Render, user user, req *http.Request) {
		for _, right := range rights {
			if !user.CanRead(right) {
				path := fmt.Sprintf("%s?%s=%s", ah.RedirectUrl, CAN_NOT_READ, req.URL.Path)
				r.Redirect(path, 302)
			}
		}
	}
}

func (ah *AuthHandler) CheckWriteRights(rights ...string) func(r render.Render, user user, req *http.Request) {
	return func(r render.Render, user user, req *http.Request) {
		for _, right := range rights {
			if !user.CanWrite(right) {
				path := fmt.Sprintf("%s?%s=%s", ah.RedirectUrl, CAN_NOT_WRITE, req.URL.Path)
				r.Redirect(path, 302)
			}
		}
	}
}