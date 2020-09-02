package auth

import (
	"errors"
	db "github.com/enorith/database"
	. "github.com/enorith/framework/contracts"
	"github.com/enorith/framework/http/contract"
	"golang.org/x/crypto/bcrypt"
	"sync"
)

var (
	UserNameField = "username"
	PasswordField = "password"
)

type ProviderRegister func() UserProvider

var (
	providerRegister = make(map[string]ProviderRegister)
	pm               = new(sync.RWMutex)
)

func getProviderRegister(name string) (ProviderRegister, bool) {
	pm.RLock()
	defer pm.RUnlock()
	rg, ok := providerRegister[name]

	return rg, ok
}

func RegisterProvider(name string, register ProviderRegister) {
	pm.Lock()
	providerRegister[name] = register
	pm.Unlock()
}

type UserProvider interface {
	FindUserById(id uint64) (User, error)
	FindUserByRequest(r contract.RequestContract) (User, error)
}

type DatabaseUserProvider struct {
	user    User
	table   string
	builder *db.QueryBuilder
}

func (d *DatabaseUserProvider) FindUserByRequest(r contract.RequestContract) (User, error) {
	result, findErr := d.builder.From(d.table).AndWhere(UserNameField, "=", r.GetString(UserNameField)).First()
	if findErr != nil {
		return nil, findErr
	}
	pass, getErr := result.GetString(PasswordField)
	if getErr != nil {
		return nil, getErr
	}

	compareErr := bcrypt.CompareHashAndPassword([]byte(pass), r.Get(PasswordField))

	if compareErr != nil {
		return nil, compareErr
	}

	return d.itemToUser(result)
}

func (d *DatabaseUserProvider) FindUserById(id uint64) (User, error) {
	result, err := d.builder.Clone().From(d.table).AndWhere(d.user.UserIdentifierName(), "=", id).First()
	if err != nil {
		return nil, err
	}
	return d.itemToUser(result)
}

func (d *DatabaseUserProvider) itemToUser(item *db.CollectionItem) (User, error) {
	if !item.IsValid() {
		return nil, errors.New("invalid user data")
	}
	user := d.user.CloneUser()

	user.Unmarshal(item)
	return user, nil
}
