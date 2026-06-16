package render

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomiceli/opengist/internal/db"
)

func TestRenderGistMetadataFile(t *testing.T) {
	labels := GistMetadataLabels{
		Title:       "Title",
		Description: "Description",
		Topics:      "Topics",
		URL:         "URL",
		PublishedAt: "Published date",
		Visibility:  "Visibility",
		Public:      "Public",
		Unlisted:    "Unlisted",
		Private:     "Private",
	}

	gist := &db.Gist{
		Title:       "My Gist",
		Description: "A sample gist",
		Topics:      []db.GistTopic{{Topic: "go"}, {Topic: "web"}},
		URL:         "my-slug",
		PublishedAt: time.Date(2024, 6, 13, 0, 0, 0, 0, time.UTC).Unix(),
		Private:     db.UnlistedVisibility,
	}

	file, err := RenderGistMetadataFile(gist, labels)
	require.NoError(t, err)
	assert.Equal(t, "metadata.md", file.Filename)
	assert.Contains(t, file.Content, "Title: My Gist")
	assert.Contains(t, file.Content, "Description: A sample gist")
	assert.Contains(t, file.Content, "Topics: go, web")
	assert.Contains(t, file.Content, "URL: my-slug")
	assert.Contains(t, file.Content, "Visibility: Unlisted")
	assert.Len(t, file.Lines, 6)
	assert.Equal(t, "markdown", file.Type)
}

func TestRenderGistMetadataFileEmpty(t *testing.T) {
	file, err := RenderGistMetadataFile(&db.Gist{}, GistMetadataLabels{})
	require.NoError(t, err)
	assert.Nil(t, file.Lines)
}

func TestRenderGistMetadataFilePublicVisibility(t *testing.T) {
	file, err := RenderGistMetadataFile(&db.Gist{
		Title:   "Hello",
		Private: db.PublicVisibility,
	}, GistMetadataLabels{
		Title:      "Title",
		Visibility: "Visibility",
		Public:     "Public",
	})
	require.NoError(t, err)
	assert.Contains(t, file.Content, "Title: Hello")
	assert.Contains(t, file.Content, "Visibility: Public")
}

func TestRenderGistMetadataFileNoTrailingBlankLine(t *testing.T) {
	file, err := RenderGistMetadataFile(&db.Gist{
		Title:       "Post title",
		Description: "this is a test sql",
		Topics:      []db.GistTopic{{Topic: "php"}, {Topic: "sql"}, {Topic: "test"}},
		Private:     db.PublicVisibility,
	}, GistMetadataLabels{
		Title:       "Title",
		Description: "Description",
		Topics:      "Topics",
		Visibility:  "Visibility",
		Public:      "Public",
	})
	require.NoError(t, err)
	assert.Len(t, file.Lines, 4)
	assert.Equal(t, "Title: Post title\nDescription: this is a test sql\nTopics: php, sql, test\nVisibility: Public", file.Content)
}
