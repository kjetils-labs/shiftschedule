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
			abspath, err := filepath.Abs(path)
			return abspath, err
		}

		if errors.Is(err, fs.ErrNotExist) {
			path = "../" + path
			continue
		}

	}
	return "", fmt.Errorf("could not find file %v", path)
}
