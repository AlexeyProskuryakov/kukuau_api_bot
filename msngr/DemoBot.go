package msngr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	// "strconv"
	"errors"
	"time"
)

type requestCommandProcessor interface {
	ProcessRequest(in inPkg) ([]Command, error)
}

type messageCommandProcessor interface {
	ProcessMessage(in inPkg) (string, error)
}

var requestCommands = map[string]requestCommandProcessor{
	"commands": CommandsHandler{},
}

var messageCommands = map[string]messageCommandProcessor{
	"information":     InformationHandler{},
	"new_order":       NewOrderHandler{},
	"cancel_order":    CancelOrderHandler{},
	"calculate_price": CalculatePriceHandler{},
}

type CommandsHandler struct{}

func (s CommandsHandler) ProcessRequest(in inPkg) ([]Command, error) {
	uh := GetUserHandler()
	state := uh.GetUserState(in.From)
	if state == ORDER_CREATE {
		return []Command{
			Command{
				Title:    "Отменить заказ",
				Action:   "cancel_order",
				Position: 0,
				Form:     Form{},
			},
		}, nil

	} else {
		taxi_call_form := Form{
			Title: "Форма вызова такси",
			Type:  "form",
			Name:  "call_taxi",
			Text:  "Откуда: ?(street_from), ?(house_from), ?(entrance). Куда: ?(street_to), ?(house_to). Когда: ?(time)",
			Fields: []Field{
				Field{
					Name:     "street_from",
					Required: true,
					Type:     "dict",
					Label:    "FromLabel",
					Value:    "FromValue",
					Attributes: FieldAttribute{
						Label:    "улица/район",
						Required: true,
						URL:      "http://foo.bar",
					},
				},
				Field{
					Name:     "house_from",
					Required: true,
					Type:     "text",
					Label:    "house_from",
					Value:    "house_from",
					Attributes: FieldAttribute{
						Label:    "дом",
						Required: true,
					},
				},
				Field{
					Name:     "entrance",
					Required: false,
					Type:     "number",
					Label:    "entrance",
					Value:    "entrance",
					Attributes: FieldAttribute{
						Label:    "подъезд",
						Required: false,
					},
				},
				Field{
					Name:     "street_to",
					Required: true,
					Type:     "text",
					Label:    "time_label",
					Value:    "time_value",
					Attributes: FieldAttribute{
						Label:    "улица/район",
						Required: true,
						URL:      "http://foo.bar",
					},
				},
				Field{
					Name:     "house_to",
					Required: true,
					Type:     "text",
					Label:    "house_to",
					Value:    "house_to",
					Attributes: FieldAttribute{
						Label:    "дом",
						Required: true,
					},
				},
				Field{
					Name:     "time",
					Required: false,
					Type:     "text",
					Label:    "time",
					Value:    "time",
					Attributes: FieldAttribute{
						Label:    "время",
						Required: false,
					},
				},
			},
		}
		commands := []Command{
			Command{
				Title:    "Вызвать такси",
				Action:   "new_order",
				Position: 0,
				Form:     taxi_call_form,
			},

			Command{
				Title:    "Рассчитать цену",
				Action:   "calculate_price",
				Position: 1,
				Form:     taxi_call_form,
			},
		}
		return commands, nil
	}
}

type InformationHandler struct{}

func (ih InformationHandler) ProcessMessage(in inPkg) (string, error) {
	return "!!! This is TAXI !!! ", nil
}

type NewOrderHandler struct{}

func (noh NewOrderHandler) ProcessMessage(in inPkg) (string, error) {
	uh := GetUserHandler()
	state := uh.GetUserState(in.From)
	if state != ORDER_CREATE {
		var from, to, hf, ht, t string
		for _, field := range in.Message.Command.Form.Fields {
			switch fn := field.Name; fn {
			case "street_from":
				from = field.Value.Text
			case "street_to":
				to = field.Value.Text
			case "house_to":
				ht = field.Value.Text
			case "house_from":
				hf = field.Value.Text
			case "time":
				fv := field.Value.Value
				if fv == "0" {
					t = "сейчас"
				} else {
					t = fmt.Sprintf("через %v минут", rand.Int31n(10)+10)
				}
			}

		}

		uh.SetUserState(in.From, ORDER_CREATE)
		result := fmt.Sprintf("Ваш заказ создан! Поедем из ул %v дом %v, на %v к дому %v. Cтоймость %v, машина прибудет %v", from, hf, to, ht, rand.Int31n(500)+50, t)
		return result, nil
	} else {
		return "Заказ уже создан!", errors.New("Заказ уже создан!")
	}

}

type CancelOrderHandler struct{}

func (coh CancelOrderHandler) ProcessMessage(in inPkg) (string, error) {
	uh := GetUserHandler()
	uh.SetUserState(in.From, ORDER_CANCELED)
	return "Ваш заказ отменен", nil
}

type CalculatePriceHandler struct {
}

func (cph CalculatePriceHandler) ProcessMessage(in inPkg) (string, error) {
	return fmt.Sprintf("Стоймость будет всего лишь %v рублей!", rand.Int31n(500)+50), nil
}

func genId() string {
	//не привязывайся ко времени, может бть в 1 микросекуну много сообщений и ид долэны ыть разными
	return fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
}

func getInPackage(r *http.Request) (inPkg, error) {
	log.Println("getting in! will retrieve body from request")
	var in inPkg

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error at reading: %q \n", err)
	}

	err = json.Unmarshal(body, &in)
	if err != nil {
		log.Printf("error at unmarshal: %q \n", err)
	}

	log.Printf("request data is:\n%+v\n", in)
	return in, err
}

func setOutPackage(w http.ResponseWriter, out outPkg) {
	log.Println("forming out! will marshaing out response")

	jsoned_out, err := json.Marshal(&out)
	if err != nil {
		log.Println(jsoned_out, err)
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "%s", string(jsoned_out))
}

func BotControlHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("processing request...")
	if r.Method != "POST" {
		http.Error(w, "I can not work with non POST methods", 405)
		return
	}

	in, err := getInPackage(r)
	out := new(outPkg)

	log.Println("forming response...")

	out.To = in.From

	if in.Request != nil {
		log.Println("processing request")
		action := in.Request.Query.Action
		out.Request = &OutRequest{ID: genId(), Type: "result"}
		out.Request.Query.Action = action
		if commandProcessor, ok := requestCommands[action]; ok {
			out.Request.Query.Result, err = commandProcessor.ProcessRequest(in)
		}

	} else if in.Message != nil {
		log.Println("processing message")
		out.Message = &OutMessage{Type: in.Message.Type, Thread: in.Message.Thread, ID: genId()}
		action := in.Message.Command.Action
		if commandProcessor, ok := messageCommands[action]; ok {
			out.Message.Body, err = commandProcessor.ProcessMessage(in)
		}

	}
	log.Printf("%+v\n", out)

	if err != nil {
		out.Message = &OutMessage{Type: "error", Thread: "0", ID: genId(), Body: fmt.Sprintf("%+v", err)}
	}
	setOutPackage(w, *out)
}
