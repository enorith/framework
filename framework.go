package framework

import (
	"fmt"
	env "github.com/enorith/environment"
	"github.com/enorith/framework/cache"
	"github.com/enorith/framework/database"
	"github.com/enorith/framework/http"
	"github.com/enorith/framework/http/router"
	"github.com/enorith/framework/kernel"
	"github.com/enorith/framework/redis"
	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp"
	"log"
	net "net/http"
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
	app.RegisterServiceProvider(&database.ServiceProvider{})
}

func registerCustomProviders(app *kernel.Application) {
	for _, v := range providers {
		app.RegisterServiceProvider(v)
	}
}

// Serve your application
func Serve(addr string, routeRegister RouteHolder, structure StandardAppStructure) {
	app := Bootstrap(structure)
	httpKernel := http.NewKernel(app)
	httpKernel.SetMiddlewareGroup(structure.MiddlewareGroup)
	httpKernel.SetMiddleware(structure.Middleware)
	defer defers(app)
	routeRegister(httpKernel.Wrapper(), httpKernel)

	fmt.Println(fmt.Sprintf("Server open at [%s]", addr))

	var err error

	if httpKernel.Handler == http.HandlerFastHttp {
		err = GetFastHttpServer(httpKernel).
			ListenAndServe(addr)

	} else if httpKernel.Handler == http.HandlerNetHttp {
		err = net.ListenAndServe(addr, httpKernel)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func GetFastHttpServer(kernel *http.Kernel) *fasthttp.Server {

	return &fasthttp.Server{
		Handler:      kernel.FastHttpHandler,
		Concurrency:  kernel.RequestCurrency,
		TCPKeepalive: kernel.IsKeepAlive(),
	}
}

func defers(app *kernel.Application) {
	app.RunDefers()
}
