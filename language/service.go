package language

import (
	"io/fs"
	"strings"

	"github.com/enorith/config"
	"github.com/enorith/framework/kernel"
	"github.com/enorith/language"
)

type LangService struct {
}

func (s LangService) Register(app *kernel.Application) {
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

func (s LangService) Boot(app *kernel.Application) {

}
