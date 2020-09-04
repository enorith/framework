package validation

type WithValidation interface {
	Rules() map[string][]interface{}
}
