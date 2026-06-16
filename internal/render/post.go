package render

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	embedQuotedRegex            = regexp.MustCompile(`(?m)\{\{\s*"([^"]+)"\s*\}\}`)
	embedBareRegex              = regexp.MustCompile(`(?m)\{\{\s*([^\s"{}]+)\s*\}\}`)
	gistEmbedPlaceholderRegex   = regexp.MustCompile(`<div class="gist-embed-placeholder" data-gist-embed="(\d+)"></div>`)
)

type PostContext struct {
	Rendered   map[string]RenderedFile
	RenderFile func(RenderedFile) string
	RawFileURL func(filename string) string
}

// GistRawFileURL builds the /raw/ URL for a gist file at a given revision.
func GistRawFileURL(externalURL, username, identifier, revision, filename string) string {
	return strings.TrimRight(externalURL, "/") + "/" + username + "/" + identifier + "/raw/" + revision + "/" + filename
}

func PostMarkdown(content string, ctx PostContext) (string, error) {
	embeds := make([]string, 0)
	processed := replaceFileEmbeds(content, ctx, &embeds)

	html, err := PostMarkdownString(processed)
	if err != nil {
		return "", err
	}

	html = gistEmbedPlaceholderRegex.ReplaceAllStringFunc(html, func(placeholder string) string {
		submatches := gistEmbedPlaceholderRegex.FindStringSubmatch(placeholder)
		if len(submatches) != 2 {
			return placeholder
		}
		idx := 0
		fmt.Sscanf(submatches[1], "%d", &idx)
		if idx < 0 || idx >= len(embeds) {
			return placeholder
		}
		return embeds[idx]
	})

	return html, nil
}

func replaceFileEmbeds(content string, ctx PostContext, embeds *[]string) string {
	// Skip inline code and fenced code blocks so embed/image syntax inside them stays literal.
	return replaceOutsideMarkdownCode(content, func(segment string) string {
		segment = embedQuotedRegex.ReplaceAllStringFunc(segment, func(match string) string {
			filename := strings.TrimSpace(embedQuotedRegex.FindStringSubmatch(match)[1])
			return replaceEmbedFilename(filename, ctx, embeds)
		})

		segment = embedBareRegex.ReplaceAllStringFunc(segment, func(match string) string {
			filename := strings.TrimSpace(embedBareRegex.FindStringSubmatch(match)[1])
			return replaceEmbedFilename(filename, ctx, embeds)
		})

		return replaceMarkdownImageEmbeds(segment, ctx)
	})
}

func replaceMarkdownImageEmbeds(content string, ctx PostContext) string {
	if ctx.RawFileURL == nil {
		return content
	}

	return markdownImageWithAttrsRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := markdownImageWithAttrsRegex.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		alt := submatches[1]
		attrsRaw := ""
		if len(submatches) > 3 {
			attrsRaw = submatches[3]
		}
		attrs := parseImageAttributes(attrsRaw)

		resolvedSrc, isRepoImage, ok := resolvePostImageSrc(submatches[2], ctx)
		if isRepoImage && !ok {
			return match
		}

		if attrs.hasStyle() {
			return "\n\n" + renderStyledImageHTML(alt, resolvedSrc, attrs) + "\n\n"
		}

		if isRepoImage {
			return fmt.Sprintf("![%s](%s)", alt, resolvedSrc)
		}

		return match
	})
}

func isExternalImageURL(path string) bool {
	return strings.Contains(path, "://") || strings.HasPrefix(path, "//")
}

func isImageRenderedFile(rendered RenderedFile) bool {
	switch f := rendered.(type) {
	case NonHighlightedFile:
		return f.MimeType.IsImage()
	case HighlightedFile:
		return f.MimeType.IsSVG()
	default:
		return false
	}
}

func renderedFileFilename(rendered RenderedFile) string {
	switch f := rendered.(type) {
	case HighlightedFile:
		return f.Filename
	case NonHighlightedFile:
		return f.Filename
	case CSVFile:
		return f.Filename
	case *CSVFile:
		return f.Filename
	default:
		return ""
	}
}

func replaceEmbedFilename(filename string, ctx PostContext, embeds *[]string) string {
	if IsReadmeFilename(filename) {
		idx := len(*embeds)
		*embeds = append(*embeds, `<div class="gist-embed-error rounded-md border border-red-300 dark:border-red-700 bg-red-50 dark:bg-red-900/30 px-4 py-2 text-sm text-red-900 dark:text-red-200">Cannot embed README.md</div>`)
		return embedPlaceholder(idx)
	}

	idx := len(*embeds)
	rendered, found := findRenderedFileByName(ctx.Rendered, filename)
	if !found {
		*embeds = append(*embeds, `<div class="gist-embed-error rounded-md border border-red-300 dark:border-red-700 bg-red-50 dark:bg-red-900/30 px-4 py-2 text-sm text-red-900 dark:text-red-200">File not found: `+filename+`</div>`)
		return embedPlaceholder(idx)
	}

	*embeds = append(*embeds, ctx.RenderFile(rendered))
	return embedPlaceholder(idx)
}

func embedPlaceholder(idx int) string {
	return fmt.Sprintf("\n\n<div class=\"gist-embed-placeholder\" data-gist-embed=\"%d\"></div>\n\n", idx)
}

func BuildRenderedFileMap(files []RenderedFile) map[string]RenderedFile {
	result := make(map[string]RenderedFile, len(files))
	for _, file := range files {
		switch f := file.(type) {
		case HighlightedFile:
			result[f.Filename] = file
		case CSVFile:
			result[f.Filename] = file
		case *CSVFile:
			result[f.Filename] = file
		case NonHighlightedFile:
			result[f.Filename] = file
		}
	}
	return result
}
