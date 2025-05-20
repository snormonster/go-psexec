package tar_io

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/sirupsen/logrus"
)

type directoryTarProvider struct {
	fullDirPath string
	filePattern string
}

func (d *directoryTarProvider) Files() <-chan *TarFile {
    filesChan := make(chan *TarFile)

    go func() {
        defer close(filesChan)

        // Walk the directory (could swap to afero.Walk later)
        if err := filepath.Walk(d.fullDirPath, func(path string, info os.FileInfo, errParam error) error {
            if errParam != nil {
                return errParam
            }

            // Filter by pattern if provided
            if d.filePattern != "" {
                match, matchErr := filepath.Match(d.filePattern, info.Name())
                if matchErr != nil {
                    return fmt.Errorf("file pattern match error for '%s': %w", d.filePattern, matchErr)
                }
                if !match {
                    return nil
                }
            }

            // Compute relative path
            relPath := strings.TrimPrefix(path, d.fullDirPath)
            if relPath == "" {
                return nil
            }
            relPath = strings.TrimPrefix(relPath, string(os.PathSeparator))

            // Open file if not a directory
            var content io.ReadCloser
            if !info.IsDir() {
                f, openErr := os.Open(path)
                if openErr != nil {
                    return fmt.Errorf("unable to open file '%s': %w", path, openErr)
                }
                content = f
            }

            // Send the TarFile down the channel
            filesChan <- NewTarFile(relPath, content, false, info)
            return nil
        }); err != nil {
            // Log the error; nothing left to do since channel is closing
            logrus.Errorf("unable to walk dir %q: %v", d.fullDirPath, err)
        }
    }()

    return filesChan
}


type emptyReader struct{}

func (e *emptyReader) Read(p []byte) (int, error) {
	return 0, nil
}
