package main

import (
	"fmt"
	"log"
	m "msngr"
	ia "msngr/infinity"
	"net/http"
	"time"
)

func form_taxi_commands(im ia.InfinityMixin, db m.DbHandlerMixin) (map[string]m.RequestCommandProcessor, map[string]m.MessageCommandProcessor) {
	var TaxiRequestCommands = map[string]m.RequestCommandProcessor{
		"commands": m.TaxiCommandsProcessor{DbHandlerMixin: db},
	}

	var TaxiMessageCommands = map[string]m.MessageCommandProcessor{
		"information":      m.TaxiInformationProcessor{},
		"new_order":        m.TaxiNewOrderProcessor{InfinityMixin: im, DbHandlerMixin: db},
		"cancel_order":     m.TaxiCancelOrderProcessor{InfinityMixin: im, DbHandlerMixin: db},
		"calculate_price":  m.TaxiCalculatePriceProcessor{InfinityMixin: im},
		"feedback":         m.TaxiFeedbackProcessor{InfinityMixin: im, DbHandlerMixin: db},
		"write_dispatcher": m.SupportMessageProcessor{},
	}
	return TaxiRequestCommands, TaxiMessageCommands
}

func form_shop_commands(db m.DbHandlerMixin) (map[string]m.RequestCommandProcessor, map[string]m.MessageCommandProcessor) {
	var ShopRequestCommands = map[string]m.RequestCommandProcessor{
		"commands": m.ShopCommandsProcessor{DbHandlerMixin: db},
	}

	var ShopMessageCommands = map[string]m.MessageCommandProcessor{
		"information":     m.ShopInformationProcessor{},
		"authorise":       m.ShopAuthoriseProcessor{DbHandlerMixin: db},
		"orders_state":    m.ShopOrderStateProcessor{DbHandlerMixin: db},
		"support_message": m.SupportMessageProcessor{},
		"log_out":         m.ShopLogOutMessageProcessor{DbHandlerMixin: db},
	}
	return ShopRequestCommands, ShopMessageCommands
}

func order_watch(db m.DbHandlerMixin, im ia.InfinityMixin, carsCache *ia.CarsCache, n *m.Notifier) {
	for {
		api_orders := im.API.Orders()
		// log.Printf("OW api have %v orders", len(api_orders))
		for _, order := range api_orders {
			order_state := db.Orders.GetState(order.ID)
			// log.Printf("state of %+v is: %v\n", order, order_state)

			if order_state == -1 {
				log.Printf("order %+v is not present in system :(\n", order)
				continue
			}
			if order.State != order_state {
				log.Printf("state of %v will persist", order)
				db.Orders.SetState(order.ID, order.State, &order)
				n.Notify(m.FormNotification(order.ID, order.State, db, carsCache))
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	conf := m.ReadConfig()

	url := &m.DictUrl
	*url = conf.Main.DictUrl
	infApi := ia.GetInfinityAPI(conf.Infinity, conf.Main.Test)
	im := ia.InfinityMixin{API: infApi}

	db := m.NewDbHandler(conf.Database.ConnString, conf.Database.Name)
	db.Users.SetUserPassword("test", "123")

	taxi_controller := m.FormBotController(form_taxi_commands(im, *db))
	shop_controller := m.FormBotController(form_shop_commands(*db))

	http.HandleFunc("/taxi", taxi_controller)
	http.HandleFunc("/shop", shop_controller)

	realInfApi := ia.GetRealInfinityAPI(conf.Infinity)
	http.HandleFunc("/_streets", func(w http.ResponseWriter, r *http.Request) {
		ia.StreetsSearchController(w, r, realInfApi)
	})

	addr := fmt.Sprintf(":%v", conf.Main.Port)

	log.Printf("\nStart listen and serving at: %v\n", addr)
	serv := &http.Server{
		Addr: addr,
	}

	//////////////////////////////////////////////////////////////////
	n_taxi := m.NewNotifier(conf.Main.CallbackAddr, conf.Main.TaxiKey)
	carsCache := ia.NewCarsCache(realInfApi)
	go order_watch(*db, im, carsCache, n_taxi)

	log.Fatal(serv.ListenAndServe())
}
