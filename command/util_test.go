package command

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/guywithnose/commandBuilder"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func removeFile(t *testing.T, fileName string) {
	assert.Nil(t, os.RemoveAll(fileName))
}

func TestHelperProcess(*testing.T) {
	commandBuilder.ErrorCodeHelper()
}

func getTestServer(responses map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response, ok := responses[r.URL.String()]
		if !ok {
			panic(r.URL.String())
		}

		if response == "error" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, response)
	}))
}

func appWithTestWriters() (*cli.App, *bytes.Buffer, *bytes.Buffer) {
	app := cli.NewApp()
	writer := new(bytes.Buffer)
	errWriter := new(bytes.Buffer)
	app.Writer = writer
	app.ErrWriter = errWriter
	return app, writer, errWriter
}

func getBaseAppAndFlagSet(t *testing.T, outputFolder string) (*cli.App, *bytes.Buffer, *bytes.Buffer, *flag.FlagSet) {
	set := flag.NewFlagSet("test", 0)
	set.String("apiKey", "fakeApiKey", "doc")
	set.String("outputFolder", outputFolder, "doc")
	xmlFile := fmt.Sprintf("%s/xmlFile", outputFolder)
	set.String("xmlFile", xmlFile, "doc")
	set.String("baseURL", "http://foo.com", "doc")
	assert.Nil(t, set.Parse([]string{"awesome"}))
	app, writer, errorWriter := appWithTestWriters()
	return app, writer, errorWriter, set
}

func runErrorTest(
	t *testing.T,
	expectedError string,
	cb *commandBuilder.Test,
	cmdFunc func(commandBuilder.Builder) func(*cli.Context) error,
) {
	outputFolder := fmt.Sprintf("%s/testFeedTube", os.TempDir())
	defer removeFile(t, outputFolder)
	assert.Nil(t, os.MkdirAll(outputFolder, 0777))
	app, _, errWriter, set := getBaseAppAndFlagSet(t, outputFolder)
	assert.Nil(t, cmdFunc(cb)(cli.NewContext(app, set, nil)))
	assert.Equal(t, []error(nil), cb.Errors)
	assert.Equal(t, expectedError, errWriter.String())
}
