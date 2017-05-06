package pb

import (
	"gopkg.in/fatih/color.v1"
	"math/rand"
	"sync"
	"text/template"
)

var templateCacheMu sync.Mutex
var templateCache = make(map[string]*template.Template)

var defaultTemplateFuncs = template.FuncMap{
	// colors
	"black":    color.New(color.FgBlack).SprintFunc(),
	"red":      color.New(color.FgRed).SprintFunc(),
	"green":    color.New(color.FgGreen).SprintFunc(),
	"yellow":   color.New(color.FgYellow).SprintFunc(),
	"blue":     color.New(color.FgBlue).SprintFunc(),
	"magenta":  color.New(color.FgMagenta).SprintFunc(),
	"cyan":     color.New(color.FgCyan).SprintFunc(),
	"white":    color.New(color.FgWhite).SprintFunc(),
	"rndcolor": rndcolor,
	"rnd":      rnd,
}

func getTemplate(tmpl string, args map[string]Element) (t *template.Template, err error) {
	// use cache only with std elements
	var cache = args == nil || len(args) == 0
	if cache {
		templateCacheMu.Lock()
		defer templateCacheMu.Unlock()
		t = templateCache[tmpl]
		if t != nil {
			// found in cache
			return
		}
	}
	t = template.New("")
	fillTemplateFuncs(t, args)
	_, err = t.Parse(tmpl)
	if err != nil {
		t = nil
		return
	}
	if cache {
		templateCache[tmpl] = t
	}
	return
}

func fillTemplateFuncs(t *template.Template, args map[string]Element) {
	t.Funcs(defaultTemplateFuncs)
	emf := make(template.FuncMap)
	elementsM.Lock()
	for k, v := range elements {
		emf[k] = v
	}
	elementsM.Unlock()
	if args != nil {
		for k, v := range args {
			emf[k] = v
		}
	}
	t.Funcs(emf)
	return
}

func rndcolor(s string) string {
	c := rand.Intn(int(color.FgWhite-color.FgBlack)) + int(color.FgBlack)
	return color.New(color.Attribute(c)).Sprint(s)
}

func rnd(args ...string) string {
	if len(args) == 0 {
		return ""
	}
	return args[rand.Intn(len(args))]
}
