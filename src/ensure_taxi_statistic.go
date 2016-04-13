package main

import (
	"log"
	d "msngr/db"
	cfg "msngr/configuration"
	t "msngr/taxi"

	"github.com/tealeg/xlsx"
	"gopkg.in/mgo.v2/bson"
)

func EnsureStatistic() {
	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell
	var err error

	config := cfg.ReadConfig()
	db := d.NewMainDb(config.Main.Database.ConnString, config.Main.Database.Name)

	file = xlsx.NewFile()
	type Source struct {
		Name string `bson:"_id"`
	}
	sources := []Source{}

	db.Orders.Collection.Pipe(
		[]bson.M{bson.M{"$group":bson.M{"_id":"$source"}}},
	).All(&sources)

	for _, source := range sources {
		sheet, err = file.AddSheet(source.Name)
		if err != nil {
			log.Printf("Error at edding sheet to file")
			return
		}
		row = sheet.AddRow()
		for _, h_cell := range []string{"Телефон", "Статус", "Дата", "Стоимость", "Адрес подачи", "Адрес назначения", "Позывной автомобиля", "ФИО водителя"} {
			cell = row.AddCell()
			cell.Value = h_cell
		}

		orders, err := db.Orders.GetBy(bson.M{"is_active":false, "source":source.Name})
		if err != nil {
			log.Printf("Error at getting orders from %+v is: %v", config.Main.Database, err)
			return
		}

		for _, order := range orders {
			log.Printf("adding row for order %+v", order)
			user, u_err := db.Users.GetUserById(order.Whom)
			if u_err != nil || user == nil {
				log.Printf("No user found at id: %v", order.Whom)
				continue
			}
			row = sheet.AddRow()

			ph_c := row.AddCell()
			ph_c.SetString(user.Phone)

			stat_c := row.AddCell()
			if state, ok := t.InfinityStatusesName[order.OrderState]; ok {
				stat_c.SetString(state)
			}else {
				stat_c.SetString("Не определен")
			}

			time_c := row.AddCell()
			time_c.SetDateTime(order.When)

			if len(order.OrderData.Content) > 0 {
				log.Printf("we have additional data of this order...")
				cost_c := row.AddCell()
				cost_c.SetInt(order.OrderData.Get("Cost").(int))

				deliv_c := row.AddCell()
				deliv_c.SetString(order.OrderData.Get("DeliveryStr").(string))

				dest_c := row.AddCell()
				dest_c.SetString(order.OrderData.Get("DestinationsStr").(string))

				car_c := row.AddCell()
				car_c.SetString(order.OrderData.Get("Car").(string))

				driver_c := row.AddCell()
				driver_c.SetString(order.OrderData.Get("Drivers").(string))
			}
			log.Printf("Added row: %+v", row)
		}
	}
	file.Save("taxi_statistic.xlsx")
}
func main() {
	EnsureStatistic()
}