package pkg

import (
	"path/filepath"
	"strings"
)

// ./a/b/c/ => /a/b/c.
func DirPathToImportPath(dirPath string) string {
	if dirPath == "" {
		return ""
	}

	var path string
	if string([]rune(dirPath)[0]) == "." {
		path = string([]rune(dirPath)[1:])
	} else {
		path = dirPath
	}

	path = strings.Trim(path, "/")
	path = string(filepath.Separator) + path
	//用于解析
	path = strings.Replace(path, string(filepath.Separator), "/", -1)
	return path
}
