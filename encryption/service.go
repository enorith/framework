package encryption

import (
	"reflect"

	"github.com/enorith/container"
	"github.com/enorith/framework"
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

	app.Bind(func(ioc container.Interface) {
		ioc.BindFunc(EncrypterType, func(c container.Interface) (interface{}, error) {
			return DefaultEncrypter, nil
		}, true)
	})

	return nil
}
