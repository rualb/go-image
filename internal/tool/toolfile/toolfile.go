// Package toolfile ...
package toolfile

import (
	"os"
	"path/filepath"
)

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		// Directory does not exist
		return false
	}
	return info != nil && info.IsDir() // Return true if it's a directory
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		// Directory does not exist
		return false
	}
	return info != nil && !info.IsDir() // Return true if it's NOT a directory
}

func FileSize(path string) int64 {
	info, err := os.Stat(path)
	if err == nil && info != nil && !info.IsDir() {
		return info.Size()
	}
	return 0
}

func MakeAllDirs(path string) error {
	// os.MkdirAll creates a directory along with any necessary parents.
	// 0755 is the permission mode (owner can read, write, and execute; others can read and execute).
	return os.MkdirAll(path, 0750)
}

func FileWrite(path string, data []byte) error {
	return os.WriteFile(path, data, 0600)
}

func FileWriteWithDir(path string, data []byte) error {

	err := MakeAllDirs(filepath.Dir(path))
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
