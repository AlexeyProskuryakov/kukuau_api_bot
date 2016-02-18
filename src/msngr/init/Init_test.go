package init

import (
	c "msngr/configuration"
	m "msngr"
	n "msngr/notify"
	d "msngr/db"
	u "msngr/utils"
	i "msngr/taxi/infinity"
	tu "msngr/test"

	"msngr/taxi"
	"msngr/taxi/geo"
	"msngr/structs"

	"testing"
	"net/http"
	"fmt"
)

func TestTaxiInfinityFail(t *testing.T) {
	conf := c.ReadTestConfigInRecursive()
	db := d.NewMainDb(conf.Main.Database.ConnString, conf.Main.Database.Name)
	taxi_conf := conf.Taxis["fake"]

	t.Logf("taxi api configuration for %+v:\n%v", taxi_conf.Name, taxi_conf.Api)
	external := i.GetTestInfAPI(taxi_conf.Api)

	external_api := external.(taxi.TaxiInterface)
	external_address_supplier := external.(taxi.AddressSupplier)

	apiMixin := taxi.ExternalApiMixin{API: external_api}

	carsCache := taxi.NewCarsCache(external_api)
	notifier := n.NewNotifier(conf.Main.CallbackAddr, taxi_conf.Key, db)

	address_handler, address_supplier := GetAddressInstruments(conf, taxi_conf.Name, external_address_supplier)

	botContext := taxi.FormTaxiBotContext(&apiMixin, db, taxi_conf, address_handler, carsCache)
	controller := m.FormBotController(botContext, db)

	t.Logf("Was create bot context: %+v\n", botContext)
	http.HandleFunc(fmt.Sprintf("/taxi/%v", taxi_conf.Name), controller)

	go func() {
		taxiContext := taxi.TaxiContext{API: external_api, DataBase: db, Cars: carsCache, Notifier: notifier}
		t.Logf("Will start order watcher for [%v]", botContext.Name)
		taxi.TaxiOrderWatch(&taxiContext, botContext)
	}()

	http.HandleFunc(fmt.Sprintf("/taxi/%v/streets", taxi_conf.Name), func(w http.ResponseWriter, r *http.Request) {
		geo.StreetsSearchController(w, r, address_supplier)
	})

	go func() {
		server_address := fmt.Sprintf(":%v", conf.Main.Port)
		t.Logf("\nStart listen and serving at: %v\n", server_address)
		server := &http.Server{
			Addr: server_address,
		}
		server.ListenAndServe()
	}()

	u.After(external_api.IsConnected, func() {
		inf_api := external.(*i.InfinityAPI)
		for i, _ := range inf_api.ConnStrings {
			inf_api.ConnStrings[i] += "w"
		}
	})

	_, err := u.GET(taxi_conf.DictUrl, &map[string]string{"q":"лесос"})
	if err != nil {
		t.Error("Error at getting street %v", err)
	}

	out_res, err := tu.POST(fmt.Sprintf("http://localhost:%v/taxi/%v", conf.Main.Port, taxi_conf.Name), &structs.InPkg{
		Message:&structs.InMessage{
			ID: u.GenId(),
			Commands:&[]structs.InCommand{
				structs.InCommand{Action:"information", Title:"information"},
			},
		},
		From:"test user",
		UserData:&structs.InUserData{Name:"TEST"},
	})

	if err != nil {
		t.Errorf("Error at unmarshal result info %v", err)
	}

	if out_res == nil {
		t.Error("Out info result is nil!")
	}else {
		if out_res.Message.Type != "chat" {
			t.Errorf("Out message type != chat, but == %v", out_res.Message.Type)
		}
		if out_res.Message.Body != taxi_conf.Information.Text {
			t.Errorf("Out message body != info in config, but == %v", out_res.Message.Body)
		}
	}

}