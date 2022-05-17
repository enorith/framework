package session

import (
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
)

var MaxHistoryPath = 64

type History []string

func (History) SessionKey() string {
	return "_hps"
}

func (hp History) Latest() string {
	l := len(hp)
	if l < 1 {
		return ""
	}
	return hp[l-1]
}

func (hp History) Last() string {
	l := len(hp)
	if l < 2 {
		return ""
	}

	return hp[l-2]
}

func (hp History) RedirectBack(r contracts.RequestContract) contracts.ResponseContract {
	return content.Redirect(r, hp.Last(), 301)
}
