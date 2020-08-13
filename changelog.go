package main

import (
	"bytes"
	"html/template"
	"sort"
	"time"

	"github.com/bitrise-steplib/steps-generate-changelog/git"
)

const changelogTmplStr = `{{range .Commits}}* [{{firstChars .Hash 7}}] {{.Message}}
{{end}}`

var tmplFuncMap = template.FuncMap{
	"firstChars": func(str string, length int) string {
		if len(str) < length {
			return str
		}

		return str[0:length]
	},
}

type changelogPrinter interface {
	content() (string, error)
}

type changelog struct {
	Commits     []git.Commit
	Version     string
	currentDate time.Time
}

func (cl *changelog) content() (string, error) {
	sort.Slice(cl.Commits, func(i, j int) bool {
		return cl.Commits[i].Date.After(cl.Commits[j].Date)
	})

	cl.currentDate = time.Now()

	tmplStr := changelogTmplStr
	tmpl := template.New("changelog_content").Funcs(tmplFuncMap)
	tmpl, err := tmpl.Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buff bytes.Buffer
	err = tmpl.Execute(&buff, cl)
	if err != nil {
		return "", err
	}

	return buff.String(), nil
}
