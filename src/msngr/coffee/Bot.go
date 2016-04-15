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

const (
	DRINKS = "drinks"
	VOLUMES = "volumes"
	ADDITIVES = "additives"
	BAKES = "bakes"
)

func FormItems(strings []string) {
	result := []s.FieldItem{}
	for _, el := range strings {
		result = append(result, s.FieldItem{
			Value:el,
			Content:s.FieldItemContent{
				Title:el,
			},
		})
	}
}

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
				Text:  "?(drink) ?(volume), ?(additive), ?(count), ?(to_time)",
				Fields: []s.OutField{
					s.OutField{
						Name: "drink",
						Type: "list-single",
						Attributes: s.FieldAttribute{
							Label:    "напиток",
							Required: true,
						},
						Items:coffeeHouseConfig.ToFieldItems(DRINKS),
					},
					s.OutField{
						Name: "volume",
						Type: "list-single",
						Attributes: s.FieldAttribute{
							Label:"объем",
							Required: true,
						},
						Items:coffeeHouseConfig.ToFieldItems(VOLUMES),
					},
					s.OutField{
						Name: "additive",
						Type: "list-single",
						Attributes: s.FieldAttribute{
							Label:    "добавка",
							Required: false,
						},
						Items:coffeeHouseConfig.ToFieldItems(ADDITIVES),
					},
					s.OutField{
						Name:"count",
						Type:"number",
						Attributes:s.FieldAttribute{
							Label:"количество",
							Required:false,
							EmptyText:"1",
						},
					},
					s.OutField{
						Name:"to_time",
						Type:"text",
						Attributes:s.FieldAttribute{
							Label:"когда",
							Required:false,
							EmptyText:"сейчас",
						},
						Items:FormItems([]string{"сейчас", "через 10 минут", "через 20 минут", "через 20 минут",, "через час"}),
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
				Text:  "Ваш заказ: ?(bake), ?(count) ?(to_time)",
				Fields: []s.OutField{
					s.OutField{
						Name: "bake",
						Type: "list-single",
						Attributes: s.FieldAttribute{
							Label:    "выпечка",
							Required: true,
						},
						Items:coffeeHouseConfig.ToFieldItems(BAKES),
					},
					s.OutField{
						Name:"count",
						Type:"number",
						Attributes:s.FieldAttribute{
							Label:"количество",
							Required:false,
							EmptyText:"1",
						},
					},
					s.OutField{
						Name:"to_time",
						Type:"text",
						Attributes:s.FieldAttribute{
							Label:"когда",
							Required:false,
							EmptyText:"сейчас",
						},
						Items:FormItems([]string{"сейчас", "через 10 минут", "через 20 минут", "через 20 минут",, "через час"}),
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
			log.Printf("COFFEE BOT error getting last order for %v is: %v", in.From, err)
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
		"information":m.NewInformationProcessor(config.Information, commandsGenerator),
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
				order, err := NewCoffeeOrderFromForm(command.Form)
				if err != nil {
					log.Printf("COFFEE BOT error at forming order from form: %v", err)
					return m.MESSAGE_DATA_ERROR_RESULT
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
				cmds, err := odp.CommandsFunc(in)
				if err != nil {
					return s.ErrorMessageResult(err, cmds)
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
				order, err := NewCoffeeOrderFromForm(command.Form)
				if err != nil {
					log.Printf("COFFEE BOT error at forming order from form: %v", err)
					return m.MESSAGE_DATA_ERROR_RESULT
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
				cmds, err := odp.CommandsFunc(in)
				if err != nil {
					return s.ErrorMessageResult(err, cmds)
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
		err := cop.Storage.Orders.SetActive(lastOrder.OrderId, lastOrder.Source, false)
		if err != nil {
			log.Printf("CB Error at setting active is false for order %v", err)
			return m.DB_ERROR_RESULT
		}
		cmds, err := cop.CommandsFunc(in)
		if err != nil {
			log.Printf("CB Error at forming commands %v", err)
		}
		return &s.MessageResult{Body:"Ваш заказ отменен!", Commands:cmds}
	}
	return &s.MessageResult{Body:"У вас нечего отменять."}
}

type RepeatOrderProcessor struct {
	CompanyName  string
	CommandsFunc m.CommandsGenerator
	Storage      *db.MainDb
}

func (rop *RepeatOrderProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	lastOrder, err := rop.Storage.Orders.GetByOwnerLast(in.From, rop.CompanyName)
	if err != nil {
		return m.DB_ERROR_RESULT
	}
	if lastOrder != nil {
		order, err := NewCoffeeOrderFromMap(lastOrder.OrderData.Content)
		if err != nil {
			log.Printf("CB Error at forming new coffee order %v", err)
			return m.MESSAGE_DATA_ERROR_RESULT
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
		cmds, err := rop.CommandsFunc(in)
		if err != nil {
			log.Printf("CB Error at forming commands %v", err)
		}
		return &s.MessageResult{Body:"Ваш заказ повторен!", Commands:cmds}
	}
	return &s.MessageResult{Body:"Нечего повторять."}
}