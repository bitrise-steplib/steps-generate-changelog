package exporter

import (
	"github.com/bitrise-io/envman/envman"
	"github.com/bitrise-io/go-steputils/v2/export"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/v2/command"
	"github.com/bitrise-io/go-utils/v2/env"
)

type EnvAndFile struct {
	envKey, filepath string
	exporter export.Exporter
}

func New(envKey, filepath string) EnvAndFile {
	exporter := export.NewExporter(command.NewFactory(env.NewRepository()))
	return EnvAndFile{envKey: envKey, filepath: filepath, exporter: exporter}
}

func (e EnvAndFile) EnvKey() string { return e.envKey }

func (e EnvAndFile) Filepath() string { return e.filepath }

func (e EnvAndFile) WriteFile(content string) error {
	return fileutil.WriteStringToFile(e.Filepath(), content)
}

func (e EnvAndFile) ExportEnv(value string) error {
	// Do not expand env vars in the generated changelog because it the input is beyond the control of the step,
	// and it could lead to surprising behavior.
	return e.exporter.ExportOutputNoExpand(e.EnvKey(), value)
}

func (e EnvAndFile) MaxEnvBytes() (int, error) {
	envmanConfigs, err := envman.GetConfigs()
	if err != nil {
		return 0, err
	}
	return envmanConfigs.EnvBytesLimitInKB * 1024, nil
}
