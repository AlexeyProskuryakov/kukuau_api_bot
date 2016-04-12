package coffee

import (
	d "msngr/db"
	m "msngr"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"sort"
	"msngr/configuration"
	"msngr/structs"
	"errors"
)

type CoffeeHouseConfiguration struct {
	Name      string `bson:"name"`
	Bakes     []string `bson:"bakes"`
	Drinks    []string `bson:"drinks"`
	Additives []string `bson:"additives"`
	Volumes   []string `bson:"volumes"`
}

type CoffeeOrder struct {
	Type     string
	Drink    string
	Bake     string
	Additive string
	Volume   string
}

func NewCoffeeOrderFromForm(form structs.InForm) (*CoffeeOrder, error) {
	result := CoffeeOrder{}
	if form.Name == "order_drink_form" {
		result.Type = "drink"
		drink, _ := form.GetValue("drink")
		result.Drink = drink
		additive, _ := form.GetValue("additive")
		result.Additive = additive
		volume, _ := form.GetValue("volume")
		result.Volume = volume
		return &result, nil
	}else if form.Name == "order_bake_form" {
		result.Type = "bake"
		bake, _ := form.GetValue("bake")
		result.Bake = bake
		return &result, nil
	}
	return nil, errors.New("Invalid form :( ")
}

func (co CoffeeOrder) ToOrderData() d.OrderData {
	return d.NewOrderData(map[string]interface{}{
		"type":co.Type,
		"drink":co.Drink,
		"additive":co.Additive,
		"volume":co.Volume,
		"bake":co.Bake,
	})
}

func (co CoffeeOrder) ToAdditionalMessageData() []d.AdditionalDataElement {
	return []d.AdditionalDataElement{
		d.AdditionalDataElement{Key:"drink", Value:co.Drink, Name:"Напиток"},
		d.AdditionalDataElement{Key:"additive", Value:co.Additive, Name:"Добавка"},
		d.AdditionalDataElement{Key:"volume", Value:co.Volume, Name:"Объем"},
		d.AdditionalDataElement{Key:"bake", Value:co.Bake, Name:"Выпечка"},
	}
}

func (cc CoffeeHouseConfiguration) _getFieldContent(fieldBsonName string) []string {
	v := reflect.ValueOf(cc)
	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		// gets us a StructField
		fi := typ.Field(i)
		if tagv := fi.Tag.Get("bson"); tagv == fieldBsonName {
			return v.Field(i).Interface().([]string)
		}
	}
	return []string{}
}

func (cc CoffeeHouseConfiguration) Autocomplete(q, fieldBsonName string) []string {
	content := cc._getFieldContent(fieldBsonName)
	by := m.ByFuzzyEquals{Data:content, Center:q}
	sort.Sort(by)
	return by.Data
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
	configCollection.EnsureIndex(mgo.Index{
		Key:        []string{"drinks"},
		Background: true,
		DropDups:   true,
		Unique: true,
	})

	configCollection.EnsureIndex(mgo.Index{
		Key:        []string{"bakes"},
		Background: true,
		DropDups:   true,
		Unique: true,
	})

	configCollection.EnsureIndex(mgo.Index{
		Key:        []string{"additives"},
		Background: true,
		DropDups:   true,
		Unique: true,
	})
	configCollection.EnsureIndex(mgo.Index{
		Key:        []string{"volumes"},
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
	}else if err != nil {
		return nil, err
	} else {
		return &result, nil
	}

}

func (cch *CoffeeConfigHandler) LoadFromConfig(conf configuration.CoffeeConfig) {
	f := CoffeeHouseConfiguration{}
	cch.Configuration.Find(bson.M{"name":conf.Name}).One(&f)
	if f.Name != conf.Name{
		cch.Configuration.Insert(bson.M{"name":conf.Name})
	}

	for _, bake := range conf.Bakes {
		cch.RemoveBake(conf.Name, bake)
		cch.AddBake(conf.Name, bake)
	}
	for _, drink := range conf.Drinks {
		cch.RemoveDrink(conf.Name, drink)
		cch.AddDrink(conf.Name, drink)
	}
	for _, additive := range conf.Additives {
		cch.RemoveAdditive(conf.Name, additive)
		cch.AddAdditive(conf.Name, additive)
	}
	for _, volume := range conf.Volumes {
		cch.RemoveVolume(conf.Name, volume)
		cch.AddVolume(conf.Name, volume)
	}
}
