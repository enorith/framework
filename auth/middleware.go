package auth

import (
	"github.com/enorith/framework/exception"
	"github.com/enorith/framework/http"
	"github.com/enorith/framework/http/content"
	"github.com/enorith/framework/http/contracts"
)

type AuthMiddleware struct {
	driver string
}

func (a *AuthMiddleware) Handle(r contracts.RequestContract, next http.PipeHandler) contracts.ResponseContract {
	if len(a.driver) > 0 {
		Auth.Use(a.driver)
	}
	user, err := Auth.Guard(r)
	if err == nil && user != nil {
		r.SetUser(user)
		return next(r)
	}

	return a.UnauthenticatedResponse()
}

func (a *AuthMiddleware) UnauthenticatedResponse() contracts.ResponseContract {

	e := exception.NewHttpException("unauthenticated", 401, 401, nil)
	return content.ErrResponse(e, 401, nil)
}

func NewAuthMiddleware(driver ...string) *AuthMiddleware {
	var d string
	if len(driver) > 0 {
		d = driver[0]
	}
	return &AuthMiddleware{d}
}
