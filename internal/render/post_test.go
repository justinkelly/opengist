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

func TestPostMarkdownLeavesImageSyntaxInFencedCodeAlone(t *testing.T) {
	ctx := imagePostContext(t)

	html, err := render.PostMarkdown("```markdown\n![screenshot](item-csv.png){align=left}\n```\n\n![live](item-csv.png)\n", ctx)
	require.NoError(t, err)
	assert.Contains(t, html, "<pre")
	assert.Contains(t, html, "align=left")
	assert.NotContains(t, html, "float:left")
	assert.Equal(t, 1, strings.Count(html, "<img "))
}

func TestPostMarkdownLeavesImageSyntaxInInlineCodeAlone(t *testing.T) {
	ctx := imagePostContext(t)

	html, err := render.PostMarkdown("Use `![screenshot](item-csv.png){align=left}` in docs.\n\n![live](item-csv.png)\n", ctx)
	require.NoError(t, err)
	assert.Contains(t, html, "<code>![screenshot](item-csv.png){align=left}</code>")
	assert.NotContains(t, html, "float:left")
	assert.Contains(t, html, `src="https://example.com/user/gist/raw/HEAD/item-csv.png"`)
}

func TestPostMarkdownLeavesEmbedSyntaxInFencedCodeAlone(t *testing.T) {
	ctx := embedPostContext(t)

	html, err := render.PostMarkdown("```go\n{{ \"example.go\" }}\n```\n\n{{ \"example.go\" }}\n", ctx)
	require.NoError(t, err)
	assert.Contains(t, html, "<pre")
	assert.Contains(t, html, "{{")
	assert.Equal(t, 1, strings.Count(html, `data-file="example.go"`))
}

func TestPostMarkdownLeavesEmbedSyntaxInInlineCodeAlone(t *testing.T) {
	ctx := embedPostContext(t)

	html, err := render.PostMarkdown("Use `{{ \"example.go\" }}` to embed.\n\n{{ \"example.go\" }}\n", ctx)
	require.NoError(t, err)
	assert.Contains(t, html, "<code>{{ &quot;example.go&quot; }}</code>")
	assert.Equal(t, 1, strings.Count(html, `data-file="example.go"`))
}

func TestPostMarkdownLeavesBareEmbedSyntaxInCodeAlone(t *testing.T) {
	ctx := embedPostContext(t)

	html, err := render.PostMarkdown("`{{ example.go }}`\n\n```\n{{ example.go }}\n```\n\n{{ example.go }}\n", ctx)
	require.NoError(t, err)
	assert.Contains(t, html, "<code>{{ example.go }}</code>")
	assert.Contains(t, html, "<pre")
	assert.Equal(t, 1, strings.Count(html, `data-file="example.go"`))
}

func embedPostContext(t *testing.T) render.PostContext {
	t.Helper()
	return render.PostContext{
		Rendered: map[string]render.RenderedFile{
			"example.go": render.HighlightedFile{
				File:  &git.File{Filename: "example.go"},
				Lines: []string{"package main"},
			},
		},
		RenderFile: testRenderFile(t),
	}
}

func TestPostMarkdownMissingEmbedFile(t *testing.T) {
	html, err := render.PostMarkdown(`{{ "missing.go" }}`, render.PostContext{
		Rendered:   map[string]render.RenderedFile{},
		RenderFile: func(render.RenderedFile) string { return "" },
	})
	require.NoError(t, err)
	assert.True(t, strings.Contains(html, "File not found"))
}

func TestBuildRenderedFileMapIncludesCSVFile(t *testing.T) {
	file := &git.File{
		Filename: "data.csv",
		Content:  "name,age\nJohn,30\nJane,25",
		MimeType: git.DetectMimeType([]byte("name,age\nJohn,30\nJane,25"), ".csv"),
	}
	require.True(t, file.MimeType.IsCSV())

	rendered := render.RenderFilesForPost([]*git.File{file})
	require.Len(t, rendered, 1)
	assert.Equal(t, "CSVFile", rendered[0].InternalType())

	m := render.BuildRenderedFileMap(rendered)
	assert.Contains(t, m, "data.csv")
}

func TestPostMarkdownEmbedsCSVFile(t *testing.T) {
	file := &git.File{
		Filename: "data.csv",
		Content:  "name,age\nJohn,30\nJane,25",
		MimeType: git.DetectMimeType([]byte("name,age\nJohn,30\nJane,25"), ".csv"),
	}
	rendered := render.RenderFilesForPost([]*git.File{file})
	m := render.BuildRenderedFileMap(rendered)

	html, err := render.PostMarkdown(`{{ "data.csv" }}`, render.PostContext{
		Rendered: m,
		RenderFile: func(file render.RenderedFile) string {
			csv := file.(*render.CSVFile)
			return `<div data-file="` + csv.Filename + `"></div>`
		},
	})
	require.NoError(t, err)
	assert.Contains(t, html, `data-file="data.csv"`)
	assert.NotContains(t, html, "File not found")
}

func TestPostMarkdownRewritesRepoImageToRawURL(t *testing.T) {
	rawURL := "https://example.com/user/gist/raw/HEAD/item-csv.png"
	ctx := render.PostContext{
		Rendered: map[string]render.RenderedFile{
			"item-csv.png": render.NonHighlightedFile{
				File: &git.File{Filename: "item-csv.png", MimeType: git.MimeType{ContentType: "image/png"}},
				Type: "Image (PNG)",
			},
		},
		RawFileURL: func(filename string) string {
			return "https://example.com/user/gist/raw/HEAD/" + filename
		},
	}

	html, err := render.PostMarkdown("![screenshot](item-csv.png)\n", ctx)
	require.NoError(t, err)
	assert.Contains(t, html, `src="`+rawURL+`"`)
	assert.Contains(t, html, `alt="screenshot"`)
	assert.NotContains(t, html, `data-file=`)
}

func TestPostMarkdownStyledRepoImageWidthPx(t *testing.T) {
	rawURL := "https://example.com/user/gist/raw/HEAD/item-csv.png"
	ctx := imagePostContext(t)

	html, err := render.PostMarkdown("![screenshot](item-csv.png){width=400}\n", ctx)
	require.NoError(t, err)
	assert.Contains(t, html, `src="`+rawURL+`"`)
	assert.Contains(t, html, `style="max-width:100%; width:400px; display:block; margin-left:auto; margin-right:auto"`)
}

func TestPostMarkdownStyledRepoImageWidthPercent(t *testing.T) {
	ctx := imagePostContext(t)

	html, err := render.PostMarkdown("![screenshot](item-csv.png){width=50%}\n", ctx)
	require.NoError(t, err)
	assert.Contains(t, html, `style="max-width:100%; width:50%; display:block; margin-left:auto; margin-right:auto"`)
}

func TestPostMarkdownStyledRepoImageAlignCenter(t *testing.T) {
	ctx := imagePostContext(t)

	html, err := render.PostMarkdown("![screenshot](item-csv.png){width=400px align=center}\n", ctx)
	require.NoError(t, err)
	assert.Contains(t, html, `width:400px`)
	assert.Contains(t, html, `margin-left:auto`)
	assert.Contains(t, html, `margin-right:auto`)
}

func TestPostMarkdownStyledExternalImage(t *testing.T) {
	ctx := imagePostContext(t)

	html, err := render.PostMarkdown("![remote](https://example.com/image.png){width=50% align=right}\n", ctx)
	require.NoError(t, err)
	assert.Contains(t, html, `src="https://example.com/image.png"`)
	assert.Contains(t, html, `width:50%`)
	assert.Contains(t, html, `float:right`)
}

func TestPostMarkdownStyledRepoImageFloatAndMargin(t *testing.T) {
	ctx := imagePostContext(t)

	html, err := render.PostMarkdown("![screenshot](item-csv.png){float=left width=200 margin=0 1rem 1rem 0}\n", ctx)
	require.NoError(t, err)
	assert.Contains(t, html, `float:left`)
	assert.Contains(t, html, `width:200px`)
	assert.Contains(t, html, `margin:0 1rem 1rem 0`)
}

func imagePostContext(t *testing.T) render.PostContext {
	t.Helper()
	return render.PostContext{
		Rendered: map[string]render.RenderedFile{
			"item-csv.png": render.NonHighlightedFile{
				File: &git.File{Filename: "item-csv.png", MimeType: git.MimeType{ContentType: "image/png"}},
				Type: "Image (PNG)",
			},
		},
		RawFileURL: func(filename string) string {
			return "https://example.com/user/gist/raw/HEAD/" + filename
		},
	}
}

func TestGistRawFileURL(t *testing.T) {
	assert.Equal(t,
		"https://opengist.example/user/mygist/raw/HEAD/photo.png",
		render.GistRawFileURL("https://opengist.example", "user", "mygist", "HEAD", "photo.png"),
	)
}

func TestPostMarkdownLeavesExternalImageURLsAlone(t *testing.T) {
	ctx := render.PostContext{
		Rendered: map[string]render.RenderedFile{
			"item-csv.png": render.NonHighlightedFile{
				File: &git.File{Filename: "item-csv.png", MimeType: git.MimeType{ContentType: "image/png"}},
			},
		},
		RawFileURL: func(filename string) string {
			return "https://example.com/user/gist/raw/HEAD/" + filename
		},
	}

	html, err := render.PostMarkdown("![remote](https://example.com/image.png)\n", ctx)
	require.NoError(t, err)
	assert.Contains(t, html, `https://example.com/image.png`)
	assert.NotContains(t, html, `data-file=`)
}

func TestPostMarkdownLeavesNonImageRepoFileAsImageTag(t *testing.T) {
	ctx := render.PostContext{
		Rendered: map[string]render.RenderedFile{
			"example.go": render.HighlightedFile{
				File:  &git.File{Filename: "example.go"},
				Lines: []string{"package main"},
			},
		},
		RenderFile: testRenderFile(t),
	}

	html, err := render.PostMarkdown("![code](example.go)\n", ctx)
	require.NoError(t, err)
	assert.Contains(t, html, `src="example.go"`)
	assert.NotContains(t, html, `data-file="example.go"`)
}
