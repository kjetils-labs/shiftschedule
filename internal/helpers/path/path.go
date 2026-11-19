// Package path provides a utility function to find a file in the current directory or any of its parent directories.
package path

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func FindFile(path string) (string, error) {
	for range 5 {

		_, err := os.Stat(path)
		if err == nil {
			var abspath string
			abspath, err = filepath.Abs(path)
			return abspath, err
		}

		if errors.Is(err, fs.ErrNotExist) {
			path = "../" + path
			continue
		}

	}
	return "", fmt.Errorf("could not find file %v", path)
}
