package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitrise-io/envman/envman"
	"github.com/bitrise-steplib/steps-generate-changelog/exporter"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockExporter struct {
	mock.Mock
}

// EnvKey ...
func (e mockExporter) EnvKey() string { //nolint
	args := e.Called()
	return args.String(0)
}

// Filepath ...
func (e mockExporter) Filepath() string { //nolint
	args := e.Called()
	return args.String(0)
}

// WriteFile ...
func (e mockExporter) WriteFile(content string) error { //nolint
	args := e.Called(content)
	return args.Error(0)
}

// ExportEnv ...
func (e mockExporter) ExportEnv(value string) error { //nolint
	args := e.Called(value)
	return args.Error(0)
}

// MaxEnvBytes ...
func (e mockExporter) MaxEnvBytes() (int, error) { //nolint
	args := e.Called()
	return args.Int(0), args.Error(1)
}

func Test_exportChangelog(t *testing.T) {
	envmanConfigs, err := envman.GetConfigs()
	require.NoError(t, err)

	t.Run("ok - under limit", func(t *testing.T) {
		mockExporter := mockExporter{}
		mockContent := "content"

		mockExporter.On("WriteFile", mockContent).Return(nil).Once()
		mockExporter.On("MaxEnvBytes").Return(0, nil).Once()
		mockExporter.On("ExportEnv", mockContent).Return(nil).Once()

		require.NoError(t, exportChangelog(mockContent, mockExporter)) //nolint

		mockExporter.AssertExpectations(t)
	})
	t.Run("ok - above limit", func(t *testing.T) {
		mockExporter := mockExporter{}
		mockContent := strings.Repeat("a", (envmanConfigs.EnvBytesLimitInKB+1)*1024)

		mockExporter.On("WriteFile", mockContent).Return(nil).Once()
		mockExporter.On("Filepath").Return("").Once()
		mockExporter.On("MaxEnvBytes").Return(envmanConfigs.EnvBytesLimitInKB*1024, nil).Once()
		mockExporter.On("ExportEnv", mock.MatchedBy( // check if input argument "content" was stripped
			func(content string) bool {
				return len(content) == envmanConfigs.EnvBytesLimitInKB*1024 &&
					content[len(content)-4:] == "\n..."
			})).Return(nil).Once()

		require.NoError(t, exportChangelog(mockContent, mockExporter)) //nolint

		mockExporter.AssertExpectations(t)
	})

	t.Run("error - unable to write file", func(t *testing.T) {
		mockExporter := mockExporter{}
		mockContent := strings.Repeat("a", (envmanConfigs.EnvBytesLimitInKB+1)*1024)

		mockExporter.On("Filepath").Return("").Once()
		mockExporter.On("WriteFile", mockContent).Return(errors.New("error")).Once()

		require.Error(t, exportChangelog(mockContent, mockExporter)) //nolint

		mockExporter.AssertExpectations(t)
	})

	t.Run("error - unable to get envman config", func(t *testing.T) {
		mockExporter := mockExporter{}
		mockContent := "content"

		mockExporter.On("WriteFile", mockContent).Return(nil).Once()
		mockExporter.On("MaxEnvBytes").Return(0, errors.New("")).Once()

		require.Error(t, exportChangelog(mockContent, mockExporter)) //nolint

		mockExporter.AssertExpectations(t)
	})

	t.Run("error - unable to export env", func(t *testing.T) {
		mockExporter := mockExporter{}
		mockContent := "content"

		mockExporter.On("WriteFile", mockContent).Return(nil).Once()
		mockExporter.On("EnvKey").Return("").Once()
		mockExporter.On("MaxEnvBytes").Return(0, nil).Once()
		mockExporter.On("ExportEnv", mockContent).Return(errors.New("")).Once()

		require.Error(t, exportChangelog(mockContent, mockExporter)) //nolint

		mockExporter.AssertExpectations(t)
	})
}

func TestContentEscaping(t *testing.T) {
	content := `This is a changelog with some env vars:
- $BITRISE_BUILD_NUMBER
- $HOME
- $PWD
- $SHELL
`
	contentEnvKey := "TEST_CHANGELOG_CONTENT"
	contentPath := filepath.Join(t.TempDir(), "changelog.md")
	exporter := exporter.New(contentEnvKey, contentPath)
	
	err := exportChangelog(content, exporter)
	require.NoError(t, err)

	b, err := os.ReadFile(contentPath)
	require.NoError(t, err)
	require.Equal(t, content, string(b))
}
