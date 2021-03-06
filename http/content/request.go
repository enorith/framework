package content

import (
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/enorith/framework/contracts"
	. "github.com/enorith/framework/http/contracts"
	"github.com/enorith/supports/byt"
	"strconv"
)

type Param string

func (p Param) Value() string {
	return string(p)
}

type ParamInt64 int64

func (p ParamInt64) Value() int64 {
	return int64(p)
}

type ParamInt int

func (p ParamInt) Value() int {
	return int(p)
}

type ParamUint64 uint64

func (p ParamUint64) Value() uint64 {
	return uint64(p)
}

type SimpleParamRequest struct {
	params      map[string][]byte
	paramsSlice [][]byte
	user        contracts.User
}

func (shr *SimpleParamRequest) Params() map[string][]byte {
	return shr.params
}

func (shr *SimpleParamRequest) ParamsSlice() [][]byte {
	return shr.paramsSlice
}

func (shr *SimpleParamRequest) Param(key string) string {
	return string(shr.params[key])
}

func (shr *SimpleParamRequest) ParamBytes(key string) []byte {
	return shr.params[key]
}

func (shr *SimpleParamRequest) ParamInt64(key string) (int64, error) {
	param, ok := shr.params[key]
	if !ok {
		return 0, errors.New(fmt.Sprintf("can not get param [%s]", key))
	}

	return byt.ToInt64(param)
}

func (shr *SimpleParamRequest) ParamUint64(key string) (uint64, error) {
	param, ok := shr.params[key]
	if !ok {
		return 0, errors.New(fmt.Sprintf("can not get param [%s]", key))
	}

	return byt.ToUint64(param)
}

func (shr *SimpleParamRequest) ParamInt(key string) (int, error) {
	param, ok := shr.params[key]
	if !ok {
		return 0, errors.New(fmt.Sprintf("can not get param [%s]", key))
	}

	return strconv.Atoi(string(param))
}

func (shr *SimpleParamRequest) SetParams(params map[string][]byte) {
	shr.params = params
}

func (shr *SimpleParamRequest) SetParamsSlice(paramsSlice [][]byte) {
	shr.paramsSlice = paramsSlice
}

func (shr *SimpleParamRequest) User() contracts.User {
	return shr.user
}

func (shr *SimpleParamRequest) SetUser(u contracts.User) {
	shr.user = u
}

func GetJsonValue(r RequestContract, key string) []byte {
	if r.RequestWithJson() {
		val, _, _, _ := jsonparser.Get(r.GetContent(), key)

		return val
	}

	return nil
}

type Request struct {
	RequestContract
}
