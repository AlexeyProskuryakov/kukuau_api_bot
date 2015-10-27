package main
import (
	"net/http"
	"log"
	"io/ioutil"
	"encoding/json"
	s "msngr/structs"
	t "msngr/taxi"
	i "msngr/taxi/infinity"
	m "msngr"
	d "msngr/db"
	n "msngr/notify"
	sh "msngr/shop"
	"time"
	"fmt"
)
func readIn(in_jsoned string) *s.InPkg {
	data, err := ioutil.ReadFile(in_jsoned)
	if err != nil {
		log.Printf("error at read: %q \n", err)
	}
	in := s.InPkg{}
	log.Printf("READ IN FROM FILE:\n%+v", string(data))
	err = json.Unmarshal(data, &in)
	if err != nil {
		log.Printf("error at unmarshal: %+v \n", err)
	}
	return &in
}

func serve_notifications(out chan s.OutPkg) {
	addr := ":9876"

	http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		log.Printf("!!!TEST notification arrivied: %+v", string(body))
		var pkg s.OutPkg
		err = json.Unmarshal(body, &pkg)
		if err != nil {
			log.Printf("ths err: %+v", err)
		}
		out <- pkg
	})

	serv := &http.Server{
		Addr: addr,
	}

	log.Fatal(serv.ListenAndServe())
}


func test_taxi() {
	conf := m.ReadConfig()

	d.DELETE_DB = true
	if d.DELETE_DB {
		log.Println("!start at test mode!")
		conf.Database.Name = conf.Database.Name + "_test"
	}
	//for reading packages from notification
	notif_chan := make(chan s.OutPkg)
	go serve_notifications(notif_chan)

	for _, taxi_conf := range conf.Taxis {
		if taxi_conf.Name == "fake" {
			log.Println(taxi_conf)
			external_api := t.GetFakeInfinityAPI(taxi_conf.Api)
			external_address_supplier := i.GetInfinityAddressSupplier(taxi_conf.Api)

			apiMixin := t.ExternalApiMixin{API: external_api}

			db := d.NewDbHandler(conf.Database.ConnString, conf.Database.Name)
			for !db.IsConnected() {
				log.Println("wait while db is connected...")
				time.Sleep(time.Second)
			}
			carsCache := t.NewCarsCache(external_api)
			notifier := n.NewNotifier(conf.Main.CallbackAddr, taxi_conf.Key)

			gah := t.NewGoogleAddressHandler(conf.Main.GoogleKey, taxi_conf.GeoOrbit, external_address_supplier)


			botContext := t.FormTaxiBotContext(&apiMixin, db, taxi_conf, gah)
			request_commands := botContext.Request_commands
			message_commands := botContext.Message_commands
			taxiContext := t.TaxiContext{API:external_api, DataBase:db, Cars:carsCache, Notifier:notifier}

			go t.TaxiOrderWatch(&taxiContext, botContext)


			streets_address := fmt.Sprintf("/taxi/%v/streets", taxi_conf.Name)

			http.HandleFunc(streets_address, func(w http.ResponseWriter, r *http.Request) {
				t.StreetsSearchController(w, r, gah)
			})

			server_address := fmt.Sprintf(":%v", conf.Main.Port)

			server := &http.Server{
				Addr: server_address,
			}
			test_url := "http://localhost" + server_address + streets_address
			log.Printf("start server... send tests to: %v?=", test_url)

			go server.ListenAndServe()


			//scenario commands
			in := s.InPkg{From:"TEST", UserData:&s.InUserData{Phone:"TEST123"}, Request:&s.InRequest{ID:"1234", Type:"get"}}
			in.Request.Query.Action = "COMMANDS"

			request_result := request_commands["commands"].ProcessRequest(&in)
			log.Println("commands: ERROR?: ", request_result.Error)

			//scenario new order
			//			in_str := "{\"message\":{\"commands\":[{\"form\":{\"fields\":[{\"data\":{\"value\":\"{\"ID\":5009785215,\"IDParent\":5009776746,\"Name\":\"\u041a\u043e\u043c\u043c\u0443\u043d\u0438\u0441\u0442\u0438\u0447\u0435\u0441\u043a\u0430\u044f\",\"ShortName\":\"\u0443\u043b\",\"ItemType\":5,\"FullName\":\"\u043e\u0431\u043b \u041d\u043e\u0432\u043e\u0441\u0438\u0431\u0438\u0440\u0441\u043a\u0430\u044f \u0440-\u043d \u041c\u043e\u0448\u043a\u043e\u0432\u0441\u043a\u0438\u0439 \u0441 \u0421\u043e\u043a\u0443\u0440 \u0443\u043b \u041a\u043e\u043c\u043c\u0443\u043d\u0438\u0441\u0442\u0438\u0447\u0435\u0441\u043a\u0430\u044f\",\"IDRegion\":5009755359,\"IDDistrict\":5009776716,\"IDCity\":0,\"IDPlace\":5009776746,\"Region\":\"\u043e\u0431\u043b \u041d\u043e\u0432\u043e\u0441\u0438\u0431\u0438\u0440\u0441\u043a\u0430\u044f\",\"District\":\"\u0440-\u043d \u041c\u043e\u0448\u043a\u043e\u0432\u0441\u043a\u0438\u0439\",\"City\":\"\",\"Place\":\"\u0441 \u0421\u043e\u043a\u0443\u0440\"}\",\"text\":\"\u041a\u043e\u043c\u043c\u0443\u043d\u0438\u0441\u0442\u0438\u0447\u0435\u0441\u043a\u0430\u044f \u0443\u043b\"},\"type\":\"dict\",\"name\":\"street_from\"},{\"type\":\"number\",\"name\":\"entrance\"},{\"data\":{\"value\":\"{\"ID\":5009782521,\"IDParent\":5009776292,\"Name\":\"\u0420\u0430\u0434\u0443\u0436\u043d\u0430\u044f\",\"ShortName\":\"\u0443\u043b\",\"ItemType\":5,\"FullName\":\"\u043e\u0431\u043b \u041d\u043e\u0432\u043e\u0441\u0438\u0431\u0438\u0440\u0441\u043a\u0430\u044f \u0440-\u043d \u041a\u043e\u043b\u044b\u0432\u0430\u043d\u0441\u043a\u0438\u0439 \u0440\u043f \u041a\u043e\u043b\u044b\u0432\u0430\u043d\u044c \u0443\u043b \u0420\u0430\u0434\u0443\u0436\u043d\u0430\u044f\",\"IDRegion\":5009755359,\"IDDistrict\":5009776291,\"IDCity\":0,\"IDPlace\":5009776292,\"Region\":\"\u043e\u0431\u043b \u041d\u043e\u0432\u043e\u0441\u0438\u0431\u0438\u0440\u0441\u043a\u0430\u044f\",\"District\":\"\u0440-\u043d \u041a\u043e\u043b\u044b\u0432\u0430\u043d\u0441\u043a\u0438\u0439\",\"City\":\"\",\"Place\":\"\u0440\u043f \u041a\u043e\u043b\u044b\u0432\u0430\u043d\u044c\"}\",\"text\":\"\u0420\u0430\u0434\u0443\u0436\u043d\u0430\u044f \u0443\u043b\"},\"type\":\"dict\",\"name\":\"street_to\"},{\"data\":{\"value\":\"2\"},\"type\":\"text\",\"name\":\"house_to\"},{\"data\":{\"value\":\"1\"},\"type\":\"text\",\"name\":\"house_from\"}]},\"repeated\":\"true\",\"fixed\":\"false\",\"action\":\"new_order\"}],\"thread\":\"2c474723-ff4d-4fad-a31f-18e8063bdff2\",\"body\":\"\u041e\u0442\u043a\u0443\u0434\u0430: \u041a\u043e\u043c\u043c\u0443\u043d\u0438\u0441\u0442\u0438\u0447\u0435\u0441\u043a\u0430\u044f \u0443\u043b, 1, \u043d\u0435\u0442. \u041a\u0443\u0434\u0430: \u0420\u0430\u0434\u0443\u0436\u043d\u0430\u044f \u0443\u043b, 2.\",\"type\":\"chat\",\"id\":\"T6rqU-93\"},\"from\":\"113c32ee9f65586882284425376e5d9d\",\"userdata\":{\"phone\":\"79231243359\"}}"
			read_in := readIn("test_res/new_order_not_here.json")
			message_result := message_commands["new_order"].ProcessMessage(read_in)
			log.Println("new order ERROR == ", message_result.Error)


			read_in = readIn("test_res/new_order_ok.json")
			message_result = message_commands["new_order"].ProcessMessage(read_in)
			log.Println("new order ERROR?:", message_result.Error)

			states := taxi_conf.Api.Fake.SendedStates
			counter := 0


			for pkg := range notif_chan {
				log.Printf("\n\nEXCEPTED PACKAGE: [%v]\n %#v \nstate: [%v]\n", counter, pkg, states[counter])
				counter += 1
				if counter == 3 { //because must be it!
					break
				}
			}

			log.Printf("will sleep 5 seconds while notifications will sended all...")
			time.Sleep(5 * time.Second)
			in = s.InPkg{From:"TEST", UserData:&s.InUserData{Phone:"TEST123"}, Request:&s.InRequest{ID:"1234", Type:"get"}}
			in.Request.Query.Action = "COMMANDS"

			request_result1 := request_commands["commands"].ProcessRequest(&in)
			log.Println("BEFORE FDBCK commands: ERROR?: ", request_result1.Error, "commands: \n", request_result1.Commands)

			read_in = readIn("test_res/feedback.json")
			message_result = message_commands["feedback"].ProcessMessage(read_in)
			log.Printf("feedback error?: %v\n body: %v, \n commands:%v", message_result.Error, message_result.Body, message_result.Commands)

			in = s.InPkg{From:"TEST", UserData:&s.InUserData{Phone:"TESTPHONE"}, Request:&s.InRequest{ID:"1234", Type:"get"}}
			in.Request.Query.Action = "COMMANDS"

			request_result2 := request_commands["commands"].ProcessRequest(&in)
			log.Println("AFTER commands: ERROR?: ", request_result2.Error, "commands: \n", request_result2.Commands)


		}
	}

}

func test_shops() {
	conf := m.ReadConfig()

	d.DELETE_DB = true
	if d.DELETE_DB {
		log.Println("!start at test mode!")
		conf.Database.Name = conf.Database.Name + "_test"
	}

	for _, shop_conf := range conf.Shops {
		db := d.NewDbHandler(conf.Database.ConnString, conf.Database.Name)
		bot_context := sh.FormShopCommands(db, &shop_conf)
		request_commands := bot_context.Request_commands

		in := s.InPkg{From:"TEST", UserData:&s.InUserData{Phone:"TEST123"}, Request:&s.InRequest{ID:"1234", Type:"get"}}
		in.Request.Query.Action = "COMMANDS"

		request_result := request_commands["commands"].ProcessRequest(&in)
		log.Println("commands: ERROR?: ", request_result.Error)

	}
}
func main() {
	test_taxi()
	//	test_shops()
}

