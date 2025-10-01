package filescanner

import "io/fs"

func CategorizeFile(path string, entry fs.DirEntry) (issueType string, shouldFlag bool) {
	return "", false
}
