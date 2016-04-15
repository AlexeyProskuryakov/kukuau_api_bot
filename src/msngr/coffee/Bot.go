package coffee

import (
	s "msngr/structs"
	m "msngr"
	u "msngr/utils"
	c "msngr/configuration"
	"msngr/db"
	"time"
	"log"
)

func getCommands(coffeeHouseConfig *CoffeeHouseConfiguration, isFirst bool, isActive bool) *[]s.OutCommand {
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
				Text:  "Ваш заказ: ?(drink) ?(volume) ?(additive), ?(count) ?(time)",
				Fields: []s.OutField{
					s.OutField{
						Name: "drink",
						Type: "list-single",
						Attributes: s.FieldAttribute{
							Label:    "напиток",
							Required: true,
						},
						Items:coffeeHouseConfig.ToFieldItems("drinks"),
					},
					s.OutField{
						Name: "volume",
						Type: "list-single",
						Attributes: s.FieldAttribute{
							Label:"объем",
							Required: false,
						},
						Items:coffeeHouseConfig.ToFieldItems("volume"),
					},
					s.OutField{
						Name: "additive",
						Type: "list-single",
						Attributes: s.FieldAttribute{
							Label:    "добавка",
							Required: false,
						},
						Items:coffeeHouseConfig.ToFieldItems("addititves"),
					},
					s.OutField{
						Name:"count",
						Type:"text",
						Attributes:s.FieldAttribute{
							Label:"сколько штук",
							Required:false,
						},
					},
					s.OutField{
						Name:"time",
						Type:"text",
						Attributes:s.FieldAttribute{
							Label:"когда",
							Required:false,
						},
					},
				},
			},
		},
		s.OutCommand{
			Title:"Выпечка",
			Action:"order_bake",
			Position:1,
			Form: &s.OutForm{
				Title: "Заказ выпечки",
				Type:  "form",
				Name:  "order_bake_form",
				Text:  "Ваш заказ: ?(bake)",
				Fields: []s.OutField{
					s.OutField{
						Name: "bake",
						Type: "list-single",
						Attributes: s.FieldAttribute{
							Label:    "выпечка",
							Required: true,
						},
						Items:coffeeHouseConfig.ToFieldItems("bakes"),
					},
				},
			},
		},
	}
	position := 1
	if !isFirst {
		position += 1
		commands = append(commands,
			s.OutCommand{
				Title:"Повторить предыдущий заказ",
				Action:"repeat",
				Position:position,
			},
		)
	}
	if isActive {
		position += 1
		commands = append(commands,
			s.OutCommand{
				Title:"Отменить текущий заказ",
				Action:"cancel",
				Position:position,
			},
		)
	}
	return &commands
}

func getAdditionalFuncs(orderId int64, companyName, userName string) []db.AdditionalFuncElement {
	context := map[string]interface{}{
		"order_id":orderId,
		"company_name":companyName,
		"user_name":userName,
	}
	result := []db.AdditionalFuncElement{
		db.AdditionalFuncElement{
			Name:"Отменить",
			Action:"cancel",
			Context:context,
		},
		db.AdditionalFuncElement{
			Name:"Начать",
			Action:"start",
			Context:context,
		},
		db.AdditionalFuncElement{
			Name:"Закончить",
			Action:"end",
			Context:context,
		},
	}
	return result
}

func FormBotCoffeeContext(config c.CoffeeConfig, store *db.MainDb, coffeeHouseConfiguration *CoffeeHouseConfiguration) *m.BotContext {

	commandsGenerator := func(in *s.InPkg) (*[]s.OutCommand, error) {
		lastOrder, err := store.Orders.GetByOwnerLast(in.From, config.Name)
		if err != nil {
			log.Printf("COFFEE BOT error getting lat order for %v is: %v", in.From, err)
			return nil, err
		}
		var isFirst, isActive bool
		if lastOrder != nil {
			isFirst = false
			isActive = lastOrder.Active
		}
		commands := getCommands(coffeeHouseConfiguration, isFirst, isActive)
		return commands, nil
	}

	result := m.BotContext{}
	result.RequestProcessors = map[string]s.RequestCommandProcessor{
		"commands":&RequestCommandsProcessor{CommandsFunc:commandsGenerator},
	}
	result.MessageProcessors = map[string]s.MessageCommandProcessor{
		"":m.NewFuncTextBodyProcessor(store, commandsGenerator, config.Name, nil),
		"information":m.NewInformationProcessor(config.Information),
		"order_bake":&OrderBakeProcessor{Storage:store, CompanyName:config.Name, CommandsFunc:commandsGenerator},
		"order_drink":&OrderDrinkProcessor{Storage:store, CompanyName:config.Name, CommandsFunc:commandsGenerator},
		"cancel":&CancelOrderProcessor{Storage:store, CompanyName:config.Name, CommandsFunc:commandsGenerator},
		"repeat":&RepeatOrderProcessor{Storage:store, CompanyName:config.Name, CommandsFunc:commandsGenerator},
	}
	return &result
}

type RequestCommandsProcessor struct {
	CommandsFunc m.CommandsGenerator
}

func (rcp *RequestCommandsProcessor)ProcessRequest(in *s.InPkg) *s.RequestResult {
	cmds, err := rcp.CommandsFunc(in)
	if err != nil {
		return &s.RequestResult{Error:err}
	}
	return &s.RequestResult{Commands:cmds}
}

type OrderDrinkProcessor struct {
	Storage      *db.MainDb
	CompanyName  string
	CommandsFunc m.CommandsGenerator
}

func (odp *OrderDrinkProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	if in.UserData != nil && in.Message != nil && in.Message.Commands != nil {
		err := odp.Storage.Users.StoreUserData(in.From, in.UserData)
		if err != nil {
			return m.DB_ERROR_RESULT
		}
		commands := *(in.Message.Commands)
		for _, command := range commands {
			if command.Action == "order_drink" && command.Form.Name == "order_drink_form" {
				cmds, err := odp.CommandsFunc(in)
				if err != nil {
					return s.ErrorMessageResult(err, cmds)
				}
				order, err := NewCoffeeOrderFromForm(command.Form)
				if err != nil {
					log.Printf("COFFEE BOT error at forming order from form: %v", err)
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
				if err != nil {
					log.Printf("CB Error at storing drink order %v", err)
					return m.DB_ERROR_RESULT
				}
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
					AdditionalFuncs:getAdditionalFuncs(id, odp.CompanyName, in.From),
					RelatedOrderState:"Отправленно в кофейню",
				})
				if err != nil {
					log.Printf("CB Error at storing drink message %v", err)
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
	Storage      *db.MainDb
	CompanyName  string
	CommandsFunc m.CommandsGenerator
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
				cmds, err := odp.CommandsFunc(in)
				if err != nil {
					return s.ErrorMessageResult(err, cmds)
				}
				order, err := NewCoffeeOrderFromForm(command.Form)
				if err != nil {
					log.Printf("COFFEE BOT error at forming order from form: %v", err)
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
				if err != nil {
					log.Printf("CB Error at storing bake order %v", err)
					return m.DB_ERROR_RESULT
				}
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
					AdditionalFuncs:getAdditionalFuncs(id, odp.CompanyName, in.From),
					RelatedOrderState:"Отправленно в кофейню",
				})
				if err != nil {
					log.Printf("CB Error at storing bake message %v", err)
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

type CancelOrderProcessor struct {
	CompanyName  string
	CommandsFunc m.CommandsGenerator
	Storage      *db.MainDb
}

func (cop *CancelOrderProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	cmds, err := cop.CommandsFunc(in)
	if err != nil {
		log.Printf("CB Error at forming commands %v", err)
	}
	lastOrder, err := cop.Storage.Orders.GetByOwnerLast(in.From, cop.CompanyName)
	if err != nil {
		return m.DB_ERROR_RESULT
	}
	if lastOrder != nil {
		err = cop.Storage.Messages.StoreMessageObject(db.MessageWrapper{
			MessageID:in.Message.ID,
			From:in.From,
			To:cop.CompanyName,
			Body:"Отменяю заказ. Простите пожалуйста. :(",
			Unread:1,
			NotAnswered:1,
			Time:time.Now(),
			TimeStamp:time.Now().Unix(),
			TimeFormatted: time.Now().Format(time.Stamp),
			Attributes:[]string{"coffee"},
		})
		if err != nil {
			log.Printf("CB Error at storing message for cancel %v", err)
			return m.DB_ERROR_RESULT
		}
		return &s.MessageResult{Body:"Ваш заказ отменен!", Commands:cmds}
	}
	return &s.MessageResult{Body:"У вас нечего отменять.", Commands:cmds}
}

type RepeatOrderProcessor struct {
	CompanyName  string
	CommandsFunc m.CommandsGenerator
	Storage      *db.MainDb
}

func (rop *RepeatOrderProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	cmds, err := rop.CommandsFunc(in)
	if err != nil {
		log.Printf("CB Error at forming commands %v", err)
	}
	lastOrder, err := rop.Storage.Orders.GetByOwnerLast(in.From, rop.CompanyName)
	if err != nil {
		return m.DB_ERROR_RESULT
	}
	if lastOrder != nil {
		order, err := NewCoffeeOrderFromMap(lastOrder.OrderData.Content)
		if err != nil {
			log.Printf("CB Error at forming new coffee order %v", err)
			return s.ErrorMessageResult(err, cmds)
		}
		id := u.GenIntId()
		err = rop.Storage.Orders.AddOrderObject(db.OrderWrapper{
			OrderId: id,
			When:time.Now(),
			Whom:in.From,
			Source:rop.CompanyName,
			Active:true,
			OrderData:order.ToOrderData(),
		})
		if err != nil {
			log.Printf("CB Error at storing repeated order %v", err)
			return m.DB_ERROR_RESULT
		}
		err = rop.Storage.Messages.StoreMessageObject(db.MessageWrapper{
			MessageID:in.Message.ID,
			From:in.From,
			To:rop.CompanyName,
			Body:"Мне бы повторить...",
			Unread:1,
			NotAnswered:1,
			Time:time.Now(),
			TimeStamp:time.Now().Unix(),
			TimeFormatted: time.Now().Format(time.Stamp),
			Attributes:[]string{"coffee"},
			AdditionalData:order.ToAdditionalMessageData(),
			AdditionalFuncs:getAdditionalFuncs(id, rop.CompanyName, in.From),
			RelatedOrderState:"Отправленно в кофейню",
		})
		if err != nil {
			log.Printf("CB Error at storing message for order %v", err)
			return m.DB_ERROR_RESULT
		}
		return &s.MessageResult{Body:"Ваш заказ повторен!", Commands:cmds}
	}
	return &s.MessageResult{Body:"Нечего повторять.", Commands:cmds}
}