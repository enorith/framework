package rule

import "github.com/enorith/language"

type Required struct {
}

func (r Required) Passes(attribute string, input interface{}) (success bool, skipAll bool) {

	return false, false
}

func (r Required) Message() string {
	s, _ := language.T("validation", "required")

	return s
}
