package console
import (
	"regexp"
	"net/http"
	"strings"
	"encoding/json"
	"log"
	"time"
	"io/ioutil"
	"fmt"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/render"

	"gopkg.in/mgo.v2/bson"

	c "msngr/configuration"
	d "msngr/db"
	t "msngr/taxi"
	"msngr/structs"

)

// The regex to check for the requested format (allows an optional trailing
// slash).
var rxExt = regexp.MustCompile(`(\.(?:xml|text|json))\/?$`)

// MapEncoder intercepts the request's URL, detects the requested format,
// and injects the correct encoder dependency for this request. It rewrites
// the URL to remove the format extension, so that routes can be defined
// without it.
func MapEncoder(c martini.Context, w http.ResponseWriter, r *http.Request) {
	// Get the format extension
	matches := rxExt.FindStringSubmatch(r.URL.Path)
	ft := ".json"
	if len(matches) > 1 {
		// Rewrite the URL without the format extension
		l := len(r.URL.Path) - len(matches[1])
		if strings.HasSuffix(r.URL.Path, "/") {
			l--
		}
		r.URL.Path = r.URL.Path[:l]
		ft = matches[1]
	}
	// Inject the requested encoder
	log.Printf("ft: %s", ft)
	switch ft {
	default:
		c.MapTo(jsonEncoder{}, (*Encoder)(nil))
		w.Header().Set("Content-Type", "application/json")
	}
}

type Encoder interface {
	Encode(v ...interface{}) (string, error)
}
type jsonEncoder struct{}

// jsonEncoder is an Encoder that produces JSON-formatted responses.
func (_ jsonEncoder) Encode(v ...interface{}) (string, error) {
	var data interface{} = v
	if v == nil {
		// So that empty results produces `[]` and not `null`
		data = []interface{}{}
	} else if len(v) == 1 {
		data = v[0]
	}
	b, err := json.Marshal(data)
	return string(b), err
}


func Run(config c.Configuration, db *d.MainDb, cs c.ConfigStorage) {
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

	m.Use(MapEncoder)

	m.Get("/", func(user auth.User, r render.Render) {
		uptime := time.Now().Sub(start_time)
		result_map := map[string]interface{}{"info":IndexInfo{
			UserName:string(user),
			CurrentTime:time.Now(),
			UpTime:uptime,
		}}
		r.HTML(200, "index", result_map)
	})

	m.Get("/console", func(r render.Render) {
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
				if user_name != "" {
					logged_users = append(logged_users, user_name)
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
		Source      string `bson:"_id"`
		Count       int `bson:"count"`
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


	m.Post("/configuration", func(request *http.Request, render render.Render) {
		input, err := ioutil.ReadAll(request.Body)
		defer request.Body.Close()
		if err != nil {
			render.JSON(500, map[string]interface{}{"Error":fmt.Sprintf("Can not read request body. Because: %v",err)})
			return
		}
		type CommandInfo struct{
			Provider string `json:"provider"`
			Name string `json:"name"`
			Command structs.OutCommand `json:"command"`
		}
		commandInfo := CommandInfo{}
		err = json.Unmarshal(input, &commandInfo)
		if err != nil {
			render.JSON(500, map[string]interface{}{"Error":fmt.Sprintf("Can not unmarshall input to command info. Because: %v", err)})
			return
		}
		cs.SaveCommand(commandInfo.Provider, commandInfo.Name, commandInfo.Command)
		render.JSON(200, map[string]interface{}{"OK":true})
	})

	m.Get("/configuration", func (request *http.Request){

	})

	log.Printf("Console will work at addr: %v", config.Main.ConsoleAddr)
	m.RunOnAddr(config.Main.ConsoleAddr)
}