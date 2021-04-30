package framework

import (
	"fmt"

	env "github.com/enorith/environment"
	"github.com/enorith/framework/cache"
	"github.com/enorith/framework/http"
	"github.com/enorith/framework/http/router"
	"github.com/enorith/framework/kernel"
	"github.com/enorith/framework/redis"
	"github.com/joho/godotenv"
)

var (
	providers []kernel.ServiceProvider
)

//RegisterProviders register custom service providers
func RegisterProviders(ps []kernel.ServiceProvider) {
	providers = ps
}

type RouteHolder func(ro *router.Wrapper, k *http.Kernel)

// Bootstrap your application
func Bootstrap(structure StandardAppStructure) *kernel.Application {
	godotenv.Load(fmt.Sprintf("%s/.env", structure.BasePath))
	kernel.Timezone = env.GetString("APP_TIMEZONE", "UTC")
	app := kernel.NewApp(env.GetString("APP_ENV", "development"), env.GetBoolean("APP_DEBUG", true), structure.BasePath)
	registerDefaultProviders(app)
	registerCustomProviders(app)
	app.Bootstrap()
	return app
}

func registerDefaultProviders(app *kernel.Application) {
	app.RegisterServiceProvider(&redis.ServiceProvider{})
	app.RegisterServiceProvider(&cache.ServiceProvider{})
}

func registerCustomProviders(app *kernel.Application) {
	for _, v := range providers {
		app.RegisterServiceProvider(v)
	}
}
