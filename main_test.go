package main

import (
	"strings"
	"testing"
)

type mockChangelogPrinter struct{}

func (mcp *mockChangelogPrinter) content() (string, error) {
	return "line 1\nline 2", nil
}

type mockTooManyChangelogPrinter struct{}

func (mcp *mockTooManyChangelogPrinter) content() (string, error) {
	return strings.Repeat("a", 25*1024), nil
}

func Test_processChangelog(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name string
		cp   changelogPrinter
		c    Config
	}{
		{
			cp: &mockChangelogPrinter{},
			c:  Config{NewVersion: "0.0.1", ChangelogPath: "/tmp/changelog.md", WorkDir: "."},
		},
		{
			cp: &mockTooManyChangelogPrinter{},
			c:  Config{NewVersion: "0.0.1", ChangelogPath: "/tmp/changelog.md", WorkDir: "."},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processChangelog(tt.cp, tt.c)
		})
	}
}
