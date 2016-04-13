package coffee

import (
	"fmt"
	s "msngr/structs"
	m "msngr"
	u "msngr/utils"
	c "msngr/configuration"
	"msngr/db"
	"time"
	"log"
)

func getCommands(dictUrlPrefix string) *[]s.OutCommand {
	drinkSearchUrl := fmt.Sprintf("%v/drink", dictUrlPrefix)
	additiveSearchUrl := fmt.Sprintf("%v/additive", dictUrlPrefix)
	volumeSearchUrl := fmt.Sprintf("%v/volume", dictUrlPrefix)
	bakeSearchUrl := fmt.Sprintf("%v/bake", dictUrlPrefix)

	commands := []s.OutCommand{
		s.OutCommand{
			Title: "Напитки",
			Action: "order_drink",
			Position:0,
			Repeated:true,
			Form: &s.OutForm{
				Title: "Заказ напитка",
				Type:  "form",
				Name:  "order_drink_form",
				Text:  "Какой ?(drink) его ?(volume) и ?(additive).",
				Fields: []s.OutField{
					s.OutField{
						Name: "drink",
						Type: "dict",
						Attributes: s.FieldAttribute{
							Label:    "напиток",
							Required: true,
							URL:      &drinkSearchUrl,
						},
					},
					s.OutField{
						Name: "volume",
						Type: "dict",
						Attributes: s.FieldAttribute{
							Label:"объем",
							Required: false,
							URL:      &volumeSearchUrl,
						},
					},
					s.OutField{
						Name: "additive",
						Type: "dict",
						Attributes: s.FieldAttribute{
							Label:    "добавка",
							Required: false,
							URL:      &additiveSearchUrl,
						},
					},
				},
			},
		},
		s.OutCommand{
			Title:"Выпечка",
			Action:"order_bake",
			Position:1,
			Repeated:true,
			Form: &s.OutForm{
				Title: "Заказ выпечки",
				Type:  "form",
				Name:  "order_bake_form",
				Text:  "Какая ?(bake) ?",
				Fields: []s.OutField{
					s.OutField{
						Name: "bake",
						Type: "dict",
						Attributes: s.FieldAttribute{
							Label:    "выпечка",
							Required: true,
							URL:      &bakeSearchUrl,
						},
					},
				},
			},
		},
	}

	return &commands
}

func FormBotCoffeeContext(config c.CoffeeConfig, store *db.MainDb) *m.BotContext {
	result := m.BotContext{}
	cmds := getCommands(config.DictUrl)

	result.RequestProcessors = map[string]s.RequestCommandProcessor{
		"commands":&RequestCommandsProcessor{DictUrlPrefix:config.DictUrl},
	}
	result.MessageProcessors = map[string]s.MessageCommandProcessor{
		"":m.NewSimpleTextBodyProcessor(store, cmds, config.Name, nil),
		"information":m.NewInformationProcessor(config.Information),
		"order_bake":&OrderBakeProcessor{Storage:store, CompanyName:config.Name, DictUrl:config.DictUrl},
		"order_drink":&OrderDrinkProcessor{Storage:store, CompanyName:config.Name, DictUrl:config.DictUrl},
	}
	return &result
}

type RequestCommandsProcessor struct {
	DictUrlPrefix string
}

func (rcp *RequestCommandsProcessor)ProcessRequest(in *s.InPkg) *s.RequestResult {
	result := s.RequestResult{Commands:getCommands(rcp.DictUrlPrefix)}
	return &result
}

type OrderDrinkProcessor struct {
	Storage     *db.MainDb
	CompanyName string
	DictUrl     string
}

func (odp *OrderDrinkProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	log.Printf("CB : in: %+v, %+v, %+v", in.UserData, in.Message, in.Message.Commands)
	if in.UserData != nil && in.Message != nil && in.Message.Commands != nil {
		err := odp.Storage.Users.StoreUserData(in.From, in.UserData)
		if err != nil {
			return m.DB_ERROR_RESULT
		}
		commands := *(in.Message.Commands)
		for _, command := range commands {
			if command.Action == "order_drink" && command.Form.Name == "order_drink_form" {
				cmds := getCommands(odp.DictUrl)
				order, err := NewCoffeeOrderFromForm(command.Form)
				if err != nil {
					return s.ErrorMessageResult(err, cmds)
				}
				id := u.GenIntId()
				err = odp.Storage.Orders.AddOrderObject(db.OrderWrapper{
					OrderId:id,
					When:time.Now(),
					Whom:in.From,
					Source:odp.CompanyName,
					Active:true,
					OrderData:order.ToOrderData(),
				})

				err = odp.Storage.Messages.StoreMessageObject(db.MessageWrapper{
					MessageID:in.Message.ID,
					From:in.From,
					To:odp.CompanyName,
					Body:"Заказ напитка!",
					Unread:1,
					NotAnswered:1,
					Time:time.Now(),
					TimeStamp:time.Now().Unix(),
					TimeFormatted: time.Now().Format(time.Stamp),
					Attributes:[]string{"coffee"},
					AdditionalData:order.ToAdditionalMessageData(),
				})

				if err != nil {
					return m.DB_ERROR_RESULT
				}

				return &s.MessageResult{
					Commands:cmds,
					Type:"chat",
					Body:"Ваш заказ создан!",
				}

			}

		}
	}
	return m.MESSAGE_DATA_ERROR_RESULT
}

type OrderBakeProcessor struct {
	Storage     *db.MainDb
	CompanyName string
	DictUrl     string
}

func (odp *OrderBakeProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	if in.UserData != nil && in.Message != nil && in.Message.Commands != nil {
		err := odp.Storage.Users.StoreUserData(in.From, in.UserData)
		if err != nil {
			return m.DB_ERROR_RESULT
		}
		commands := *(in.Message.Commands)
		for _, command := range commands {
			if command.Action == "order_bake" && command.Form.Name == "order_bake_form" {
				cmds := getCommands(odp.DictUrl)
				order, err := NewCoffeeOrderFromForm(command.Form)
				if err != nil {
					return s.ErrorMessageResult(err, cmds)
				}
				id := u.GenIntId()
				odp.Storage.Orders.AddOrderObject(db.OrderWrapper{
					OrderId:id,
					When:time.Now(),
					Whom:in.From,
					Source:odp.CompanyName,
					Active:true,
					OrderData:order.ToOrderData(),
				})
				err = odp.Storage.Messages.StoreMessageObject(db.MessageWrapper{
					MessageID:in.Message.ID,
					From:in.From,
					To:odp.CompanyName,
					Body:"Заказ выпечки!",
					Unread:1,
					NotAnswered:1,
					Time:time.Now(),
					TimeStamp:time.Now().Unix(),
					TimeFormatted: time.Now().Format(time.Stamp),
					Attributes:[]string{"coffee"},
					AdditionalData:order.ToAdditionalMessageData(),
				})
				if err != nil {
					return m.DB_ERROR_RESULT
				}

				return &s.MessageResult{
					Commands:cmds,
					Type:"chat",
					Body:"Ваш заказ создан!",
				}
			}

		}
	}
	return m.MESSAGE_DATA_ERROR_RESULT
}
