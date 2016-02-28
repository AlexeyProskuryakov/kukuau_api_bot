package infinity

import (
	"testing"
	c "msngr/configuration"
	"log"
	t "msngr/taxi"
)
func testEq(a, b []t.Markup) bool {

    if a == nil && b == nil {
        return true;
    }

    if a == nil || b == nil {
        return false;
    }

    if len(a) != len(b) {
        return false
    }

    for i := range a {
        if a[i].ID != b[i].ID {
            return false
        }
    }

    return true
}
func TestRequest(t *testing.T){
	conf := c.ReadTestConfigInRecursive()
	tconf := conf.Taxis["fake"]
	inf := initInfinity(tconf.Api, "fake")
	mrkps := inf.Markups()
	log.Printf("mrkps before session: %v", mrkps)

	inf.LoginResponse.SessionID = "foo"
	new_mrkps := inf.Markups()
	log.Printf("mrkps after session: %v", mrkps)

	if !testEq(mrkps, new_mrkps){
		t.Error("session must refresh")
	}
}