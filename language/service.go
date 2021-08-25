package language

import (
	"io/fs"
	"strings"

	"github.com/enorith/config"
	"github.com/enorith/container"
	"github.com/enorith/framework"
	"github.com/enorith/http/contracts"
	"github.com/enorith/language"
)

var Dir = "."

type LangService struct {
	locale string
	fs     fs.FS
}

func (s *LangService) Register(app *framework.App) error {
	if s.locale != "" {
		language.DefaultLanguage = s.locale
	}

	de, e := fs.ReadDir(s.fs, Dir)
	if e == nil {
		for _, d := range de {
			lang := d.Name()
			if d.IsDir() {
				langPath := d.Name()
				del, e := fs.ReadDir(s.fs, langPath)
				if e != nil {
					return e
				}
				for _, langF := range del {
					if !langF.IsDir() {
						filename := langF.Name()
						key := strings.Split(filename, ".")[0]
						var data map[string]string
						config.UnmarshalFS(s.fs, langPath+"/"+filename, &data)
						language.Register(key, lang, data)
					}
				}
			}
		}
	}

	return e
}

//Lifetime container callback
// usually register request lifetime instance to IoC-Container (per-request unique)
// this function will run before every request
func (s *LangService) Lifetime(ioc container.Interface, request contracts.RequestContract) {
}

func NewService(fs fs.FS, locale ...string) *LangService {

	var l string

	if len(locale) > 0 {
		l = locale[0]
	}

	return &LangService{
		locale: l,
		fs:     fs,
	}
}
