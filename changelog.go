package main

import (
	"bytes"
	"html/template"
	"sort"
	"time"

	"github.com/godrei/steps-generate-changelog/git"
)

// ChangelogTmplStr ...
const ChangelogTmplStr = `{{range .Commits}}* [{{firstChars .Hash 7}}] {{.Message}}
{{end}}`

var tmplFuncMap = template.FuncMap{
	"firstChars": func(str string, length int) string {
		if len(str) < length {
			return str
		}

		return str[0:length]
	},
}

// Changelog ..
type Changelog struct {
	Commits     []git.Commit
	Version     string
	CurrentDate time.Time
}

// ChangelogContent ...
func ChangelogContent(commits []git.Commit, version string) (string, error) {
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Date.After(commits[j].Date)
	})
	changelog := Changelog{
		Commits:     commits,
		Version:     version,
		CurrentDate: time.Now(),
	}

	tmplStr := ChangelogTmplStr
	tmpl := template.New("changelog_content").Funcs(tmplFuncMap)
	tmpl, err := tmpl.Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buff bytes.Buffer
	err = tmpl.Execute(&buff, changelog)
	if err != nil {
		return "", err
	}

	return buff.String(), nil
}
