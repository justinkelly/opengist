package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomiceli/opengist/internal/git"
)

func TestIsMarkdownFilename(t *testing.T) {
	assert.True(t, IsMarkdownFilename("notes.md"))
	assert.True(t, IsMarkdownFilename("dir/Guide.MD"))
	assert.False(t, IsMarkdownFilename("main.go"))
}

func TestRenderFilesForPostRendersMarkdownAsHTML(t *testing.T) {
	files := []*git.File{{
		Filename: "guide.md",
		Content:  "# Title\n\nParagraph.",
		MimeType: git.DetectMimeType([]byte("# Title\n\nParagraph."), ".md"),
	}}

	rendered := RenderFilesForPost(files)
	require.Len(t, rendered, 1)

	md, ok := rendered[0].(HighlightedFile)
	require.True(t, ok)
	require.Contains(t, md.HTML, "<h1")
	require.Contains(t, md.HTML, "Title")
	require.Empty(t, md.Lines)
}

func TestRenderFilesKeepsMarkdownHighlighted(t *testing.T) {
	files := []*git.File{{
		Filename: "guide.md",
		Content:  "# Title\n\nParagraph.",
		MimeType: git.DetectMimeType([]byte("# Title\n\nParagraph."), ".md"),
	}}

	rendered := RenderFiles(files)
	require.Len(t, rendered, 1)

	md, ok := rendered[0].(HighlightedFile)
	require.True(t, ok)
	require.Empty(t, md.HTML)
	require.NotEmpty(t, md.Lines)
}
