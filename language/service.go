package language

import (
	"github.com/enorith/framework/kernel"
	"github.com/enorith/framework/kernel/config"
	"github.com/enorith/language"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

type Service struct {
}

func (s Service) Register(app *kernel.Application) {

	base := filepath.Join(app.GetBasePath(), "lang")
	fs, e := ioutil.ReadDir(base)

	if e != nil {
		log.Printf("unable load languages %s", e)
	}

	for _, f := range fs {
		lang := f.Name()
		langPath := filepath.Join(base, lang)
		fss, e := ioutil.ReadDir(langPath)

		if e != nil {
			log.Printf("unable load languages %s path: %s", e, langPath)
			continue
		}

		for _, ff := range fss {
			filename := ff.Name()
			if ff.IsDir() {
				log.Printf("unable load languages %s path: %s", filename, langPath)
				break
			}
			key := strings.Split(filename, ".")[0]
			file := filepath.Join(langPath, filename)
			var data map[string]string
			e = config.LoadTo(file, &data)
			if e != nil {
				log.Printf("unable load languages %s file: %s", key, file)
				break
			}
			language.Register(key, lang, data)
		}
	}

}

func (s Service) Boot(app *kernel.Application) {

}
