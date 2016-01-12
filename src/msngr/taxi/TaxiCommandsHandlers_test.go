package taxi

import (
	"testing"
	"strings"

	c "msngr/configuration"
	d "msngr/db"
	s "msngr/structs"
	tst "msngr/test"

)

type FakeAddressSupplier struct {

}

func (f FakeAddressSupplier)IsConnected() bool {
	return true
}

func (f FakeAddressSupplier)AddressesAutocomplete(query string) AddressPackage {
	return AddressPackage{}
}

func NewFakeAddressSupplier() AddressSupplier {
	return FakeAddressSupplier{}
}

func TestOrderPrice(t *testing.T) {
	config := c.ReadConfig()
	taxi_conf := config.Taxis["fake"]


	external_api := GetFakeAPI(taxi_conf.Api)
	external_address_supplier := NewFakeAddressSupplier()

	db := d.NewMainDb(config.Main.Database.ConnString, config.Main.Database.Name + "_test")
	gah := NewGoogleAddressHandler(config.Main.GoogleKey, taxi_conf.GeoOrbit, external_address_supplier)

	apiMixin := ExternalApiMixin{API: external_api}
	botContext := FormTaxiBotContext(&apiMixin, db, taxi_conf, gah, NewCarsCache(&apiMixin))

	new_order_package := tst.ReadTestFile("new_order_ok.json")

	nop, ok := botContext.Message_commands["new_order"]
	if !ok {
		t.Fail()
	}

	s.StartAfter(botContext.Check, func() {
		result := nop.ProcessMessage(new_order_package)

		price_in_message := strings.Contains(result.Body, "Стоимость")
		if taxi_conf.Api.NotSendPrice && price_in_message {
			t.Errorf("in config not_send_price is true but in result message price is sended: \n%+v", result.Body)
		}

		order_created_in_message := strings.Contains(result.Body, "Ваш заказ создан!")
		if !order_created_in_message {
			t.Errorf("[Ваш заказ создан!] not in result message:\n%+v", result.Body)
		}
	})
}
