package taxi

import (
	"testing"
	"strings"
	c "msngr/configuration"
	"fmt"
)

func TestApplyTransformation(t *testing.T) {
	tr := []c.Transformation{c.Transformation{Field:"phone", RegexCode:"\\+?7([\\d]{8,10})", To:"8$1"}}

	o := NewOrderInfo{Phone:"+79231378736", Notes:"test banana"}
	to := ApplyTransforms(&o, tr)
	if !strings.HasPrefix(to.Phone, "8") {
		t.Error(fmt.Sprintf("Phone not contains 8 at first cur: %v", to.Phone))
	}

	o = NewOrderInfo{Phone:"79992095923"}
	to = ApplyTransforms(&o, tr)
	if !strings.EqualFold(to.Phone, "89992095923") {
		t.Error("If phone must be changed if not have +")
	}

	o = NewOrderInfo{Phone:"9992095923"}
	to = ApplyTransforms(&o, tr)
	if !strings.EqualFold(to.Phone, o.Phone) {
		t.Error("If phone is not valid to regexp not change it! Non 7")
	}

	o = NewOrderInfo{Phone:"79992"}
	to = ApplyTransforms(&o, tr)
	if !strings.EqualFold(to.Phone, o.Phone) {
		t.Error("If phone is not valid to regexp not change it! So little")
	}

	o = NewOrderInfo{Phone:"test fooo bar"}
	to = ApplyTransforms(&o, tr)
	if !strings.EqualFold(to.Phone, o.Phone) {
		t.Error("If phone is not valid to regexp not change it! Chars")
	}

	o = NewOrderInfo{Phone:"   "}
	to = ApplyTransforms(&o, tr)
	if !strings.EqualFold(to.Phone, o.Phone) {
		t.Error("If phone is not valid to regexp not change it! Spaces")
	}
}