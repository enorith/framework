package framework_test

import (
	"os"
	"testing"

	"github.com/enorith/framework"
	"github.com/enorith/framework/cache"
	"github.com/enorith/framework/database"
)

func TestBootstrap(t *testing.T) {
	app := framework.NewApp(os.DirFS("test_assets/config"))
	app.Register(&cache.Service{})
	app.Register(&database.Service{})
	_, e := app.Bootstrap()
	if e != nil {
		t.Fatal(e)
	}

	if app.GetConfig().Env != "test" {
		t.Log(app.GetConfig())
		t.Fail()
	}
	t.Log(app.GetConfig())
}
