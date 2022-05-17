package session

import (
	"net/http"

	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/pipeline"
	"github.com/enorith/session"
)

type Middleware struct {
	manager     *session.Manager
	id          string
	cookieName  string
	maxLifeTime int
	domain      string
	path        string
	secure      bool
	httpOnly    bool
	sameSite    http.SameSite
}

func (m Middleware) Handle(r contracts.RequestContract, next pipeline.PipeHandler) contracts.ResponseContract {
	err := m.manager.Start(m.id)
	if err != nil {
		return content.ErrResponseFromError(err, 500, nil)
	}
	ses := m.manager.Get(m.id)
	hp := make(History, 0)
	ses.Get(hp.SessionKey(), &hp)
	path := string(r.GetUri())
	if hp.Latest() != path {
		hp = append(hp, string(r.GetUri()))
		if len(hp) > MaxHistoryPath {
			hp = hp[1:]
		}
		ses.Set(hp.SessionKey(), hp)
	}

	resp := next(r)
	if rc, ok := resp.(contracts.WithResponseCookies); ok {
		rc.SetCookie(&http.Cookie{
			Name:     m.cookieName,
			Value:    m.id,
			MaxAge:   m.maxLifeTime,
			Domain:   m.domain,
			Path:     m.path,
			Secure:   m.secure,
			HttpOnly: m.httpOnly,
			SameSite: m.sameSite,
		})
	}

	m.manager.Save(m.id)

	return resp
}
