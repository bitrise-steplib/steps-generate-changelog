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

type changelog struct {
	Commits     []git.Commit
	CurrentDate time.Time
}

func changelogContent(commits []git.Commit) (string, error) {
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Date.After(commits[j].Date)
	})
	chlog := changelog{
		Commits:     commits,
		CurrentDate: time.Now(),
	}

	tmplStr := changelogTmplStr
	tmpl := template.New("changelog_content").Funcs(tmplFuncMap)
	tmpl, err := tmpl.Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buff bytes.Buffer
	if err := tmpl.Execute(&buff, chlog); err != nil {
		return "", err
	}

	return buff.String(), nil
}
