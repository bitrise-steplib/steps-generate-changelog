package main

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/tools"
	"github.com/godrei/steps-generate-changelog/git"
	version "github.com/hashicorp/go-version"
)

func failf(format string, args ...interface{}) {
	log.Errorf(format, args...)
	os.Exit(1)
}

func releaseCommits(newVersion string) ([]git.Commit, error) {
	ver, err := version.NewVersion(newVersion)
	if err != nil {
		return nil, err
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	firstCommit, err := git.FirstCommit(dir)
	if err != nil {
		return nil, err
	}

	lastCommit, err := git.LastCommit(dir)
	if err != nil {
		return nil, err
	}

	taggedCommits, err := git.VersionTaggedCommits(dir)
	if err != nil {
		return nil, err
	}

	startCommit := firstCommit
	includeFirst := true
	endCommit := lastCommit
	if len(taggedCommits) > 0 {
		lastTaggedCommit := taggedCommits[len(taggedCommits)-1]
		lastVersion, err := version.NewVersion(lastTaggedCommit.Tag)
		if err != nil {
			return nil, err
		}

		if ver.LessThan(lastVersion) {
			// collecting previous changelogs
			return nil, fmt.Errorf("new version (%s) is less then the most recent tag (%s)", ver, lastTaggedCommit.Tag)
		} else if ver.GreaterThan(lastVersion) {
			// collecting changelog for a new version
			startCommit = taggedCommits[len(taggedCommits)-1]
			includeFirst = false
		} else if len(taggedCommits) > 1 {
			// collecting changelog between existing versions
			startCommit = taggedCommits[len(taggedCommits)-2]
			endCommit = taggedCommits[len(taggedCommits)-1]
			includeFirst = false
		}
	}

	commits, err := git.Commits(dir)
	if err != nil {
		return nil, err
	}

	var releaseCommits []git.Commit
	for _, commit := range commits {
		if commit.Date.Before(startCommit.Date) {
			continue
		}
		if !includeFirst && commit.Date.Equal(startCommit.Date) {
			continue
		}
		if commit.Date.After(endCommit.Date) {
			break
		}
		releaseCommits = append(releaseCommits, commit)
	}

	return releaseCommits, nil
}

func main() {
	version := os.Getenv("new_version")
	changelogPth := os.Getenv("changelog_pth")

	log.Infof("Configs:")
	log.Printf("new_version: %s", version)
	log.Printf("changelog_pth: %s", changelogPth)

	if version == "" {
		failf("Next version not defined")
	}

	if changelogPth == "" {
		failf("Changelog path not defined")
	}

	commits, err := releaseCommits(version)
	if err != nil {
		panic(err)
	}

	changelog, err := ChangelogContent(commits, version)
	if err != nil {
		panic(err)
	}

	if err := fileutil.WriteStringToFile(changelogPth, changelog); err != nil {

	}

	log.Infof("\nChangelog:")
	log.Printf(changelog)

	if err := tools.ExportEnvironmentWithEnvman("BITRISE_CHANGELOG", changelog); err != nil {
		failf("Failed to export changelog: %s", err)
	}

	log.Donef("\nThe changelog content is available in the BITRISE_CHANGELOG environment variable")
}
