package git

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bitrise-io/go-utils/command"
	version "github.com/hashicorp/go-version"
)

const (
	hashPrefix    = "commit: "
	datePrefix    = "date: "
	authorPrefix  = "author: "
	messagePrefix = "message: "
)

func parseDate(unixTimeStampStr string) (time.Time, error) {
	i, err := strconv.ParseInt(unixTimeStampStr, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	if i < 0 {
		return time.Time{}, fmt.Errorf("invalid time stamp (%s)", unixTimeStampStr)
	}
	return time.Unix(i, 0), nil
}

func parseCommit(out string) (Commit, error) {
	// commit b738generate-release19a4d5
	// commit: b738generate-release19a4d5
	// date: 1455631980
	// author: Bitrise Bot
	// message: Merge branch 'master' of github.com:bitrise-bot/generate-changelog

	/*
		commit: 7d32generate-release43a6e9
		date: 1455788198
		author: Bitrise Developer
		message: FIX: parsing git commits
	*/
	hash := ""
	dateStr := ""
	author := ""
	message := ""
	messageStart := false
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, hashPrefix) {
			hash = strings.TrimPrefix(line, hashPrefix)
		} else if strings.HasPrefix(line, datePrefix) {
			dateStr = strings.TrimPrefix(line, datePrefix)
		} else if strings.HasPrefix(line, authorPrefix) {
			author = strings.TrimPrefix(line, authorPrefix)
		} else if strings.HasPrefix(line, messagePrefix) {
			messageStart = true
		}

		if messageStart {
			if strings.HasPrefix(line, messagePrefix) {
				message += strings.TrimPrefix(line, messagePrefix)
			} else {
				message += fmt.Sprintf("\n%s", line)
			}
		}
	}

	if hash == "" || dateStr == "" || author == "" {
		return Commit{}, fmt.Errorf("missing 'date: ' / 'author: ' / 'commit: ' fields in: %s", out)
	}

	date, err := parseDate(dateStr)
	if err != nil {
		return Commit{}, err
	}

	return Commit{
		Hash:    hash,
		Message: message,
		Date:    date,
		Author:  author,
	}, nil
}

// VersionTaggedCommits ...
func VersionTaggedCommits(repoDir string) ([]Commit, error) {
	cmd := command.New("git", "tag", "--list").SetDir(repoDir)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s failed: %s", cmd.PrintableCommandArgs(), out)
	}

	var taggedCommits []Commit
	for _, tag := range strings.Split(out, "\n") {
		tag = strings.TrimSpace(tag)

		// is tag sem-ver tag?
		if _, err := version.NewVersion(tag); err != nil {
			continue
		}

		cmd := command.New("git", "rev-list", "-n", "1", `--pretty=format:commit: %H%ndate: %ct%nauthor: %an%nmessage: %s`, tag).SetDir(repoDir)
		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("%s failed: %s", cmd.PrintableCommandArgs(), out)
		}

		commit, err := parseCommit(out)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse commit: %#v", err)
		}
		commit.Tag = tag

		taggedCommits = append(taggedCommits, commit)
	}

	sort.Slice(taggedCommits, func(i, j int) bool {
		return taggedCommits[i].Date.Before(taggedCommits[j].Date)
	})
	return taggedCommits, nil
}

// FirstCommit ...
func FirstCommit(repoDir string) (Commit, error) {
	cmd := command.New("git", "rev-list", "--max-parents=0", `--pretty=format:commit: %H%ndate: %ct%nauthor: %an%nmessage: %s`, "HEAD").SetDir(repoDir)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return Commit{}, fmt.Errorf("%s failed: %s", cmd.PrintableCommandArgs(), out)
	}
	return parseCommit(out)
}

// LastCommit ...
func LastCommit(repoDir string) (Commit, error) {
	cmd := command.New("git", "log", "-1", `--pretty=format:commit: %H%ndate: %ct%nauthor: %an%nmessage: %s`).SetDir(repoDir)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return Commit{}, fmt.Errorf("%s failed: %s", cmd.PrintableCommandArgs(), out)
	}
	return parseCommit(out)
}

func parseCommitList(out string) ([]Commit, error) {
	/*
		commit: 7e429002426846bb0837cb8b06603eb47a17d846
		date: 1523740143
		author: Krisztián Gödrei
		message: release wf (#2)
	*/
	var commits []Commit

	hash := ""
	dateStr := ""
	author := ""
	message := ""
	messageStart := false
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, hashPrefix) {
			if hash != "" {
				date, err := parseDate(dateStr)
				if err != nil {
					return nil, err
				}

				commits = append(commits, Commit{
					Hash:    hash,
					Message: message,
					Date:    date,
					Author:  author,
				})
				message = ""
				messageStart = false
			}
			hash = strings.TrimPrefix(line, hashPrefix)
		} else if strings.HasPrefix(line, datePrefix) {
			dateStr = strings.TrimPrefix(line, datePrefix)
		} else if strings.HasPrefix(line, authorPrefix) {
			author = strings.TrimPrefix(line, authorPrefix)
		} else if strings.HasPrefix(line, messagePrefix) {
			messageStart = true
		}

		if messageStart {
			if strings.HasPrefix(line, messagePrefix) {
				message += strings.TrimPrefix(line, messagePrefix)
			} else {
				message += fmt.Sprintf("\n%s", line)
			}
		}
	}
	date, err := parseDate(dateStr)
	if err != nil {
		return nil, err
	}

	commits = append(commits, Commit{
		Hash:    hash,
		Message: message,
		Date:    date,
		Author:  author,
	})

	return commits, nil
}

// Commits ...
func Commits(repoDir string) ([]Commit, error) {
	cmd := command.New("git", "log", `--pretty=format:commit: %H%ndate: %ct%nauthor: %an%nmessage: %s`).SetDir(repoDir)
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s failed: %s", cmd.PrintableCommandArgs(), out)
	}

	commits, err := parseCommitList(out)
	if err != nil {
		return []Commit{}, err
	}

	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Date.Before(commits[j].Date)
	})

	return commits, nil
}
