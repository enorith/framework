package authentication

import (
	"github.com/enorith/authenticate"
	"github.com/enorith/http/contracts"
)

type Attempter interface {
	Attempt(r contracts.RequestContract) (authenticate.User, error)
}

type AuthProvider interface {
	authenticate.UserProvider
	Attempter
}
