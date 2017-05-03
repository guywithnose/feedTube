package command_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/guywithnose/feedTube/command"
	"github.com/stretchr/testify/assert"
)

func TestCleanupUnrelatedFilesRemovesUnrelatedFiles(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	defer removeFile(t, outputFolder)
	unrelatedFile := fmt.Sprintf("%s/t-vId1.mp3", outputFolder)
	_, err := os.Create(unrelatedFile)
	assert.Nil(t, err)
	writer := new(bytes.Buffer)
	assert.Nil(t, command.NewDirectoryCleaner(outputFolder).CleanupUnrelatedFiles([]string{}, writer))
	assert.Equal(t, "Removing file: /tmp/testFeedTube/t-vId1.mp3\n", writer.String())
}

func TestCleanupUnrelatedFilesDoesntRemoveRelatedFiles(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	defer removeFile(t, outputFolder)
	relatedFile := fmt.Sprintf("%s/t-vId1.mp3", outputFolder)
	_, err := os.Create(relatedFile)
	assert.Nil(t, err)
	writer := new(bytes.Buffer)
	assert.Nil(t, command.NewDirectoryCleaner(outputFolder).CleanupUnrelatedFiles([]string{relatedFile}, writer))
	assert.Equal(t, "", writer.String())
}

func TestCleanupUnrelatedFilesDoesntRemoveDirectories(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	unrelatedDirectory := fmt.Sprintf("%s/dir", outputFolder)
	assert.Nil(t, os.MkdirAll(unrelatedDirectory, 0777))
	defer removeFile(t, outputFolder)
	unrelatedFile := fmt.Sprintf("%s/t-vId1.mp3", unrelatedDirectory)
	_, err := os.Create(unrelatedFile)
	assert.Nil(t, err)
	writer := new(bytes.Buffer)
	assert.EqualError(
		t,
		command.NewDirectoryCleaner(outputFolder).CleanupUnrelatedFiles([]string{}, writer),
		"could not remove unrelated file: remove /tmp/testFeedTube/dir: directory not empty",
	)
	assert.Equal(t, "Removing file: /tmp/testFeedTube/dir\n", writer.String())
}
