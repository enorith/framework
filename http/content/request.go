package content

import (
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/enorith/framework/contracts"
	"github.com/enorith/framework/http/contract"
	"strconv"
)

type simpleParamRequest struct {
	params map[string]string
	user   contracts.User
}

func (shr *simpleParamRequest) Params() map[string]string {
	return shr.params
}

func (shr *simpleParamRequest) Param(key string) string {
	return shr.params[key]
}

func (shr *simpleParamRequest) ParamInt64(key string) (int64, error) {
	param, ok := shr.params[key]
	if !ok {
		return 0, errors.New(fmt.Sprintf("can not get param [%s]", key))
	}

	return strconv.ParseInt(param, 10, 64)
}

func (shr *simpleParamRequest) ParamUint64(key string) (uint64, error) {
	param, ok := shr.params[key]
	if !ok {
		return 0, errors.New(fmt.Sprintf("can not get param [%s]", key))
	}

	return strconv.ParseUint(param, 10, 64)
}

func (shr *simpleParamRequest) ParamInt(key string) (int, error) {
	param, ok := shr.params[key]
	if !ok {
		return 0, errors.New(fmt.Sprintf("can not get param [%s]", key))
	}

	return strconv.Atoi(param)
}

func (shr *simpleParamRequest) SetParams(params map[string]string) {
	shr.params = params
}

func (shr *simpleParamRequest) User() contracts.User {
	return shr.user
}

func (shr *simpleParamRequest) SetUser(u contracts.User) {
	shr.user = u
}

func GetJsonValue(r contract.RequestContract, key string) []byte {
	if r.ExceptsJson() {
		val, _, _, _ := jsonparser.Get(r.GetContent(), key)

		return val
	}

	return nil
}

type Request struct {
	contract.RequestContract
}
