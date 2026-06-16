package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/thomiceli/opengist/internal/db"
	"github.com/thomiceli/opengist/internal/git"
)

type GistMetadataLabels struct {
	Title       string
	Description string
	Topics      string
	URL         string
	PublishedAt string
	Visibility  string
	Public      string
	Unlisted    string
	Private     string
}

func RenderGistMetadataFile(gist *db.Gist, labels GistMetadataLabels) (HighlightedFile, error) {
	markdown, err := gistMetadataMarkdown(gist, labels)
	if err != nil || markdown == "" {
		return HighlightedFile{}, err
	}

	file := &git.File{
		Filename:  "metadata.md",
		Content:   markdown,
		Size:      uint64(len(markdown)),
		HumanSize: humanize.Bytes(uint64(len(markdown))),
		MimeType: git.MimeType{ContentType: "text/markdown"},
	}

	return highlightMetadataFile(file)
}

func highlightMetadataFile(file *git.File) (HighlightedFile, error) {
	highlighted, err := highlightFile(file)
	if err != nil {
		return highlighted, err
	}
	highlighted.Lines = trimTrailingEmptyHighlightedLines(highlighted.Lines)
	return highlighted, nil
}

func trimTrailingEmptyHighlightedLines(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func gistMetadataMarkdown(gist *db.Gist, labels GistMetadataLabels) (string, error) {
	if !gist.HasDisplayMetadata() {
		return "", nil
	}

	var lines []string
	if gist.Title != "" {
		lines = append(lines, metadataLine(labels.Title, gist.Title))
	}
	if gist.Description != "" {
		lines = append(lines, metadataLine(labels.Description, gist.Description))
	}
	if len(gist.Topics) > 0 {
		lines = append(lines, metadataLine(labels.Topics, strings.Join(gist.TopicsSlice(), ", ")))
	}
	if gist.URL != "" {
		lines = append(lines, metadataLine(labels.URL, gist.URL))
	}
	if gist.PublishedAt > 0 {
		lines = append(lines, metadataLine(labels.PublishedAt, time.Unix(gist.PublishedAt, 0).Format("02/01/2006")))
	}
	visibility := labels.Public
	if gist.Private == db.UnlistedVisibility {
		visibility = labels.Unlisted
	} else if gist.Private == db.PrivateVisibility {
		visibility = labels.Private
	}
	lines = append(lines, metadataLine(labels.Visibility, visibility))

	return strings.Join(lines, "\n"), nil
}

func metadataLine(label, value string) string {
	return fmt.Sprintf("%s: %s", label, value)
}
