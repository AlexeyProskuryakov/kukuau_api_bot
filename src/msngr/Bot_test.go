package msngr

import (
	"log"

	s "msngr/structs"
	d "msngr/db"
	sh "msngr/shop"
	t "msngr/test"
	c "msngr/configuration"

	"flag"
	"fmt"
	"net/http"
	"io/ioutil"
	"bytes"
	"testing"

)




func send_post(fn, url string) []byte {
	result := []byte{}
	ffn := t.GetTestFileName(fn)
	if ffn == nil {
		return result
	}
	data, err := ioutil.ReadFile(*ffn)
	if err != nil {
		log.Panic(err)
	}
	body := bytes.NewBuffer(data)


	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Panic(err)
		return result
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err)
		return result
	}
	if resp != nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Panic(err)
		} else {
			log.Printf("<<< %+v", string(body))
			return body
		}
		defer resp.Body.Close()
	}
	return result
}


func TestBot(t *testing.T) {
	conf := c.ReadConfig()
	var test = flag.Bool("test", false, "go in test use?")
	flag.Parse()
	DEBUG = true
	d.DELETE_DB = *test
	log.Printf("Is test? [%+v] Will delete db? [%+v]", *test, d.DELETE_DB)
	if d.DELETE_DB {
		log.Println("!start at test mode!")
		conf.Main.Database.Name = conf.Main.Database.Name + "_test"
	}
	db := d.NewMainDb(conf.Main.Database.ConnString, conf.Main.Database.Name)

	for _, shop_conf := range conf.Shops {
		bot_context := sh.FormShopCommands(db, &shop_conf)
		shop_controller := FormBotController(bot_context)
		shop_route := fmt.Sprintf("/shop/%v", shop_conf.Name)
		log.Println("will wait requests at :", shop_route)
		http.HandleFunc(shop_route, shop_controller)
	}

	server_address := fmt.Sprintf(":%v", conf.Main.Port)
	s.StartAfter(func() (string, bool) {
		return "", db.Check()
	}, func() {
		log.Printf("\nStart listen and serving at: %v\n", server_address)
		server := &http.Server{
			Addr: server_address,
		}

		log.Fatal(server.ListenAndServe())
	})

	log.Printf("will send requests....")

	addr := fmt.Sprintf("http://localhost:%v/shop/test_shop", conf.Main.Port)
	send_post("shop_balance_ok.json", addr)
	send_post("shop_balance_error.json", addr)
	send_post("request_commands.json", addr)
	t.Log("test ended...")
}