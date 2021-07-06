package language

import (
	"io/fs"
	"strings"

	"github.com/enorith/config"
	"github.com/enorith/framework"
	"github.com/enorith/language"
)

type LangService struct {
	locale string
}

func (s *LangService) Register(app *framework.Application) {
	if s.locale != "" {
		language.DefaultLanguage = s.locale
	}

	de, e := fs.ReadDir(app.AssetFS(), "lang")
	if e == nil {
		for _, d := range de {
			lang := d.Name()
			if d.IsDir() {
				langPath := "lang/" + d.Name()
				del, e := fs.ReadDir(app.AssetFS(), langPath)
				if e == nil {
					for _, langF := range del {
						if !langF.IsDir() {
							filename := langF.Name()
							key := strings.Split(filename, ".")[0]
							var data map[string]string
							config.UnmarshalFS(app.AssetFS(), langPath+"/"+filename, &data)
							language.Register(key, lang, data)
						}
					}
				}
			}
		}
	}
}

func (s *LangService) Boot(app *framework.Application) {

}

func NewService(locale ...string) *LangService {

	var l string

	if len(locale) > 0 {
		l = locale[0]
	}

	return &LangService{
		locale: l,
	}
}
