package tar_io

import (
	"os"
	
	"github.com/sirupsen/logrus"
)

type fileTarProvider struct {
	fullFilePath string
}

func (f *fileTarProvider) Files() <-chan *TarFile {
    filesChan := make(chan *TarFile)

    go func() {
        defer close(filesChan)

        // Get file info
        info, err := os.Stat(f.fullFilePath)
        if err != nil {
            logrus.Errorf("unable to stat file %q: %v", f.fullFilePath, err)
            return
        }

        // Open file for reading
        rc, err := os.Open(f.fullFilePath)
        if err != nil {
            logrus.Errorf("unable to open file %q: %v", f.fullFilePath, err)
            return
        }

        // Send the single TarFile down the channel
        filesChan <- NewTarFile(f.fullFilePath, rc, true, info)
    }()

    return filesChan
}
