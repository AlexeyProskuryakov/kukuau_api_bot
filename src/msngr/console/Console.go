package console

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/render"
	"fmt"
	"time"
	d "msngr/db"
	msngr "msngr"
	t "msngr/taxi"
	i "msngr/taxi/infinity"
	"log"
	"gopkg.in/mgo.v2/bson"
)

var users = map[string]string{
	"alesha":"sederfes100500",
	"leha":"qwerty100500",
}


type IndexInfo struct {
	UserName    string
	CurrentTime time.Time
	UpTime      time.Duration
}

type ConsoleInfo struct {
	TaxiOrdersCount int
	TaxiUserOrders  map[string]string

	ShopUsersCount  int
	ShopLoginUsers  []string

}
func Run(config msngr.Configuration, db *d.DbHandlerMixin) {
	m := martini.Classic()

	martini.Env = martini.Dev
	start_time := time.Now()

	m.Use(render.Renderer(render.Options{
		Layout: "layout",
		Extensions: []string{".tmpl", ".html"},
		Charset: "UTF-8",
		IndentJSON: true,
		IndentXML: true,
	}))

	m.Use(auth.BasicFunc(func(username, password string) bool {
		pwd, ok := users[username]
		return ok && pwd == password
	}))

	m.Get("/", func(user auth.User, r render.Render) {
		uptime := time.Now().Sub(start_time)
		result_map := map[string]interface{}{"info":IndexInfo{
			UserName:string(user),
			CurrentTime:time.Now(),
			UpTime:uptime,
		}}
		r.HTML(200, "index", result_map)
	})

	m.Get("/console", func(user auth.User, r render.Render) {
		user_order_statuses := map[string]string{}
		for order_id, _ := range t.PreviousStates {
			res, err := db.Orders.GetBy(bson.M{"order_id":order_id})
			if err != nil {
				continue
			}
			if res != nil {
				order := res[0]
				state_id := order.OrderState
				state, ok := i.StatusesMap[state_id]
				if ok {
					user_order_statuses[order.Whom] = state
				}
			}
		}
		logged_users := []string{}
		logged_users_info, err := db.Users.GetBy(bson.M{"user_state":d.LOGIN})
		if err == nil {
			for _, user_info := range *logged_users_info {
				user_name := user_info.UserName
				if user_name != nil {
					logged_users = append(logged_users, *user_name)
				}
			}
		}
		ci := ConsoleInfo{
			TaxiOrdersCount:db.Orders.Count(),
			TaxiUserOrders:user_order_statuses,

			ShopUsersCount:db.Users.Count(),
			ShopLoginUsers:logged_users,
		}

		result_map := map[string]interface{}{"info":ci}
		r.HTML(200, "console", result_map)
	})

	m.Get("/statistic", func(user auth.User) (int, string) {
		return 200, fmt.Sprintf("This is statistic! %v", user)
	})
	log.Printf("will work at addr: %v", config.Main.ConsoleAddr)
	m.RunOnAddr(config.Main.ConsoleAddr)
}
