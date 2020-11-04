package main

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/steps-generate-changelog/exporter"
	"github.com/bitrise-steplib/steps-generate-changelog/git"
	"github.com/bitrise-tools/go-steputils/stepconf"
	"github.com/pkg/errors"
)

const changelogContentEnvKey = "BITRISE_CHANGELOG"

func failf(format string, args ...interface{}) {
	log.Errorf(format, args...)
	os.Exit(1)
}

func releaseCommits(dir, newVersion string) ([]git.Commit, error) {
	startCommit, err := git.FirstCommit(dir)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	endCommit, err := git.LastCommit(dir)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	taggedCommits, err := git.TaggedCommits(dir)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	includeFirst := true
	if len(taggedCommits) > 0 {
		// there is at least one version
		if endCommit.Hash != taggedCommits[len(taggedCommits)-1].Hash {
			// last commit is not the same as the last tag
			// collecting changelog since last version
			startCommit = taggedCommits[len(taggedCommits)-1]
			includeFirst = false
		} else if len(taggedCommits) > 1 {
			// last commit has a tag and there are at least two versions
			// collecting changelog between last two versions
			startCommit = taggedCommits[len(taggedCommits)-2]
			includeFirst = false
		}
		// otherwise there is only one tag and is on the last commit
		// nothing to do here, will collect all commits
	}

	commits, err := git.Commits(dir)
	if err != nil {
		return nil, errors.WithStack(err)
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

// Config ...
type Config struct {
	NewVersion    string `env:"new_version,required"`
	ChangelogPath string `env:"changelog_pth,required"`
	WorkDir       string `env:"working_dir,required"`
}

type outputExporter interface {
	EnvKey() string
	Filepath() string
	WriteFile(content string) error
	ExportEnv(value string) error
	MaxEnvBytes() (int, error)
}

func exportChangelog(changelog string, e outputExporter) error {
	if err := e.WriteFile(changelog); err != nil {
		return fmt.Errorf("unable to write changelog to (%s), error: %s", e.Filepath(), err)
	}

	maxEnvBytes, err := e.MaxEnvBytes()
	if err != nil {
		return fmt.Errorf("unable to load envman configs, error: %s", err)
	}

	if maxEnvBytes > 0 {
		if len(changelog) > maxEnvBytes {
			log.Warnf("Changelog content exceeds the maximum allowed size to set in an environment variable. (%dKB)", maxEnvBytes/1024)
			log.Warnf("The changelog's content will be trimmed to fit the maximum allowed size.")
			log.Warnf("It is possible to modify the limit by following this guide: https://devcenter.bitrise.io/tips-and-tricks/increasing-the-size-limit-of-env-vars")
			log.Warnf("or you can use the exported changelog file(%s) also.", e.Filepath())
			changelog = changelog[:maxEnvBytes-4] + "\n..."
		}
	}

	if err := e.ExportEnv(changelog); err != nil {
		return fmt.Errorf("unable to export environment variable with envman (%s), error: %s", e.EnvKey(), err)
	}

	return nil
}

func main() {
	var c Config
	if err := stepconf.Parse(&c); err != nil {
		failf("Failed to parse configs, error: %s", err)
	}
	stepconf.Print(c)

	commits, err := releaseCommits(c.WorkDir, c.NewVersion)
	if err != nil {
		failf("Failed to get release commits, error: %v", err)
	}

	content, err := changelogContent(commits, c.NewVersion)
	if err != nil {
		failf("Failed to get changelog content, error: %s", err)
	}

	log.Infof("\nChangelog:")
	log.Printf(content)

	if err := exportChangelog(content, exporter.New(changelogContentEnvKey, c.ChangelogPath)); err != nil {
		failf("Failed to export changelog: %s", err)
	}

	log.Donef("\nThe changelog content is available in the " + changelogContentEnvKey + " environment variable")
}
