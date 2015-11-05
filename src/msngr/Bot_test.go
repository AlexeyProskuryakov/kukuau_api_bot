package msngr

import (
	"log"

	s "msngr/structs"
	d "msngr/db"
	sh "msngr/shop"
	"flag"

	"fmt"
	"net/http"
	"io/ioutil"
	"bytes"
	"testing"
)

func send_post(fn, url string) {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Panic(err)
	}
	body := bytes.NewBuffer(data)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Panic(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err)
		return
	}
	if resp != nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Panic(err)
		} else {
			log.Printf("<<< %+v", string(body))

		}
		defer resp.Body.Close()
	}
}


func TestBot(t *testing.T) {
	conf := ReadConfig()
	var test = flag.Bool("test", false, "go in test use?")
	flag.Parse()
	DEBUG = true
	d.DELETE_DB = *test
	log.Printf("Is test? [%+v] Will delete db? [%+v]", *test, d.DELETE_DB)
	if d.DELETE_DB {
		log.Println("!start at test mode!")
		conf.Database.Name = conf.Database.Name + "_test"
	}
	db := d.NewDbHandler(conf.Database.ConnString, conf.Database.Name)

	for _, shop_conf := range conf.Shops {
		bot_context := sh.FormShopCommands(db, &shop_conf)
		shop_controller := FormBotController(bot_context)
		shop_route := fmt.Sprintf("/shop/%v", shop_conf.Name)
		log.Println("will wait requests at :", shop_route)
		http.HandleFunc(shop_route, shop_controller)
	}

	server_address := fmt.Sprintf(":%v", conf.Main.Port)
	s.StartAfter(db.Check, func() {
		log.Printf("\nStart listen and serving at: %v\n", server_address)
		server := &http.Server{
			Addr: server_address,
		}

		log.Fatal(server.ListenAndServe())
	})

	log.Printf("will send requests....")

	addr := fmt.Sprintf("http://localhost:%v/shop/test_shop", conf.Main.Port)
	send_post("test_res/shop_balance_ok.json", addr)
	send_post("test_res/shop_balance_error.json", addr)
	send_post("test_res/request_commands.json", addr)
}

func main() {
	t := testing.T{}

	TestBot(&t)
	log.Printf("%+v", t)
}