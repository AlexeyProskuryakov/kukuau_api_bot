package console

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/render"

	"gopkg.in/mgo.v2/bson"

	"time"
	"log"

	c "msngr/configuration"
	d "msngr/db"
	t "msngr/taxi"
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
func Run(config c.Configuration, db *d.DbHandlerMixin) {
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
				state, ok := t.InfinityStatusesName[state_id]
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

	type SourceCount struct {
		Source string `bson:"_id"`
		Count  int `bson:"count"`
		ActiveCount int `bson:"active_count"`
	}

	m.Get("/statistic", func(user auth.User, r render.Render) {
		pipe := db.Orders.Collection.Pipe([]bson.M{
			bson.M{"$match":bson.M{"$exists":bson.M{"source":true}}},
			bson.M{"$group":bson.M{"_id":"$source",
									"count":bson.M{"$sum":1},
									"active_count":bson.M{"$add":[]bson.M{bson.M{"$eq":bson.M{"active":true}}}},
			}},
		})
		result := []SourceCount{}
		err := pipe.All(&result)
		if err != nil {
			log.Printf("error at get aggregate result")
		}

		r.HTML(200, "statistic", result)
	})

	log.Printf("Console will work at addr: %v", config.Main.ConsoleAddr)
	m.RunOnAddr(config.Main.ConsoleAddr)
}
