package command

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestCompleteChannel(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	app.Commands = []cli.Command{
		{
			Name: "channel",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "apiKey, k"},
				cli.StringFlag{Name: "filter, f"},
				cli.StringFlag{Name: "outputFolder, o"},
				cli.StringFlag{Name: "xmlFile, x"},
				cli.StringFlag{Name: "baseURL, b"},
				cli.StringFlag{Name: "after, a"},
				cli.BoolFlag{Name: "cleanupUnrelatedFiles"},
			},
		},
	}
	os.Args = []string{os.Args[0], "channel", "--completion"}
	Completion(cli.NewContext(app, set, nil))
	assert.Equal(t, "--apiKey\n--filter\n--outputFolder\n--xmlFile\n--baseURL\n--after\n--cleanupUnrelatedFiles\n", writer.String())
}

func TestCompleteChannelApiKey(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	os.Args = []string{os.Args[0], "channel", "--apiKey", "--completion"}
	Completion(cli.NewContext(app, set, nil))
	assert.Equal(t, "", writer.String())
}

func TestCompleteChannelFilter(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	os.Args = []string{os.Args[0], "channel", "--filter", "--completion"}
	Completion(cli.NewContext(app, set, nil))
	assert.Equal(t, "", writer.String())
}

func TestCompleteChannelAfter(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	os.Args = []string{os.Args[0], "channel", "--after", "--completion"}
	Completion(cli.NewContext(app, set, nil))
	assert.Equal(t, "", writer.String())
}

func TestCompleteChannelOutputFolder(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	os.Args = []string{os.Args[0], "channel", "--outputFolder", "--completion"}
	Completion(cli.NewContext(app, set, nil))
	assert.Equal(t, "fileCompletion\n", writer.String())
}

func TestCompleteChannelXMLFile(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	app, writer, _ := appWithTestWriters()
	os.Args = []string{os.Args[0], "channel", "--xmlFile", "--completion"}
	Completion(cli.NewContext(app, set, nil))
	assert.Equal(t, "fileCompletion\n", writer.String())
}
