package coffee

import (
	"net/http"
	"github.com/martini-contrib/render"
	"io/ioutil"
	"encoding/json"
	"fmt"
	ntf "msngr/notify"
	d "msngr/db"
	c "msngr/configuration"
	"msngr/chat"
	"log"
)

type CoffeeFunctionData struct {
	Action    string `json:"action"`
	Context   struct {
			  OrderId     int64 `json:"order_id"`
			  UserName    string `json:"user_name"`
			  CompanyName string `json:"company_name"`
		  } `json:"context"`
	MessageId string `json:"message_id"`
}

func GetMessageAdditionalFunctionsHandler(start_addr string, notifier *ntf.Notifier, db *d.MainDb, config c.ChatConfig, chc *CoffeeHouseConfiguration) http.Handler {
	m := chat.GetMartini(config.Name, config.CompanyId, start_addr, db)
	m.Post(start_addr, func(render render.Render, req *http.Request) {
		cfd := CoffeeFunctionData{}
		request_body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Printf("Coffee serv: error at read data: %v", err)
			render.JSON(500, map[string]interface{}{"ok":false, "detail":"can not read request body"})
			return
		}
		err = json.Unmarshal(request_body, &cfd)
		if err != nil {
			log.Printf("Coffee serv: error at unmarshal json: %v", err)
			render.JSON(500, map[string]interface{}{"ok":false, "detail":fmt.Sprintf("can not unmarshal request body %v \n %s", err, request_body)})
			return
		}
		log.Printf("Coffee serv: message function: %+v", cfd)

		var result string

		switch cfd.Action {
		case "cancel":
			commands := getCommands(chc, false, false)
			notifier.NotifyTextWithCommands(cfd.Context.UserName, "Ваш заказ отменили!", commands)
			db.Orders.SetActive(cfd.Context.OrderId, cfd.Context.CompanyName, false)
			result = "Отменено"

		case "start":
			notifier.NotifyText(cfd.Context.UserName, "Ваш заказ в процессе приготовления!")
			result = "В процессе"

		case "end":
			notifier.NotifyText(cfd.Context.UserName, "Ваш заказ готов!")
			db.Orders.SetActive(cfd.Context.OrderId, cfd.Context.CompanyName, false)
			result = "Оконченно"

		}
		db.Messages.UpdateMessageRelatedOrderState(cfd.MessageId, result)

		render.JSON(200, map[string]interface{}{"ok":true, "result":result})

	})
	return m
}