package auth

import (
	"errors"
	"fmt"
	db "github.com/enorith/database"
	. "github.com/enorith/framework/contracts"
	"github.com/enorith/framework/database"
	"github.com/enorith/framework/http/contracts"
	"sync"
)

type AuthenticatorRegister func() Authenticator

var (
	authenticatorRegisters = make(map[string]AuthenticatorRegister)
	am                     = new(sync.RWMutex)
)

var (
	DefaultUser      = new(GenericUser)
	DefaultUserTable = "users"
	DefaultProvider  = "database"
	DefaultDriver    = "jwt"
	Auth             *AuthenticateManager
)

type Authenticator interface {
	Guard(r contracts.RequestContract) (User, error)
	Check(r contracts.RequestContract) bool
	Auth(u User) contracts.ResponseContract
}

type GenericAuthenticator struct {
	provider     UserProvider
	providerName string
}

func (a *GenericAuthenticator) GetUserProvider() UserProvider {
	if a.provider == nil {
		if register, ok := getProviderRegister(a.providerName); ok {
			a.provider = register()
			return a.provider
		}
	}

	return a.provider
}

type GenericUser struct {
	item *db.CollectionItem
}

func (u *GenericUser) MarshalJSON() ([]byte, error) {
	return u.item.MarshalJSON()
}

func (u *GenericUser) UserIdentifier() uint64 {
	id, _ := u.item.GetUint(u.UserIdentifierName())

	return id
}

func (u *GenericUser) UserIdentifierName() string {
	return "id"
}

func (u *GenericUser) CloneUser() User {
	return &GenericUser{}
}

func (u *GenericUser) Unmarshal(data *db.CollectionItem) {
	u.item = data
}

type AuthenticateManager struct {
	authenticator     Authenticator
	authenticatorName string
}

func (m *AuthenticateManager) Guard(r contracts.RequestContract) (User, error) {
	return m.authenticator.Guard(r)
}

func (m *AuthenticateManager) Check(r contracts.RequestContract) bool {
	return m.authenticator.Check(r)
}

func (m *AuthenticateManager) Auth(u User) contracts.ResponseContract {
	return m.authenticator.Auth(u)
}

func (m *AuthenticateManager) Use(name string) error {
	if register, ok := getRegister(name); ok {
		if name != m.authenticatorName {
			m.authenticator = register()
			m.authenticatorName = name
		}
		return nil
	}

	return errors.New(fmt.Sprintf("auth: authenticator [%s] not registerd", name))
}

func getRegister(name string) (func() Authenticator, bool) {
	am.RLock()
	defer am.RUnlock()
	auth, ok := authenticatorRegisters[name]

	return auth, ok
}

func RegisterDriver(name string, register AuthenticatorRegister) {
	am.Lock()
	authenticatorRegisters[name] = register
	am.Unlock()
}

func init() {
	RegisterProvider("database", func() UserProvider {
		return &DatabaseUserProvider{
			user:    DefaultUser,
			table:   DefaultUserTable,
			builder: database.NewDefaultBuilder(),
		}
	})
	RegisterDriver("jwt", func() Authenticator {
		return NewJwtAuthFromDefault()
	})
	Auth = &AuthenticateManager{}
	Auth.Use(DefaultDriver)
}
