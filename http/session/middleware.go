package session

import (
	"net/http"

	"github.com/enorith/http/contracts"
	"github.com/enorith/http/pipeline"
	"github.com/enorith/session"
)

type Middleware struct {
	manager    *session.Manager
	id         string
	cookieName string
}

func (m Middleware) Handle(r contracts.RequestContract, next pipeline.PipeHandler) contracts.ResponseContract {
	m.manager.Start(m.id)
	resp := next(r)
	if rc, ok := resp.(contracts.WithResponseCookies); ok {
		rc.SetCookie(&http.Cookie{
			Name:  m.cookieName,
			Value: m.id,
		})
	}

	m.manager.Save(m.id)

	return resp
}
