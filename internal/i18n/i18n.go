// Package i18n lang
package i18n

import (
	"fmt"
	"go-image/internal/config"
	"go-image/internal/util/utilconfig"
	"maps"
	"slices"
	"strings"
)

// TextLang text lang
type TextLang interface {
	Lang(text string, args ...any) string
}

// UserLang for single lang
type UserLang interface {
	Lang(text string, args ...any) string
	LangCode() string
	LangWords() map[string]string
}

// AppLang all langs
type AppLang interface {
	UserLang(code string) UserLang
	HasLang(code string) bool
}

func MustNewAppLang(config *config.AppConfig) AppLang {

	res := &appLang{
		langs: config.Lang.Langs,
		data:  map[string]map[string]string{},
	}

	if len(res.langs) == 0 {
		panic(fmt.Errorf("error no any lang in app config")) // Fatal
	}

	// config.ConfigPath == []string{".", os.Getenv("APP_CONFIG"), flagAppConfig}
	res.loadFromConfigFiles(config.ConfigPath, res.langs)

	for _, k := range res.langs {
		name := res.data[k][k]
		res.names = append(res.names, name)
	}

	res.defaultLang = res.langs[0]

	return res
}

type appLang struct {
	defaultLang string
	langs       []string                     // lang codes [en,es]
	names       []string                     // lang names [English,Spanish]
	data        map[string]map[string]string // words map {en{"Sing in":"Login"},es{"Sing in":"Iniciar sesión"}}
}
type userLang struct {
	code string
	data map[string]string
}

func (x *appLang) HasLang(code string) bool {
	return slices.Contains(x.langs, code)
}

// UserLang get lang words
func (x *appLang) UserLang(code string) UserLang {

	if !slices.Contains(x.langs, code) {
		code = x.defaultLang
	}

	data := x.data[code]

	return &userLang{
		code: code,
		data: data,
	}
}

// loadFromConfigFiles load lang data from resources if file exists
func (x *appLang) loadFromConfigFiles(configPath []string, langs []string) {

	// Initialize the result map
	result := make(map[string]map[string]string)

	// Iterate over the matched files
	for _, langCode := range langs {

		for i := 0; i < len(configPath); i++ {
			dir := configPath[i] // Directory containing the lang.*.json files
			fileName := fmt.Sprintf("lang.%s.json", langCode)

			var fileData map[string]string

			err := utilconfig.LoadConfig(&fileData, dir, fileName)
			if err != nil {
				panic(fmt.Errorf("error reading file: %v", err))
			}

			result[langCode] = fileData // override

		}
	}

	maps.Copy(x.data, result)
}

// Lang translate en-to-es Lang("Hello, {0}","Jon") to "Hola, Jon"
func (x *userLang) Lang(text string, args ...any) string {

	if x.data != nil {
		tmp := x.data[text]
		if tmp != "" {
			text = tmp
		}
	}

	if len(args) > 0 {

		for i, v := range args {
			ph := fmt.Sprintf("{%v}", i)
			str := fmt.Sprintf("%v", v)
			text = strings.ReplaceAll(text, ph, str)
		}

	}

	return text
}
func (x *userLang) LangCode() string {

	return x.code
}
func (x *userLang) LangWords() map[string]string {

	return x.data
}
