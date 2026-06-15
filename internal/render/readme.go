package render

import (
	"path/filepath"
	"strings"

	"github.com/thomiceli/opengist/internal/git"
)

func IsReadmeFilename(filename string) bool {
	return strings.EqualFold(filepath.Base(filename), "README.md")
}

func IsMarkdownFilename(filename string) bool {
	return strings.EqualFold(filepath.Ext(filename), ".md")
}

func FindReadmeFile(files []*git.File) *git.File {
	for _, file := range files {
		if IsReadmeFilename(file.Filename) {
			return file
		}
	}
	return nil
}

func HasReadmeInFilenames(filenames []string) bool {
	for _, name := range filenames {
		if IsReadmeFilename(name) {
			return true
		}
	}
	return false
}

func HasPostTab(filenames []string) bool {
	return len(filenames) > 0
}

func findRenderedFileByName(files map[string]RenderedFile, name string) (RenderedFile, bool) {
	if rendered, ok := files[name]; ok {
		return rendered, true
	}
	lower := strings.ToLower(name)
	for filename, rendered := range files {
		if strings.ToLower(filename) == lower {
			return rendered, true
		}
	}
	return nil, false
}
