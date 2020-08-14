package exporter

import (
	"github.com/bitrise-io/envman/envman"
	"github.com/bitrise-io/go-steputils/tools"
	"github.com/bitrise-io/go-utils/fileutil"
)

// EnvAndFile ...
type EnvAndFile struct {
	envKey, filepath string
}

// New ...
func New(envKey, filepath string) EnvAndFile {
	return EnvAndFile{envKey: envKey, filepath: filepath}
}

// EnvKey ...
func (e EnvAndFile) EnvKey() string { return e.envKey }

// Filepath ...
func (e EnvAndFile) Filepath() string { return e.filepath }

// WriteFile ...
func (e EnvAndFile) WriteFile(content string) error {
	return fileutil.WriteStringToFile(e.Filepath(), content)
}

// ExportEnv ...
func (e EnvAndFile) ExportEnv(value string) error {
	return tools.ExportEnvironmentWithEnvman(e.EnvKey(), value)
}

// MaxEnvBytes ...
func (e EnvAndFile) MaxEnvBytes() (int, error) {
	envmanConfigs, err := envman.GetConfigs()
	if err != nil {
		return 0, err
	}
	return envmanConfigs.EnvBytesLimitInKB * 1024, nil
}
