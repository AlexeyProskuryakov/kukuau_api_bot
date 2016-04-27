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
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"msngr/utils"
)

type CoffeeFunctionData struct {
	Action    string `json:"action"`
	Context   struct {
			  OrderId     string `json:"order_id"`
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
		orderId, err := strconv.ParseInt(cfd.Context.OrderId, 10, 64)
		if err != nil {
			log.Printf("Coffee serv: error at parse int ")
		}
		var result string
		switch cfd.Action {
		case "cancel":
			commands := getCommands(chc, false, false)
			notifier.NotifyTextWithCommands(cfd.Context.UserName, "Ваш заказ отменили!", commands)
			db.Orders.SetActive(orderId, cfd.Context.CompanyName, false)
			result = "Отменено"

		case "start":
			notifier.NotifyText(cfd.Context.UserName, "Ваш заказ в процессе приготовления!")
			result = "В процессе"

		case "end":
			notifier.NotifyText(cfd.Context.UserName, "Ваш заказ готов!")
			db.Orders.SetActive(orderId, cfd.Context.CompanyName, false)
			result = "Окончено"

		case "confirm":
			notifier.NotifyText(cfd.Context.UserName, "Вы уверены что хотите сделать заказ?")
			result = "Подтверждение отправлено"
		}
		db.Messages.UpdateMessageRelatedOrderState(cfd.MessageId, result)

		render.JSON(200, map[string]interface{}{"ok":true, "result":result})
	})
	return m
}
func getActiveOrderMessages(company_id string, db *d.MainDb) ([]d.MessageWrapper, error) {
	orders, err := db.Orders.GetOrdersSort(bson.M{"source":company_id, "is_active":true}, "-when")
	messages := []d.MessageWrapper{}
	for _, order := range orders {
		message, err := db.Messages.GetMessageByRelatedOrder(order.OrderId)
		if err != nil || message == nil {
			log.Printf("Coffee serv: error or can not find message by related order")
			continue
		}
		user, err := db.Users.GetUserById(message.From)
		if err != nil || message == nil {
			log.Printf("Coffee serv: error or can not find user from message")
			continue
		}
		message.FromName = user.GetName()
		messages = append(messages, *message)
	}
	if err != nil {
		log.Printf("Coffee serv error at read orders ")
		return messages, err
	}
	return messages, nil
}

func GetOrdersPageFunctionHandler(start_addr, prefix string, db *d.MainDb, config c.ChatConfig, company_id string) http.Handler {
	m := chat.GetMartini(config.Name, config.CompanyId, prefix, db)
	m.Get(start_addr, func(ren render.Render) {
		log.Printf("Coffee serv getting orders for %v", company_id)
		messages, _ := getActiveOrderMessages(company_id, db)
		ren.HTML(200, "order_page", map[string]interface{}{"order_messages":messages}, render.HTMLOptions{Layout:"base"})
	})
	return m
}

func GetOrdersPageSupplierFunctionHandler(start_addr, prefix string, db *d.MainDb, config c.ChatConfig, company_id string) http.Handler {
	m := chat.GetMartini(config.Name, config.CompanyId, prefix, db)
	m.Post(start_addr, func(ren render.Render, req *http.Request) {
		type ExceptMessages struct {
			Except []string `json:"except"`
		}
		exceptMessages := ExceptMessages{}
		request_body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Printf("Coffee serv: error at read data: %v", err)
			ren.JSON(500, map[string]interface{}{"ok":false, "detail":"can not read request body"})
			return
		}
		err = json.Unmarshal(request_body, &exceptMessages)
		if err != nil {
			log.Printf("Coffee serv: error at unmarshal json: %v", err)
			ren.JSON(500, map[string]interface{}{"ok":false, "detail":fmt.Sprintf("can not unmarshal request body %v \n %s", err, request_body)})
			return
		}
		log.Printf("Coffee serv getting orders for %v, except %+v", company_id, exceptMessages.Except)
		messages, err := getActiveOrderMessages(company_id, db)
		if err != nil {
			ren.JSON(500, map[string]interface{}{"ok":false, "detail":err.Error()})
		}
		result := []d.MessageWrapper{}
		for _, message := range messages {
			if utils.InS(message.MessageID, exceptMessages.Except) {
				continue
			}
			result = append(result, message)
		}
		ren.JSON(200, map[string]interface{}{"ok":true, "order_messages":result})
	})
	return m
}
