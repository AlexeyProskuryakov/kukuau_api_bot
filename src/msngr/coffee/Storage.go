package coffee

import (
	d "msngr/db"
	m "msngr"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"sort"
	"msngr/configuration"
	s "msngr/structs"
	"encoding/json"
	"log"
)

type CoffeeHouseConfiguration struct {
	Name      string `bson:"name"`
	Bakes     map[string]string `bson:"bakes"`
	Drinks    map[string]string `bson:"drinks"`
	Additives map[string]string `bson:"additives"`
	Syrups    map[string]string `bson:"syrups"`
	Volumes   []string `bson:"volumes"`
}

func (cc CoffeeHouseConfiguration) getFieldContent(fieldBsonName string) []string {
	v := reflect.ValueOf(cc)
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fi := typ.Field(i)
		if tagv := fi.Tag.Get("bson"); tagv == fieldBsonName {
			return v.Field(i).Interface().([]string)
		}
	}
	return []string{}
}

func (chc *CoffeeHouseConfiguration) ToFieldItems(fName string) []s.FieldItem {
	content := chc.getFieldContent(fName)
	result := []s.FieldItem{}
	for _, el := range content {
		result = append(result, s.FieldItem{
			Value:el,
			Content:s.FieldItemContent{
				Title:el,
			},
		})
	}
	return result
}

func (cc CoffeeHouseConfiguration) Autocomplete(q, fieldBsonName string) []string {
	content := cc.getFieldContent(fieldBsonName)
	by := m.ByFuzzyEquals{Data:content, Center:q}
	sort.Sort(by)
	return by.Data
}

type CoffeeOrder struct {
	Type     string `json:"type"`
	Drink    string `json:"drink"`
	Bake     string `json:"bake"`
	Additive string `json:"additive"`
	Syrup    string `json:"syrup"`
	Sugar    string `json:"sugar"`
	Count    string `json:"count"`
	ToTime   string `json:"to_time"`
}

func NewCoffeeOrderFromForm(form s.InForm) (*CoffeeOrder, error) {
	result := CoffeeOrder{}
	if form.Name == "order_drink_form" {
		result.Type = "drink"
		drink, _ := form.GetAny("drink")
		result.Drink = drink
		additive, _ := form.GetAny("additive")
		result.Additive = additive
		syrup, _ := form.GetAny("syrup")
		result.Syrup = syrup
		sugar, _ := form.GetAny("sugar")
		result.Sugar = sugar

	} else if form.Name == "order_bake_form" {
		result.Type = "bake"
		bake, _ := form.GetAny("bake")
		result.Bake = bake
	}
	count, _ := form.GetAny("count")
	result.Count = count
	time, _ := form.GetAny("to_time")
	result.ToTime = time
	return &result, nil
}

func (co CoffeeOrder) ToOrderData() d.OrderData {
	rawData, err := json.Marshal(co)
	if err != nil {
		log.Printf("CoffeeSt tod ERROR at marshall %v", err)
	}
	data := map[string]interface{}{}
	if err = json.Unmarshal(rawData, &data); err != nil {
		log.Printf("CoffeeSt tod ERROR at unmarshall %v", err)
	}
	return d.NewOrderData(data)
}

func NewCoffeeOrderFromMap(data map[string]interface{}) (*CoffeeOrder, error) {
	rawData, err := json.Marshal(data)
	if err != nil {
		log.Printf("CoffeeSt ncofm ERROR at marshall %v", err)
		return nil, err
	}

	result := CoffeeOrder{}
	if err = json.Unmarshal(rawData, &result); err != nil {
		log.Printf("CoffeeSt ncofm ERROR at unmarshall %v", err)
		return nil, err
	}
	return &result, nil
}

func (co CoffeeOrder) ToAdditionalMessageData() []d.AdditionalDataElement {
	return []d.AdditionalDataElement{
		d.AdditionalDataElement{Key:"drink", Value:co.Drink, Name:"Напиток"},
		d.AdditionalDataElement{Key:"additive", Value:co.Additive, Name:"Добавка"},
		d.AdditionalDataElement{Key:"syrup", Value:co.Syrup, Name:"Сироп"},
		d.AdditionalDataElement{Key:"sugar", Value:co.Syrup, Name:"Сахар"},
		d.AdditionalDataElement{Key:"bake", Value:co.Bake, Name:"Выпечка"},
		d.AdditionalDataElement{Key:"count", Value:co.Count, Name:"Количество"},
		d.AdditionalDataElement{Key:"to_time", Value:co.ToTime, Name:"Ко времени"},
	}
}

type CoffeeConfigHandler struct {
	d.DbHelper
	Configuration *mgo.Collection
}

func (cch *CoffeeConfigHandler) ensureIndexes() {
	configCollection := cch.Session.DB(cch.DbName).C("coffee_config")
	configCollection.EnsureIndex(mgo.Index{
		Key:        []string{"name"},
		Background: true,
		DropDups:   true,
		Unique: true,
	})

	cch.Configuration = configCollection
}

func NewCoffeeConfigHandler(main_db *d.MainDb) *CoffeeConfigHandler {
	result := &CoffeeConfigHandler{DbHelper:main_db.DbHelper}
	result.ensureIndexes()
	return result
}

func (cch *CoffeeConfigHandler) AddDrink(name, drink string) error {
	err := cch.Configuration.Update(bson.M{"name":name}, bson.M{"$addToSet":bson.M{"drinks":drink}})
	return err
}
func (cch *CoffeeConfigHandler) RemoveDrink(name, drink string) error {
	err := cch.Configuration.Update(bson.M{"name":name}, bson.M{"$pull":bson.M{"drinks":drink}})
	return err
}

func (cch *CoffeeConfigHandler) AddBake(name, bake string) error {
	err := cch.Configuration.Update(bson.M{"name":name}, bson.M{"$addToSet":bson.M{"bakes":bake}})
	return err
}
func (cch *CoffeeConfigHandler) RemoveBake(name, bake string) error {
	err := cch.Configuration.Update(bson.M{"name":name}, bson.M{"$pull":bson.M{"bakes":bake}})
	return err
}

func (cch *CoffeeConfigHandler) AddVolume(name, volume string) error {
	err := cch.Configuration.Update(bson.M{"name":name}, bson.M{"$addToSet":bson.M{"volumes":volume}})
	return err
}
func (cch *CoffeeConfigHandler) RemoveVolume(name, volume string) error {
	err := cch.Configuration.Update(bson.M{"name":name}, bson.M{"$pull":bson.M{"volumes":volume}})
	return err
}

func (cch *CoffeeConfigHandler) AddAdditive(name, additive string) error {
	err := cch.Configuration.Update(bson.M{"name":name}, bson.M{"$addToSet":bson.M{"additives":additive}})
	return err
}

func (cch *CoffeeConfigHandler) RemoveAdditive(name, additive string) error {
	err := cch.Configuration.Update(bson.M{"name":name}, bson.M{"$pull":bson.M{"additives":additive}})
	return err
}

func (cch *CoffeeConfigHandler) GetConfig(name string) (*CoffeeHouseConfiguration, error) {
	result := CoffeeHouseConfiguration{}
	err := cch.Configuration.Find(bson.M{"name":name}).One(&result)
	if err == mgo.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else {
		return &result, nil
	}

}

func (cch *CoffeeConfigHandler) LoadFromConfig(conf configuration.CoffeeConfig) (*CoffeeHouseConfiguration, error) {
	chc := CoffeeHouseConfiguration{}
	err := cch.Configuration.Find(bson.M{"name":conf.Name}).One(&chc)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err

	} else {
		newChc := CoffeeHouseConfiguration{Name: conf.Name, Additives:conf.Additives, Bakes:conf.Bakes, Drinks:conf.Drinks, Syrups:conf.Syrups, Volumes:conf.Volumes}
		if err != mgo.ErrNotFound {
			cch.Configuration.Remove(bson.M{"name":conf.Name})
		}
		err := cch.Configuration.Insert(newChc)
		return &newChc, err
	}

}
