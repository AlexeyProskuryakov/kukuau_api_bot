package web

import (
	"log"
	"github.com/martini-contrib/sessions"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessionauth"
	"github.com/martini-contrib/binding"
	"net/http"
	"fmt"

	"msngr/db"
	"msngr/utils"
	"msngr/configuration"
)

type User struct {
	handler   *db.UserHandler
	Data      *db.UserData
	LoginName string `form:"login"`
	Password  string `form:"password"`
	Auth      bool
	Anonymous bool
}

func InitUserObject(data *db.UserData, uh *db.UserHandler) *User {
	log.Printf("will init user object")
	result := User{Data:data, handler:uh, Auth:data.Auth, Anonymous:false}
	return &result
}
func AnonymousUserInitializer(uh *db.UserHandler) func() sessionauth.User {
	log.Printf("will form func for init user")
	return func() sessionauth.User {
		log.Printf("initialize anonymus user")
		return &User{handler:uh, Anonymous:true}
	}
}

func (u *User) CanRead(right string) bool {
	if utils.InS(right, u.Data.ReadRights) {
		return true
	}
	return false
}

func (u *User) CanWrite(right string) bool {
	if utils.InS(right, u.Data.WriteRights) {
		return true
	}
	return false
}

func (u *User) IsAuthenticated() bool {
	log.Printf("is auth? %+v", u)
	return u.Auth
}

func (u *User) Login() {
	newData, err := u.handler.UserLogin(u.Data.UserId)
	if err == nil {
		u.Data = newData
		u.Auth = true
		u.Anonymous = false
	}
}

func (u *User) Logout() {
	log.Printf("logout %v", u.Data.UserId)
	newData, _ := u.handler.UserLogout(u.Data.UserId)
	u.Data = newData
	u.Auth = false
	u.Anonymous = true
}

func (u *User) UniqueId() interface{} {
	log.Printf("unique id %v", u.Data.UserId)
	return u.Data.UserId
}

func (u *User) GetById(id interface{}) error {
	found, err := u.handler.GetUserById(id.(string))
	log.Printf("found %+v", found)
	u.Data = found
	log.Printf("get by id %v", u.Data.UserId)
	return err
}

type AuthHandler struct {
	RedirectUrl string
}

const (
	NOT_BELONG_ROLES = "not_belong_role"
	CAN_NOT_READ = "can_not_read"
	CAN_NOT_WRITE = "can_not_write"
)

func (ah *AuthHandler) CheckIncludeAnyRole(roles ...string) func(r render.Render, sUser sessionauth.User, req *http.Request) {
	return func(r render.Render, sUser sessionauth.User, req *http.Request) {
		for _, role := range roles {
			user := sUser.(*User)
			if user.Data.Role == role {
				return
			}
		}
		path := fmt.Sprintf("%s?%s=%s", ah.RedirectUrl, NOT_BELONG_ROLES, req.URL.Path)
		r.Redirect(path, 302)
	}
}

func (ah *AuthHandler) CheckReadRights(rights ...string) func(r render.Render, sUser sessionauth.User, req *http.Request) {
	return func(r render.Render, sUser sessionauth.User, req *http.Request) {
		for _, right := range rights {
			user := sUser.(*User)
			if !user.CanRead(right) {
				path := fmt.Sprintf("%s?%s=%s", ah.RedirectUrl, CAN_NOT_READ, req.URL.Path)
				r.Redirect(path, 302)
			}
		}
	}
}

func (ah *AuthHandler) CheckWriteRights(rights ...string) func(r render.Render, sUser sessionauth.User, req *http.Request) {
	return func(r render.Render, sUser sessionauth.User, req *http.Request) {
		for _, right := range rights {
			user := sUser.(*User)
			if !user.CanWrite(right) {
				path := fmt.Sprintf("%s?%s=%s", ah.RedirectUrl, CAN_NOT_WRITE, req.URL.Path)
				r.Redirect(path, 302)
			}
		}
	}
}

func GetSessions() martini.Handler {
	store := sessions.NewCookieStore([]byte("ALESHA_ХОРОШИЙ"))
	store.Options(sessions.Options{
		MaxAge: 0,
		//Secure:true,
	})
	result := sessions.Sessions("klichat_sessions", store)
	return result
}

var (
	config = configuration.ReadConfig()
	mainDb = db.NewMainDb(config.Main.Database.ConnString, config.Main.Database.Name)
	_sessions = GetSessions()
	_user = sessionauth.SessionUser(AnonymousUserInitializer(mainDb.Users))
)

func FillSession(m *martini.Martini) {
	log.Printf("session: %+v", _sessions)
	m.Use(_sessions)
	m.Use(_user)
}

func NewSessionAuthorisationHandler(mainDb *db.MainDb, ) http.Handler {
	m := martini.New()
	m.Use(NonJsonLogger())
	m.Use(martini.Recovery())
	m.Use(martini.Static("static"))
	m.Use(render.Renderer(render.Options{
		Directory:"templates/auth",
		Extensions: []string{".tmpl", ".html"},
		Charset: "UTF-8",
	}))

	FillSession(m)

	sessionauth.RedirectUrl = "/auth"
	sessionauth.RedirectParam = "from"
	sessionauth.SessionKey = "klichat_session_key"

	flash := Flash{}
	r := martini.NewRouter()

	r.Get("/auth", func(r render.Render, prms martini.Params, req *http.Request) {
		flashMessage, fType := flash.GetMessage()
		query := req.URL.Query()

		result := map[string]interface{}{
			fmt.Sprintf("flash_%v", fType):flashMessage,
			"from": query.Get("from"),
		}
		r.HTML(200, "login", result, render.HTMLOptions{Layout:"base"})
	})

	r.Post("/auth", binding.Bind(User{}), func(session sessions.Session, postedUser User, r render.Render, req *http.Request) {
		log.Printf("session %+v", session)
		userData, err := mainDb.Users.GetUserDataForWeb(postedUser.LoginName, postedUser.Password)
		if err != nil {
			log.Printf("User %+v not found", postedUser)
			flash.SetMessage("К сожалению, пользователь с такими данными не найден", "error")
			r.Redirect(sessionauth.RedirectUrl)
			return
		} else {
			log.Printf("fond user data: %+v", userData)
		}
		user := InitUserObject(userData, mainDb.Users)
		err = sessionauth.AuthenticateSession(session, user)
		if err != nil {
			r.JSON(500, err)
		}
		params := req.URL.Query()
		redirect := params.Get(sessionauth.RedirectParam)
		log.Printf("redirect: %v, params: %+v", redirect, params)
		r.Redirect(redirect)
		return
	})

	m.Action(r.Handle)
	return m
}

//func TestRun(mainDb *db.MainDb) {
//	store := sessions.NewCookieStore([]byte("ALESHA_ХОРОШИЙ"))
//	m := martini.Classic()
//	m.Use(render.Renderer(render.Options{
//		Directory:"templates/auth",
//		Extensions: []string{".tmpl", ".html"},
//		Charset: "UTF-8",
//		IndentJSON: true,
//	}))
//
//	store.Options(sessions.Options{
//		MaxAge: 0,
//	})
//	m.Use(sessions.Sessions("klichat_session", store))
//	m.Use(sessionauth.SessionUser(anonymousUserInitializer(mainDb.Users)))
//	m.Use(martini.Static("static"))
//
//	sessionauth.RedirectUrl = "/"
//	sessionauth.RedirectParam = "from"
//
//	flash := Flash{}
//	ah := AuthHandler{RedirectUrl:"/"}
//	mainDb.Users.AddOrUpdateUserObject(db.UserData{
//		UserId:"1",
//		UserName:"1",
//		Password:utils.PHash("1"),
//		Email:"1@1.ru",
//		Role:"foo",
//		ReadRights:[]string{"some"},
//		WriteRights:[]string{"some"},
//	})
//
//	m.Get("/auth", func(r render.Render) {
//		flashMessage, fType := flash.GetMessage()
//		result := map[string]interface{}{fmt.Sprintf("flash_%v", fType):flashMessage}
//		log.Printf("/ result: %+v", result)
//		r.HTML(200, "login", result, render.HTMLOptions{Layout:"base"})
//	})
//
//	m.Post("/auth", binding.Bind(User{}), func(session sessions.Session, postedUser User, r render.Render, req *http.Request) {
//		userData, err := mainDb.Users.GetUserDataForWeb(postedUser.LoginName, postedUser.Password)
//		if err != nil {
//			log.Printf("User %+v not found", postedUser)
//			flash.SetMessage("К сожалению, пользователь с такими данными не найден", "error")
//			r.Redirect(sessionauth.RedirectUrl)
//			return
//		}
//		user := InitUserObject(userData, mainDb.Users)
//		err = sessionauth.AuthenticateSession(session, user)
//		if err != nil {
//			r.JSON(500, err)
//		}
//		params := req.URL.Query()
//		redirect := params.Get(sessionauth.RedirectParam)
//		r.Redirect(redirect)
//		return
//	})
//	m.Get("/private", sessionauth.LoginRequired, func(r render.Render, user sessionauth.User) {
//		r.HTML(200, "private", map[string]interface{}{"user":user.(*User).Data}, render.HTMLOptions{Layout:"base"})
//	})
//
//	m.Get("/logout", sessionauth.LoginRequired, func(session sessions.Session, user sessionauth.User, r render.Render) {
//		sessionauth.Logout(session, user)
//		r.Redirect("/")
//	})
//	m.Get("/private_some_read", sessionauth.LoginRequired, ah.CheckReadRights("some"), func(r render.Render, user sessionauth.User) {
//		r.HTML(200, "private", map[string]interface{}{"read_rights":user.(*User).Data.ReadRights}, render.HTMLOptions{Layout:"base"})
//	})
//
//	m.Get("/private_some_write", sessionauth.LoginRequired, ah.CheckWriteRights("some"), func(r render.Render, user sessionauth.User) {
//		r.HTML(200, "private", map[string]interface{}{"write_rights":user.(*User).Data.WriteRights}, render.HTMLOptions{Layout:"base"})
//	})
//
//	m.Get("/private_some_role", sessionauth.LoginRequired, ah.CheckIncludeAnyRole("manager", "foo", "bar"), func(r render.Render, user sessionauth.User) {
//		r.HTML(200, "private", map[string]interface{}{"role":user.(*User).Data.Role}, render.HTMLOptions{Layout:"base"})
//	})
//	m.Run()
//}