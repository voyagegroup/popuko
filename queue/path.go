package queue

import (
	"errors"
	"path"
	"path/filepath"
	"strings"
)

// Based on https://github.com/golang/go/blob/98842cabb6133ab7b8f2b323754a48085eed82f3/src/net/http/fs.go#L26-L50
// This code's license is https://github.com/golang/go/blob/98842cabb6133ab7b8f2b323754a48085eed82f3/LICENSE.
func createAbs(root, name string) (p string, err error) {
	if root == "" || name == "" {
		return "", errors.New("all arguments must not be empty")
	}

	if filepath.Separator != '/' && strings.ContainsRune(name, filepath.Separator) ||
		strings.Contains(name, "\x00") {
		return "", errors.New("invalid character in file path")
	}

	f := filepath.Join(filepath.FromSlash(path.Clean("/"+root)), filepath.FromSlash(path.Clean("/"+name)))
	if f == "/" {
		return "", errors.New("the result is `/`")
	}

	return f, nil
}
