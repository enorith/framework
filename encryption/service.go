package encryption

import (
	"reflect"

	"github.com/enorith/container"
	"github.com/enorith/framework"
	"github.com/enorith/http/contracts"
)

var DefaultEncrypter Encrypter

var EncrypterType = reflect.TypeOf((*Encrypter)(nil)).Elem()

type Service struct {
}

//Register service when app starting, before http server start
// you can configure service, prepare global vars etc.
// running at main goroutine
func (s Service) Register(app *framework.App) error {

	DefaultEncrypter = NewAesEncrypter([]byte(app.GetConfig().Key))

	return nil
}

//Lifetime container callback
// usually register request lifetime instance to IoC-Container (per-request unique)
// this function will run before every request handling
func (s Service) Lifetime(ioc container.Interface, request contracts.RequestContract) {
	ioc.BindFunc(EncrypterType, func(c container.Interface) (interface{}, error) {
		return DefaultEncrypter, nil
	}, true)
}
