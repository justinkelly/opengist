package render_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomiceli/opengist/internal/git"
	"github.com/thomiceli/opengist/internal/render"
)

func TestIsReadmeFilename(t *testing.T) {
	assert.True(t, render.IsReadmeFilename("README.md"))
	assert.True(t, render.IsReadmeFilename("readme.MD"))
	assert.False(t, render.IsReadmeFilename("notes.md"))
}

func TestHasPostTab(t *testing.T) {
	assert.False(t, render.HasPostTab(nil))
	assert.False(t, render.HasPostTab([]string{}))
	assert.True(t, render.HasPostTab([]string{"main.go"}))
	assert.True(t, render.HasPostTab([]string{"README.md"}))
}

func testRenderFile(t *testing.T) func(render.RenderedFile) string {
	return func(file render.RenderedFile) string {
		switch f := file.(type) {
		case render.HighlightedFile:
			return `<div data-file="` + f.Filename + `">` + strings.Join(f.Lines, "") + `</div>`
		case render.NonHighlightedFile:
			return `<div data-file="` + f.Filename + `"></div>`
		default:
			return `<div data-file=""></div>`
		}
	}
}

func TestPostMarkdownEmbedsFileShortcode(t *testing.T) {
	codeFile := &git.File{
		Filename:  "example.go",
		Content:   "package main",
		HumanSize: "12 B",
		MimeType:  git.MimeType{ContentType: "text/plain"},
	}
	rendered := render.HighlightedFile{
		File:  codeFile,
		Type:  "Go",
		Lines: []string{"package main"},
	}

	ctx := render.PostContext{
		Rendered: map[string]render.RenderedFile{
			"example.go": rendered,
		},
		RenderFile: testRenderFile(t),
	}

	html, err := render.PostMarkdown("# Title\n\n{{ \"example.go\" }}\n", ctx)
	require.NoError(t, err)
	assert.Contains(t, html, "<h1>Title</h1>")
	assert.Contains(t, html, "data-file=\"example.go\"")
	assert.Contains(t, html, "package main")

	html, err = render.PostMarkdown("# Title\n\n{{ example.go }}\n", ctx)
	require.NoError(t, err)
	assert.Contains(t, html, "data-file=\"example.go\"")
}

func TestPostMarkdownEmbedsMultipleFiles(t *testing.T) {
	ctx := render.PostContext{
		Rendered: map[string]render.RenderedFile{
			"test.sql": render.HighlightedFile{
				File:  &git.File{Filename: "test.sql"},
				Lines: []string{"SELECT 1;"},
			},
			"3csv.png": render.NonHighlightedFile{File: &git.File{Filename: "3csv.png"}, Type: "PNG"},
		},
		RenderFile: testRenderFile(t),
	}

	html, err := render.PostMarkdown("{{ test.sql }}\n\n{{ \"3csv.png\" }}\n", ctx)
	require.NoError(t, err)
	assert.Contains(t, html, "data-file=\"test.sql\"")
	assert.Contains(t, html, "SELECT 1;")
	assert.Contains(t, html, "3csv.png")
}

func TestPostMarkdownLeavesFencedCodeAlone(t *testing.T) {
	html, err := render.PostMarkdown("```python\nprint('hi')\n```\n", render.PostContext{
		Rendered:   map[string]render.RenderedFile{},
		RenderFile: func(render.RenderedFile) string { return "" },
	})
	require.NoError(t, err)
	assert.NotContains(t, html, "data-file=")
	assert.Contains(t, html, "print")
}

func TestPostMarkdownMissingEmbedFile(t *testing.T) {
	html, err := render.PostMarkdown(`{{ "missing.go" }}`, render.PostContext{
		Rendered:   map[string]render.RenderedFile{},
		RenderFile: func(render.RenderedFile) string { return "" },
	})
	require.NoError(t, err)
	assert.True(t, strings.Contains(html, "File not found"))
}
