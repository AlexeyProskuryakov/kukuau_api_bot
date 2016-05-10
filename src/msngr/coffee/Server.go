package coffee

import (
	"net/http"
	"github.com/martini-contrib/render"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"log"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"msngr/utils"
	"github.com/go-martini/martini"
	"strings"
	"sort"
	"msngr/users"
	"time"

	ntf "msngr/notify"
	d "msngr/db"
	c "msngr/configuration"
	"msngr/chat"
	"msngr/web"
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

func GetMartini(cName, cId, start_addr string, db *d.MainDb) *martini.ClassicMartini {
	m := martini.Classic()
	m.Use(web.GetRenderer(cName, cId, start_addr, "coffee"))
	m.MapTo(db, (*d.DB)(nil))
	return m
}

func GetMessageAdditionalFunctionsHandler(start_addr string, notifier *ntf.Notifier, db *d.MainDb, config c.ChatConfig, chc *CoffeeHouseConfiguration) http.Handler {
	m := GetMartini(config.Name, config.CompanyId, start_addr, db)
	m.Post(start_addr,
		web.LoginRequired,
		web.AutHandler.CheckReadRights(config.CompanyId),
		web.AutHandler.CheckWriteRights(config.CompanyId),
		func(render render.Render, req *http.Request) {
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

			db.Messages.SetMessageFunctionUsed(cfd.MessageId, cfd.Action)
			db.Messages.UpdateMessageRelatedOrderState(cfd.MessageId, result)

			found, err := db.Orders.GetByOwner(cfd.Context.UserName, cfd.Context.CompanyName, true)
			render.JSON(200, map[string]interface{}{"ok":true, "result":result, "user_have_active_orders":found != nil})
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
	m := GetMartini(config.Name, config.CompanyId, prefix, db)
	m.Get(start_addr,
		web.LoginRequired,
		web.AutHandler.CheckReadRights(config.CompanyId),
		web.AutHandler.CheckWriteRights(config.CompanyId),
		func(ren render.Render) {
			log.Printf("Coffee serv getting orders for %v", company_id)
			messages, _ := getActiveOrderMessages(company_id, db)
			ren.HTML(200, "order_page", map[string]interface{}{"order_messages":messages}, render.HTMLOptions{Layout:"base"})
		})
	return m
}

func GetOrdersPageSupplierFunctionHandler(start_addr, prefix string, db *d.MainDb, config c.ChatConfig, company_id string) http.Handler {
	m := GetMartini(config.Name, config.CompanyId, prefix, db)
	m.Post(start_addr,
		web.LoginRequired,
		web.AutHandler.CheckReadRights(config.CompanyId),
		web.AutHandler.CheckWriteRights(config.CompanyId),
		func(ren render.Render, req *http.Request) {
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

func GetContacts(db *d.MainDb, to_name string) ([]users.Contact, error) {
	resp := []users.Contact{}
	err := db.Messages.MessagesCollection.Pipe([]bson.M{
		bson.M{"$match":bson.M{"to":to_name, "is_deleted":false}},
		bson.M{"$group": bson.M{"_id":"$from", "unread_count":bson.M{"$sum":"$unread"}, "name":bson.M{"$first":"$from"}, "time":bson.M{"$max":"$time_stamp"}}}}).All(&resp)
	if err != nil {
		return resp, err
	}
	result := []users.Contact{}
	for i, cnt := range resp {
		user, _ := db.Users.GetUserById(cnt.ID)
		if user != nil {
			if user.ShowedName != "" {
				resp[i].Name = user.ShowedName
			} else {
				resp[i].Name = user.UserName
			}
			resp[i].Phone = user.Phone
			last_order, _ := db.Orders.GetByOwnerLast(resp[i].ID, to_name)
			if last_order != nil {
				resp[i].HaveActiveOrder = last_order.Active
			}
			result = append(result, resp[i])
		}
	}
	sort.Sort(users.ByContactsLastMessageTime(result))
	return result, nil
}
func GetChatConfigHandler(start_addr, prefix string, db *d.MainDb, config c.ChatConfig) http.Handler {
	m := GetMartini(config.Name, config.CompanyId, prefix, db)
	m.Get(start_addr,
		web.LoginRequired,
		web.AutHandler.CheckWriteRights(config.CompanyId),
		func(ren render.Render, req *http.Request) {
			ren.HTML(200, "config", map[string]interface{}{}, render.HTMLOptions{Layout:"base"})
		})
	return m
}

func GetChatMainHandler(start_addr string, notifier *ntf.Notifier, db *d.MainDb, config c.ChatConfig) http.Handler {
	m := GetMartini(config.Name, config.CompanyId, start_addr, db)
	m.Get(start_addr,
		web.LoginRequired,
		web.AutHandler.CheckReadRights(config.CompanyId),
		web.AutHandler.CheckWriteRights(config.CompanyId),
		func(r render.Render, params martini.Params, req *http.Request) {
			var with string
			result_data := map[string]interface{}{}
			query := req.URL.Query()
			for key, value := range query {
				if key == "with" && len(value) > 0 {
					with = value[0]
					log.Printf("CONSOLE CHAT: [with] == [%v]", with)
					break
				}
			}
			if with == "" {
				with = chat.ALL
			}
			db.Messages.SetMessagesRead(with)

			var messages []d.MessageWrapper
			if with != chat.ALL {
				user, _ := db.Users.GetUserById(with)
				if user != nil {
					messages, _ = db.Messages.GetMessages(bson.M{
						"$or":[]bson.M{
							bson.M{"from":user.UserId, "to":config.CompanyId},
							bson.M{"to":user.UserId, "from":config.CompanyId},
						},
					})
					for i, message := range messages {
						if message.From == user.UserId {
							messages[i].From = user.GetName()
						} else if message.To == user.UserId {
							messages[i].To = user.GetName()
						}
					}
				}
			}

			if strings.Contains(with, chat.ALL) {
				messages, _ = db.Messages.GetMessages(bson.M{"to":with})
			}

			result_data["with"] = with
			result_data["messages"] = messages
			result_data["companyId"] = config.CompanyId

			contactsWithOrder := []users.Contact{}
			contactsWithoutOrder := []users.Contact{}
			if contacts, err := GetContacts(db, config.CompanyId); err == nil {
				for _, contact := range contacts {
					if contact.HaveActiveOrder {
						contactsWithOrder = append(contactsWithOrder, contact)
					} else {
						contactsWithoutOrder = append(contactsWithoutOrder, contact)
					}
				}
			}
			result_data["contacts_with_order"] = contactsWithOrder
			result_data["contacts"] = contactsWithoutOrder
			log.Printf("\ncontacts: %+v\ncontacts_with_order: %+v", contactsWithoutOrder, contactsWithOrder)
			r.HTML(200, "chat", result_data, render.HTMLOptions{Layout:"base"})
		})
	return m
}

func GetChatContactsHandler(start_addr string, notifier *ntf.Notifier, db *d.MainDb, config c.ChatConfig) http.Handler {
	m := GetMartini(config.Name, config.CompanyId, start_addr, db)
	m.Post(start_addr, func(render render.Render, req *http.Request) {
		type NewContactsReq struct {
			Exist []string `json:"exist"`
		}
		cr := NewContactsReq{}
		request_body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":"can not read request body"})
			return
		}
		err = json.Unmarshal(request_body, &cr)
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":fmt.Sprintf("can not unmarshal request body %v \n %s", err, request_body)})
			return
		}
		contacts, err := GetContacts(db, config.CompanyId)
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":fmt.Sprintf("db err body %v", err)})
			return
		}
		new_contacts := []users.Contact{}
		old_contacts := []users.Contact{}

		for _, contact := range contacts {
			if utils.InS(contact.ID, cr.Exist) {
				old_contacts = append(old_contacts, contact)
			} else {
				new_contacts = append(new_contacts, contact)
			}
		}
		log.Printf("COFFEE SERV get contacts. \nExist: %+v\nOld: %+v\nNew: %+v", cr.Exist, old_contacts, new_contacts)
		render.JSON(200, map[string]interface{}{
			"ok":true,
			"new":new_contacts,
			"old":old_contacts,
			"next_":time.Now().Unix(),
		})

	})
	return m
}

func GetChatLogoutHandler(start_addr, prefix string, db *d.MainDb, config c.ChatConfig) http.Handler {
	m := GetMartini(config.Name, config.CompanyId, prefix, db)
	m.Get(start_addr,
		web.LoginRequired,
		func(user web.User, db d.DB, ren render.Render, req *http.Request, w http.ResponseWriter) {
			err := db.UsersStorage().LogoutUser(user.UniqueId())
			if err != nil {
				log.Printf("COFFEE error at logout user: %v", err)
			}
			web.StopAuthSession(w)

			ren.Redirect(web.AUTH_URL, 302)
		})
	return m
}
