package command

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DirectoryCleaner cleans up unrelated files
type DirectoryCleaner struct {
	outputFolder string
}

// NewDirectoryCleaner returns a new directoryCleaner
func NewDirectoryCleaner(outputFolder string) *DirectoryCleaner {
	return &DirectoryCleaner{
		outputFolder: outputFolder,
	}
}

// CleanupUnrelatedFiles searches the outputFolder for files that are not in relatedFiles and deletes them
func (cleaner DirectoryCleaner) CleanupUnrelatedFiles(relatedFiles []string, writer io.Writer) error {
	unrelatedFiles := cleaner.findUnrelatedFiles(relatedFiles)

	for _, unrelatedFile := range unrelatedFiles {
		fmt.Fprintf(writer, "Removing file: %s\n", unrelatedFile)
		err := os.Remove(unrelatedFile)
		if err != nil {
			return fmt.Errorf("could not remove unrelated file: %v", err)
		}
	}

	return nil
}

func (cleaner DirectoryCleaner) findUnrelatedFiles(relatedFiles []string) []string {
	dir, _ := os.Open(cleaner.outputFolder)
	files, _ := dir.Readdir(-1)

	unrelatedFiles := make([]string, 0, len(files))
	for _, file := range files {
		absoluteFileName, err := filepath.Abs(fmt.Sprintf("%s/%s", cleaner.outputFolder, file.Name()))
		if err == nil && !ContainsString(absoluteFileName, relatedFiles) {
			unrelatedFiles = append(unrelatedFiles, absoluteFileName)
		}
	}

	return unrelatedFiles
}
