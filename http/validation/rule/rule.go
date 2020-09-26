package rule

type Rule interface {
	Passes(attribute string, input interface{}) (success bool, skipAll bool)
	Message() string
}
