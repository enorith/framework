package validation

import (
	"github.com/enorith/framework/http/contracts"
	"github.com/enorith/framework/http/validation/rule"
	"strings"
	"sync"
)

type WithValidation interface {
	Rules() map[string][]interface{}
}

var DefaultValidator *Validator

type ValidateError map[string]string

func (v ValidateError) StatusCode() int {
	return 422
}

func (v ValidateError) Error() string {
	var first string

	for _, s := range v {
		first = s
		break
	}

	return first
}

func Register(name string, register RuleRegister) {
	DefaultValidator.Register(name, register)
}

type RuleRegister func(r contracts.RequestContract, args ...string) rule.Rule

type Validator struct {
	registers map[string]RuleRegister
	mu        sync.RWMutex
}

func (v *Validator) Register(name string, register RuleRegister) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.registers[name] = register
}

func (v *Validator) GetRule(name string) (RuleRegister, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	r, ok := v.registers[name]

	return r, ok
}

func (v *Validator) Passes(req contracts.RequestContract, attribute string, input interface{}, rules []interface{}) (errors []string) {
	for _, rl := range rules {
		if s, ok := rl.(string); ok {
			ss := strings.Split(s, ":")
			var args []string
			if len(ss) > 1 {
				args = strings.Split(ss[1], ",")
			}
			r, exist := v.GetRule(ss[0])
			if exist {
				inputRule := r(req, args...)
				success, skip := inputRule.Passes(attribute, input)
				if !success {
					errors = append(errors, inputRule.Message())
				}

				if skip {
					break
				}
			}
		} else if rr, ok := rl.(rule.Rule); ok {
			success, skip := rr.Passes(attribute, input)
			if skip {
				break
			}

			if !success {
				errors = append(errors, rr.Message())
			}
		}
	}

	return
}

func init() {
	DefaultValidator = &Validator{registers: map[string]RuleRegister{}, mu: sync.RWMutex{}}
	Register("required", func(r contracts.RequestContract, args ...string) rule.Rule {
		return rule.Required{}
	})
}
