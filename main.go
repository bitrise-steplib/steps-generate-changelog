package main

import (
	"os"

	"github.com/bitrise-io/envman/envman"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/steps-generate-changelog/git"
	"github.com/bitrise-tools/go-steputils/stepconf"
	"github.com/bitrise-tools/go-steputils/tools"
	"github.com/pkg/errors"
)

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
	if len(taggedCommits) > 1 {
		// collecting changelog between existing versions
		startCommit = taggedCommits[len(taggedCommits)-2]
		endCommit = taggedCommits[len(taggedCommits)-1]
		includeFirst = false
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

func processChangelog(cp changelogPrinter, c Config) {
	changelog, err := cp.content()
	if err != nil {
		failf("Failed to get changelog content, error: %s", err)
	}

	if err := fileutil.WriteStringToFile(c.ChangelogPath, changelog); err != nil {
		failf("Failed to write changelog to (%s), error: %s", c.ChangelogPath, err)
	}

	log.Infof("\nChangelog:")
	log.Printf(changelog)

	envmanConfigs, err := envman.GetConfigs()
	if err != nil {
		failf("Failed to load envman configs, error: %s", err)
	}

	if envmanConfigs.EnvBytesLimitInKB > 0 {
		envBytesLimit := envmanConfigs.EnvBytesLimitInKB * 1024

		if len(changelog) > envBytesLimit {
			log.Warnf("Changelog content exceeds the maximum allowed size to set in an environment variable. (%dKB)", envmanConfigs.EnvBytesLimitInKB)
			log.Warnf("The changelog's content will be trimmed to fit the maximum allowed size.")
			log.Warnf("It is possible to modify the limit by following this guide: https://devcenter.bitrise.io/tips-and-tricks/increasing-the-size-limit-of-env-vars")
			changelog = changelog[:envBytesLimit-4] + "\n..."
		}
	}

	if err := tools.ExportEnvironmentWithEnvman("BITRISE_CHANGELOG", changelog); err != nil {
		failf("Failed to export changelog to (BITRISE_CHANGELOG), error: %s", err)
	}

	log.Donef("\nThe changelog content is available in the BITRISE_CHANGELOG environment variable")
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

	processChangelog(&changelog{Commits: commits, Version: c.NewVersion}, c)
}
