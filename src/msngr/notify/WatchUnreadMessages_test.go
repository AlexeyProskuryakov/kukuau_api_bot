package notify

import (
	"testing"
	"msngr/configuration"
	"msngr/db"
	"time"
)

func TestWum(t *testing.T) {
	conf := configuration.ReadTestConfigInRecursive()
	db := db.NewMainDb(conf.Main.Database.ConnString, conf.Main.Database.Name)
	wum := NewWatchManager(db, "test")
	go wum.WatchUnreadMessages()

	for i, el := range []string{"one", "two", "three"} {
		wum.AddConfiguration(el, "test", "key", 2)
		if len(wum.Configs) != i+1 {
			t.Error("bad configs count")
		}
		time.Sleep(10)
	}
}
