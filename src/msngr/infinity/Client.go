package infinity

import (
	"encoding/json"
	//"errors"
	//"flag"
	"fmt"
	"io/ioutil"
	"log"
	//"net"
	"net/http"

	//"os"
	"strconv"
	//"sync"
	"time"
)

/*

from: “username”,
	request: {
		(данные запроса)
}
}

●	from(*) | to(*) - имя получателя или отправителя
●	request(*) - тело запроса
○	id(*) - уникальный идентификатор
○	type(*) - тип запроса, может содержать значения:
■	get - запрос на получение каких-либо данных
■	result - результат запроса
■	error - ошибка возникшая в результате запроса
○	query(*) - если поле “type” имеет значение “get” содержит строку с названием команды, если имеет значение “result”, то должно содержать результат в виде списка элементов. Если элементы отсутствуют, список должен быть пустым.


*/
type inPkg struct {
	From    string `json:"from"`
	Message struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Thread  string `json:"thread"`
		Body    string `json:"body"`
		Command struct {
			Title  string `json:"title"`
			Action string `json:"action"`
			Form   struct {
				Title  string  `json:"title"`
				Type   string  `json:"type"`
				Name   string  `json:"name"`
				Fields []Field `json:"fields"`
				Text   string  `json:"text"`
			} `json:"form"`
		} `json:"command"`
	} `json:"message"`
	Request struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Query struct {
			Title  string `json:"title"`
			Action string `json:"action"`
			Form   struct {
				Title  string  `json:"title"`
				Type   string  `json:"type"`
				Name   string  `json:"name"`
				Fields []Field `json:"fields"`
				Text   string  `json:"text"`
			} `json:"form"`
		} `json:"query"`
		//Query *QueryType `json:"query,omitempty"`
	} `json:"request"`
}

type ResultItem struct {
	Title    string `json:"title"`
	Action   string `json:"action"`
	Position int    `json:"position"`
	Form     struct {
		Title  string  `json:"title"`
		Type   string  `json:"type"`
		Name   string  `json:"name"`
		Fields []Field `json:"fields"`
		Text   string  `json:"text"`
	} `json:"form"`
}
type Field struct {
	Name  string `json:"name"`
	Type  string `json:"type,omitempty"`
	Attrs struct {
		Required bool   `json:"required,omitempty"`
		Label    string `json:"label,omitempty"`
		Regexp   string `json:"regexp,omitempty"`
	} `json:"attrs,omitempty"`
	URL   string      `json:"url,omitempty"`
	Value *FieldValue `json:"value,omitempty"`
}

// Omitemty fix for empty structure
type FieldValue struct {
	Value string `json:"value,omitempty"`
	Text  string `json:"text,omitempty"`
}
type FormType struct {
	Title  string  `json:"title,omitempty"`
	Type   string  `json:"type,omitempty"`
	Name   string  `json:"name,omitempty"`
	Label  string  `json:"label,omitempty"`
	URL    string  `json:"url,omitempty"`
	Fields []Field `json:"fields,omitempty"`
}
type QueryType struct {
	Title  string       `json:"title,omitempty"`
	Action string       `json:"action,omitempty"`
	Text   string       `json:"text,omitempty"`
	Form   *FormType    `json:"form,omitempty"`
	Result []ResultItem `json:"result,omitempty"`
}
type RequestType struct {
	ID    string     `json:"id,omitempty"`
	Type  string     `json:"type,omitempty"`
	Query *QueryType `json:"query,omitempty"`
}
type MessageType struct {
	ID     string `json:"id,omitempty"`
	Type   string `json:"type,omitempty"`
	Thread string `json:"thread,omitempty"`
	Body   string `json:"body,omitempty"`
}
type outPkg struct {
	To      string       `json:"to"`
	Message *MessageType `json:"message,omitempty"`
	Request *RequestType `json:"request,omitempty"`
}

func getPrice(inbox *inPkg, p *Infinity) string {
	var order NewOrder_type
	var to Dest
	order.IdService = 5001753333
	for _, v := range inbox.Message.Command.Form.Fields {
		switch v.Name {
		case "street_from":
			var street Address
			err := json.Unmarshal([]byte(v.Value.Value), &street)
			warn(err)
			order.Delivery.IdStreet = street.ID
			order.Delivery.IdRegion = street.IDRegion
			//order.Delivery.IdCity = street.IDCity
		case "entrance":
			order.Delivery.Entrance = v.Value.Value
		case "house_from":
			order.Delivery.House = v.Value.Value
		case "street_to":
			var street Address
			err := json.Unmarshal([]byte(v.Value.Value), &street)
			warn(err)
			to.IdStreet = street.ID
			to.IdRegion = street.IDRegion
			//to.IdCity = street.IDCity
		case "house_to":
			to.House = v.Value.Value
		case "time":
			order.DeliveryTime = v.Value.Value
		}
	}
	if order.DeliveryTime == "" {
		t := time.Now()
		order.DeliveryTime = fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d\n",
			t.Year(), t.Month(), t.Day(),
			t.Hour(), t.Minute(), t.Second())
	}
	order.Destinations = append(order.Destinations, to)
	s, details := p.CalcOrderCost(order)
	log.Println(details)
	cost := strconv.Itoa(s)
	return cost
}
func getCmdList(ServerHost string) []ResultItem {
	//ServerHost := "localhost"
	var results []ResultItem
	// Вызвать такси
	item := new(ResultItem)
	item.Title = "Вызвать такси"
	item.Action = "newOrder"
	item.Position = 0
	item.Form.Title = "Новый заказ"
	item.Form.Type = "form"
	item.Form.Name = "newOrder"
	item.Form.Text = "Откуда: ?(street_from), ?(house_from),?(entrance). Куда: ?(street_to), ?(house_to). Когда: ?(time)."

	// Form description
	/*	field := new(Field)
		field.Name = "from"
		field.Value = "Откуда:"
		field.Type = "static"
		item.Form.Fields = append(item.Form.Fields, *field)
	*/
	field := new(Field)

	field.Name = "street_from"
	field.Attrs.Label = "улица/район"
	field.Type = "dict"
	field.Attrs.Required = true
	field.URL = ServerHost + "/street"
	item.Form.Fields = append(item.Form.Fields, *field)

	field = new(Field)
	field.Name = "house_from"
	field.Attrs.Label = "дом"
	field.Type = "text"
	field.Attrs.Required = true
	item.Form.Fields = append(item.Form.Fields, *field)

	field = new(Field)
	field.Name = "entrance"
	field.Attrs.Label = "подъезд"
	field.Type = "text"
	field.Attrs.Required = true
	item.Form.Fields = append(item.Form.Fields, *field)

	/*	field = new(Field)
		field.Name = "to"
		field.Value = "Куда:"
		field.Type = "static"
		item.Form.Fields = append(item.Form.Fields, *field)
	*/
	field = new(Field)
	field.Name = "street_to"
	field.Attrs.Label = "улица/район"
	field.Type = "text"
	field.Attrs.Required = true
	item.Form.Fields = append(item.Form.Fields, *field)

	field = new(Field)
	field.Name = "house_to"
	field.Attrs.Label = "дом"
	field.Type = "text"
	field.Attrs.Required = true
	item.Form.Fields = append(item.Form.Fields, *field)

	/*field = new(Field)
	field.Name = "where"
	field.Value = "Когда:"
	field.Type = "static"
	item.Form.Fields = append(item.Form.Fields, *field)
	*/
	field = new(Field)
	field.Name = "time"
	field.Attrs.Label = "Время"
	field.Attrs.Required = false
	field.Type = "datetime"
	item.Form.Fields = append(item.Form.Fields, *field)

	results = append(results, *item)

	item = new(ResultItem)
	item.Title = "Написать отзыв"
	item.Action = "comment"
	item.Position = 1
	item.Form.Title = "Отзыв"
	item.Form.Type = "form"
	item.Form.Name = "comment"
	item.Form.Text = "?(comment)"

	field = new(Field)
	field.Name = "comment"
	field.Attrs.Label = "Текст отзыва..."
	field.Attrs.Required = true
	field.Type = "text"
	item.Form.Fields = append(item.Form.Fields, *field)

	results = append(results, *item)

	return results
}

func controlHandler(w http.ResponseWriter, r *http.Request, ServerHost string, p *Infinity) {
	w.WriteHeader(200)
	log.Println("Request...")
	var inbox inPkg
	var outbox outPkg
	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		warn(err)
		log.Println(string(body))
		err = json.Unmarshal(body, &inbox)
		log.Println(inbox)
		log.Println(&inbox.Request.Query)
		warn(err)
		w.Header().Set("Content-type", "application/json")
		// TODO Check general required records in inbox
		outbox.To = inbox.From
		// Request or Message?
		// Simple Message
		/*if inbox.Message.Body != "" {

			outbox.Message.Type = inbox.Message.Type
			outbox.Message.Thread = inbox.Message.Thread
			outbox.Message.ID = fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
			if inbox.Message.Body == "Ping" {
				_, outbox.Message.Body = InfinityAPI.Ping()
			} else {
				outbox.Message.Body = "Unsupported method"
			}
		} else { */
		// Message.Command
		if inbox.Message.Command.Action != "" {
			switch inbox.Message.Command.Action {
			case "information":
				outbox.Message = new(MessageType)
				outbox.Message.Body = "Основная информация по сервису, описание базовых возможностей"
			case "neworder":
				outbox.Message = new(MessageType)
				outbox.Message.Body = "Стоимость вашего заказа будет равна ..." + getPrice(&inbox, p)
				/*
						Здесь нужно дергать InfinityAPI в inbox будет лежать след. структура:
						Клиент
					{
						from: "username",
						message: {
							id: 764578364,
							type: "chat",
							command: {
								action: "neworder",
								form: {
									type: "submit",
									name: "neworder",
									fields: [
										{
											name: "street_from",
											type: "dict",
											value: {
												value: "5009756736"
					}
										},
										{
											name: "house_from",
											type: "text",
											value: {
												text: “8”
					}
										},
										{
											name: "entrance",
											type: "number",
											value: {
												text: “3”
					}
										},
										{
											name: "street_to",
											type: "dict",
											value: {
												text: “5009756248”
					}
										},
										{
											name: "house_to",
											type: "text",
											value: {
												text: “34”
					}
										},
										{
											name: "time",
											type: "datetime",
											value: {
												value: “0”
					}
										}
									]
								}
							}
						}

				*/
			}

		} else {
			// Request?
			if inbox.Request.Query.Action != "" {
				switch inbox.Request.Query.Action {
				case "commands":
					outbox.Request = new(RequestType)
					outbox.Request.Type = "result"
					outbox.Request.ID = fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
					outbox.Request.Query = new(QueryType)
					outbox.Request.Query.Action = "commands"
					outbox.Request.Query.Result = getCmdList(ServerHost)
				}
			} else {
				// Unsupported command
				outbox.Message.Type = inbox.Message.Type
				outbox.Message.Thread = inbox.Message.Thread
				outbox.Message.ID = fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
				outbox.Message.Body = "Unsupported method"
			}
		}
		//}
		t, err := json.Marshal(&outbox)
		warn(err)
		fmt.Fprintf(w, "%s", string(t))
	}
}

// func searchHandler(w http.ResponseWriter, r *http.Request, p Infinity) {
// 	w.WriteHeader(200)
// 	log.Println("Searching address...")
// 	if r.Method == "GET" {
// 		params := url.Values{}
// 		params = r.URL.Query()
// 		query := params.Get("q")
// 		var results []DictItem
// 		if query != "" {
// 			rows := p.AddressesSearch(query).Rows
// 			for _, nitem := range rows {
// 				var item DictItem
// 				var err error
// 				t, err := json.Marshal(nitem)
// 				item.V = string(t)
// 				warn(err)
// 				item.L = nitem.Name
// 				results = append(results, item)
// 			}
// 		}
// 		ans, err := json.Marshal(results)
// 		warn(err)
// 		fmt.Fprintf(w, "%s", string(ans))
// 	}
// }
