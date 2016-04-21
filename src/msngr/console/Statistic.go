package console

import (
	"log"
	d "msngr/db"
	cfg "msngr/configuration"
	t "msngr/taxi"

	"github.com/tealeg/xlsx"
	"gopkg.in/mgo.v2/bson"
	"msngr/utils"
	"fmt"
)

var config = cfg.ReadConfig()

func getDbsWithOrders() []string {
	dbHelper := d.NewDbHelper(config.Main.Database.ConnString, "local")
	result := []string{}
	if names, err := dbHelper.Session.DatabaseNames(); err == nil {
		for _, dbName := range names {
			collections, err := dbHelper.Session.DB(dbName).CollectionNames()
			if err != nil {
				continue
			}
			if utils.InS("orders", collections) {
				result = append(result, dbName)
			}
		}
	}
	return result
}

func EnsureStatistic(toPath string) error{
	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell
	var err error

	file = xlsx.NewFile()
	type Source struct {
		Name string `bson:"_id"`
	}
	sources := []Source{}
	dbsWithOrders := getDbsWithOrders()
	log.Printf("Databases with orders is: %+v", dbsWithOrders)

	for _, dbName := range dbsWithOrders {
		db := d.NewMainDb(config.Main.Database.ConnString, dbName)
		sheet, err = file.AddSheet(dbName)
		if err != nil {
			log.Printf("Error at edding sheet to file")
			continue
		}
		db.Orders.Collection.Pipe(
			[]bson.M{bson.M{"$group":bson.M{"_id":"$source"}}},
		).All(&sources)

		for _, source := range sources {
			row = sheet.AddRow()
			cell = row.AddCell()
			cell.Value = source.Name

			row = sheet.AddRow()
			for _, h_cell := range []string{"Телефон", "Статус", "Дата", "Стоимость", "Адрес подачи", "Адрес назначения", "Позывной автомобиля", "ФИО водителя"} {
				cell = row.AddCell()
				cell.Value = h_cell
			}

			orders, err := db.Orders.GetBy(bson.M{"source":source.Name})
			if err != nil {
				log.Printf("Error at getting orders from %+v is: %v", config.Main.Database, err)
				return err
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
					log.Printf("we have additional data of order %v", order.OrderId)
					cost := order.OrderData.Get("Cost")
					if cost != nil{
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
				}
				log.Printf("Added row: %+v", row)
			}
		}
	}
	fName := fmt.Sprintf("%v/statistic.xlsx", toPath)
	file.Save(fName)
	return nil
}