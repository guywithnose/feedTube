package command_test

import (
	"flag"
	"os"
	"testing"

	"github.com/guywithnose/feedTube/command"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestCompleteChannel(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	app.Commands = command.Commands
	os.Args = []string{os.Args[0], "channel", "--completion"}
	command.Completion(cli.NewContext(app, set, nil))
	assert.Equal(t, "--apiKey\n--filter\n--outputFolder\n--xmlFile\n--baseURL\n--cleanupUnrelatedFiles\n--overrideTitle\n--quality\n--after\n", writer.String())
}

func TestCompleteChannelApiKey(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	os.Args = []string{os.Args[0], "channel", "--apiKey", "--completion"}
	command.Completion(cli.NewContext(app, set, nil))
	assert.Equal(t, "", writer.String())
}

func TestCompleteChannelFilter(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	os.Args = []string{os.Args[0], "channel", "--filter", "--completion"}
	command.Completion(cli.NewContext(app, set, nil))
	assert.Equal(t, "", writer.String())
}

func TestCompleteChannelAfter(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	os.Args = []string{os.Args[0], "channel", "--after", "--completion"}
	command.Completion(cli.NewContext(app, set, nil))
	assert.Equal(t, "", writer.String())
}

func TestCompleteChannelOutputFolder(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	os.Args = []string{os.Args[0], "channel", "--outputFolder", "--completion"}
	command.Completion(cli.NewContext(app, set, nil))
	assert.Equal(t, "fileCompletion\n", writer.String())
}

func TestCompleteChannelXMLFile(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	os.Args = []string{os.Args[0], "channel", "--xmlFile", "--completion"}
	command.Completion(cli.NewContext(app, set, nil))
	assert.Equal(t, "fileCompletion\n", writer.String())
}

func TestRootCompletion(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	app.Commands = append(command.Commands, cli.Command{Hidden: true, Name: "don't show"})
	command.RootCompletion(cli.NewContext(app, set, nil))
	assert.Equal(t, "channel:Builds your rss file from a youtube channel\nplaylist:Builds your rss file from a youtube playlist\n", writer.String())
}
