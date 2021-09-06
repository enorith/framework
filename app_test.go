package framework_test

import (
	"os"
	"testing"

	"github.com/enorith/authenticate"
	"github.com/enorith/framework"
	"github.com/enorith/framework/authentication"
	"github.com/enorith/framework/cache"
	"github.com/enorith/framework/database"
	"github.com/enorith/framework/redis"
	"github.com/enorith/http/contracts"
)

type UserProvider struct {
}

func (up UserProvider) FindUserById(_ authenticate.UserIdentifier) (authenticate.User, error) {
	panic("not implemented") // TODO: Implement
}

func (up UserProvider) Attempt(r contracts.RequestContract) (authenticate.User, error) {
	panic("not implemented") // TODO: Implement
}

func TestBootstrap(t *testing.T) {
	app := framework.NewApp(os.DirFS("test_assets/config"))
	app.Register(cache.Service{})
	app.Register(database.NewService())
	app.Register(redis.Service{})
	authentication.AuthManager.WithProvider("users", UserProvider{})
	app.Register(authentication.NewAuthService())
	e := app.Bootstrap()
	if e != nil {
		t.Fatal(e)
	}

	if app.GetConfig().Env != "test" {
		t.Log(app.GetConfig())
		t.Fail()
	}

	t.Log(app.GetConfig())
}
