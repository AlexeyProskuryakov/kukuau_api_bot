package web

import (
	"log"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/binding"
	"net/http"
	"fmt"

	d "msngr/db"
	"msngr/utils"
	"time"
)

type User interface {
	CanRead(right string) bool
	CanWrite(right string) bool
	IsAuthenticated() bool
	UniqueId() string
	RoleName() string
	BelongsToCompany() string
}

type user struct {
	*d.UserData
	LoginName string `form:"login"`
	Password  string `form:"password"`
}

func NewUser(data *d.UserData) User {
	result := user{UserData:data, LoginName:data.UserName}
	return &result
}

func (u *user) CanRead(right string) bool {
	if utils.InS(right, u.ReadRights) {
		return true
	}
	return false
}

func (u *user) CanWrite(right string) bool {
	if utils.InS(right, u.WriteRights) {
		return true
	}
	return false
}

func (u *user) IsAuthenticated() bool {
	log.Printf("User is auth %v", u.Auth)
	return u.Auth
}

func (u *user) UniqueId() string {
	log.Printf("User unique id [%+v]", u.UserId)
	return u.UserId
}

func (u *user) RoleName() string {
	log.Printf("User role: %v", u.Role)
	return u.Role
}

func (u *user) BelongsToCompany() string {
	log.Printf("User belongs to %v", u.BelongsTo)
	return u.BelongsTo
}

const (
	AUTH_URL = "/"
	REDIRECT_PARAM = "from"
	COOKIE_NAME = "current_user_id"
)

var flash = Flash{}

func StartAuthSession(user User, w http.ResponseWriter) {
	expiration := time.Now().Add(7 * 24 * time.Hour)
	cookie := http.Cookie{Name: COOKIE_NAME, Value: user.UniqueId(), Expires: expiration, Path:"/"}
	log.Printf("AUTH setting cookie: %+v", cookie)
	http.SetCookie(w, &cookie)
}

func StopAuthSession(w http.ResponseWriter) {
	cookie := http.Cookie{Name: COOKIE_NAME, Value:"delete", MaxAge:-1, Path:"/", Expires:time.Now().AddDate(-1, 0, 0)}
	log.Printf("AUTH removing cookie: %+v", cookie)
	http.SetCookie(w, &cookie)
}

func EnsureAuth(r martini.Router, mainDb *d.MainDb) martini.Router {
	r.Get("/", func(r render.Render, prms martini.Params, req *http.Request) {
		flashMessage, fType := flash.GetMessage()
		query := req.URL.Query()

		result := map[string]interface{}{
			fmt.Sprintf("flash_%v", fType):flashMessage,
			"from": query.Get("from"),
		}
		r.HTML(200, "login", result, render.HTMLOptions{Layout:"base"})
	})

	r.Post("/", binding.Bind(user{}), func(postedUser user, r render.Render, req *http.Request, w http.ResponseWriter) {
		userData, err := mainDb.Users.LoginUser(postedUser.LoginName, postedUser.Password)
		if err != nil {
			log.Printf("AUTH user %+v not found: %v", postedUser, err)
			flash.SetMessage("К сожалению, пользователь с такими данными не найден.", "error")
			r.Redirect(AUTH_URL)
			return
		} else {
			log.Printf("AUTH found user data: %v, %v, %v", userData.UserId, userData.UserName, userData.Auth)
		}
		user := NewUser(userData)
		StartAuthSession(user, w)
		redirect := req.URL.Query().Get(REDIRECT_PARAM)
		if redirect == "" {
			redirect = DefaultUrlMap.GetDefaultUrl(user.BelongsToCompany())
		}
		http.Redirect(w, req, redirect, 302)
	})
	return r
}

func NewSessionAuthorisationHandler(mainDb *d.MainDb, ) http.Handler {
	m := martini.New()
	m.Use(NonJsonLogger())
	m.Use(martini.Recovery())
	m.Use(martini.Static("static"))
	m.Use(render.Renderer(render.Options{
		Directory:"templates/auth",
		Extensions: []string{".tmpl", ".html"},
		Charset: "UTF-8",
	}))

	r := martini.NewRouter()
	r = EnsureAuth(r, mainDb)
	m.Action(r.Handle)
	return m
}

func LoginRequired(db d.DB, c martini.Context, r render.Render, req *http.Request) {
	cookie, err := req.Cookie(COOKIE_NAME)
	if err != nil {
		log.Printf("cookie getting error %v, redirect", err)
		path := fmt.Sprintf("%s?%s=%s", AUTH_URL, REDIRECT_PARAM, req.URL.Path)
		r.Redirect(path, 302)
		return
	}
	userId := cookie.Value
	log.Printf("found userId : [%v] cookie: {%+v}", userId, cookie)
	userData, err := db.UsersStorage().GetUserById(userId)
	if err != nil {
		log.Printf("can not find user by [%v], because: %v", userId, err)
		path := fmt.Sprintf("%s?%s=%s", AUTH_URL, REDIRECT_PARAM, req.URL.Path)
		r.Redirect(path, 302)
		return
	}
	log.Printf("load user data: %+v", userData)
	user := NewUser(userData)
	if user.IsAuthenticated() {
		c.MapTo(user, (*User)(nil))
		return
	}
	log.Printf("user not authentificated :( ")
	path := fmt.Sprintf("%s?%s=%s", AUTH_URL, REDIRECT_PARAM, req.URL.Path)
	r.Redirect(path, 302)
}