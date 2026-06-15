package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParsePublishedAtDate(t *testing.T) {
	require.Equal(t, int64(0), parsePublishedAtDate(""))
	require.Equal(t, int64(0), parsePublishedAtDate("not-a-date"))

	want := time.Date(2024, 6, 13, 0, 0, 0, 0, time.UTC).Unix()
	require.Equal(t, want, parsePublishedAtDate("2024-06-13"))
}

func TestFormatPublishedAtDate(t *testing.T) {
	require.Equal(t, "", formatPublishedAtDate(0))

	unix := time.Date(2024, 6, 13, 0, 0, 0, 0, time.UTC).Unix()
	require.Equal(t, "2024-06-13", formatPublishedAtDate(unix))
}

func TestGistDTOPublishedAtRoundTrip(t *testing.T) {
	dto := &GistDTO{
		Title:           "Post",
		PublishedAtDate: "2024-06-13",
	}
	gist := dto.ToGist()
	require.Equal(t, time.Date(2024, 6, 13, 0, 0, 0, 0, time.UTC).Unix(), gist.PublishedAt)

	dto.PublishedAtDate = ""
	gist = dto.ToExistingGist(gist)
	require.Equal(t, int64(0), gist.PublishedAt)
}

func TestGistDTOHasMetadataIncludesPublishedAtDate(t *testing.T) {
	dto := &GistDTO{PublishedAtDate: "2024-06-13"}
	require.True(t, dto.HasMetadata())
}
