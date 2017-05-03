package command

import (
	"fmt"
	"os"
	"testing"

	"github.com/guywithnose/commandBuilder"
	"github.com/stretchr/testify/assert"
)

func TestDownloader(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	videos := []*VideoData{getVideoData("vId1", "t"), getVideoData("vId2", "t2")}
	cmdBuilder := getTestCommandBuilder(videos)
	downloader := NewDownloader(cmdBuilder, outputFolder)
	assert.Nil(t, downloader.DownloadVideos(videos))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cmdBuilder.ExpectedCommands)
	assert.Equal(t, []error(nil), cmdBuilder.Errors)
}

func TestDownloaderDoesntReDownloadExisitngFiles(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	defer removeFile(t, outputFolder)
	fileToSkip := fmt.Sprintf("%s/t-vId1.mp3", outputFolder)
	_, err := os.Create(fileToSkip)
	assert.Nil(t, err)
	videos := []*VideoData{getVideoData("vId1", "t"), getVideoData("vId2", "t2")}
	cmdBuilder := getTestCommandBuilder(videos[1:])
	downloader := NewDownloader(cmdBuilder, outputFolder)
	assert.Nil(t, downloader.DownloadVideos(videos))
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cmdBuilder.ExpectedCommands)
	assert.Equal(t, []error(nil), cmdBuilder.Errors)
}

func TestDownloaderDownloadError(t *testing.T) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	videos := []*VideoData{getVideoData("vId1", "t")}
	cmdBuilder := getTestErrorCommandBuilder(videos)
	downloader := NewDownloader(cmdBuilder, outputFolder)
	assert.EqualError(
		t,
		downloader.DownloadVideos(videos),
		"could not download t-vId1: exit status 1\nParams: '/usr/bin/youtube-dl' '-x' '--audio-format' 'mp3' '--audio-quality' '0' '-o' "+
			"'/tmp/testFeedTube/t-vId1.%(ext)s' 'https://youtu.be/vId1': error downloading video vId1",
	)
	assert.Equal(t, []*commandBuilder.ExpectedCommand{}, cmdBuilder.ExpectedCommands)
	assert.Equal(t, []error(nil), cmdBuilder.Errors)
}

func getVideoData(id, title string) *VideoData {
	return &VideoData{
		GUID:     id,
		Title:    title,
		FileName: fmt.Sprintf("%s-%s", title, id),
	}
}

func getTestCommandBuilder(videos []*VideoData) *commandBuilder.Test {
	expectedCommands := make([]*commandBuilder.ExpectedCommand, 0, len(videos))
	for _, video := range videos {
		expectedCommands = append(expectedCommands, getExpectedCommandForVideoData(video))
	}

	return &commandBuilder.Test{ExpectedCommands: expectedCommands}
}

func getTestErrorCommandBuilder(videos []*VideoData) *commandBuilder.Test {
	expectedCommands := make([]*commandBuilder.ExpectedCommand, 0, len(videos))
	for _, video := range videos {
		expectedCommands = append(expectedCommands, getExpectedCommandWithErrorForVideoData(video))
	}

	return &commandBuilder.Test{ExpectedCommands: expectedCommands}
}

func getExpectedCommandForVideoData(video *VideoData) *commandBuilder.ExpectedCommand {
	return commandBuilder.NewExpectedCommand(
		"",
		getDownloadCommand(video),
		fmt.Sprintf("output for video %s", video.GUID),
		0,
	)
}

func getExpectedCommandWithErrorForVideoData(video *VideoData) *commandBuilder.ExpectedCommand {
	return commandBuilder.NewExpectedCommand(
		"",
		getDownloadCommand(video),
		fmt.Sprintf("error downloading video %s", video.GUID),
		1,
	)
}

func getDownloadCommand(video *VideoData) string {
	return fmt.Sprintf(
		"/usr/bin/youtube-dl -x --audio-format mp3 --audio-quality 0 -o /tmp/testFeedTube/%s-%s.%%(ext)s https://youtu.be/%s",
		video.Title,
		video.GUID,
		video.GUID,
	)
}

func TestHelperProcess(*testing.T) {
	commandBuilder.ErrorCodeHelper()
}

func removeFile(t *testing.T, fileName string) {
	assert.Nil(t, os.RemoveAll(fileName))
}
