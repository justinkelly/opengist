package render

import (
	"html"
	"regexp"
	"strings"
)

var (
	markdownImageWithAttrsRegex = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)(?:\s*\{([^}]+)\})?`)
	imageWidthValueRegex        = regexp.MustCompile(`^\d+(\.\d+)?(%|px)?$`)
	imageMarginPartRegex        = regexp.MustCompile(`^\d+(\.\d+)?(px|rem|em|%)?$`)
)

type imageAttributes struct {
	width  string // CSS value, e.g. "400px" or "50%"
	align  string // left, center, or right
	float  string // left or right
	margin string // CSS margin shorthand, e.g. "1rem" or "0 1rem 1rem 0"
}

func parseImageAttributes(raw string) imageAttributes {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return imageAttributes{}
	}

	tokens := strings.Fields(raw)
	var attrs imageAttributes
	for idx := 0; idx < len(tokens); {
		part := tokens[idx]
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			idx++
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)

		switch key {
		case "margin":
			marginParts := []string{value}
			idx++
			for idx < len(tokens) && !strings.Contains(tokens[idx], "=") && len(marginParts) < 4 {
				marginParts = append(marginParts, tokens[idx])
				idx++
			}
			if m := sanitizeImageMargin(strings.Join(marginParts, " ")); m != "" {
				attrs.margin = m
			}
			continue
		case "width":
			if w := sanitizeImageWidth(value); w != "" {
				attrs.width = w
			}
		case "align":
			switch strings.ToLower(value) {
			case "left", "center", "right":
				attrs.align = strings.ToLower(value)
			}
		case "float":
			switch strings.ToLower(value) {
			case "left", "right":
				attrs.float = strings.ToLower(value)
			}
		}
		idx++
	}
	return attrs
}

func sanitizeImageWidth(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || !imageWidthValueRegex.MatchString(value) {
		return ""
	}
	if strings.HasSuffix(value, "%") || strings.HasSuffix(value, "px") {
		return value
	}
	return value + "px"
}

func sanitizeImageMargin(value string) string {
	parts := strings.Fields(strings.TrimSpace(value))
	if len(parts) == 0 || len(parts) > 4 {
		return ""
	}
	for _, part := range parts {
		if !imageMarginPartRegex.MatchString(part) {
			return ""
		}
	}
	return strings.Join(parts, " ")
}

func (a imageAttributes) hasStyle() bool {
	return a.width != "" || a.align != "" || a.float != "" || a.margin != ""
}

func buildImageStyle(attrs imageAttributes) string {
	if !attrs.hasStyle() {
		return ""
	}

	styles := []string{"max-width:100%"}
	if attrs.width != "" {
		styles = append(styles, "width:"+attrs.width)
	}

	floatDir := attrs.float
	if floatDir == "" && (attrs.align == "left" || attrs.align == "right") {
		floatDir = attrs.align
	}
	if floatDir == "left" || floatDir == "right" {
		styles = append(styles, "float:"+floatDir)
	}

	if attrs.margin != "" {
		styles = append(styles, "margin:"+attrs.margin)
	} else {
		switch attrs.align {
		case "left":
			styles = append(styles, "margin:0 1rem 1rem 0")
		case "right":
			styles = append(styles, "margin:0 0 1rem 1rem")
		default:
			if floatDir == "" {
				styles = append(styles, "display:block", "margin-left:auto", "margin-right:auto")
			}
		}
	}

	return strings.Join(styles, "; ")
}

func renderStyledImageHTML(alt, src string, attrs imageAttributes) string {
	style := buildImageStyle(attrs)
	if style == "" {
		return `<img src="` + html.EscapeString(src) + `" alt="` + html.EscapeString(alt) + `" />`
	}
	return `<img src="` + html.EscapeString(src) + `" alt="` + html.EscapeString(alt) + `" style="` + html.EscapeString(style) + `" />`
}

func resolvePostImageSrc(src string, ctx PostContext) (resolved string, isRepoImage bool, ok bool) {
	src = strings.TrimSpace(src)
	src = strings.TrimPrefix(src, "./")
	if isExternalImageURL(src) {
		return src, false, true
	}

	rendered, found := findRenderedFileByName(ctx.Rendered, src)
	if !found || !isImageRenderedFile(rendered) {
		return src, false, false
	}
	if ctx.RawFileURL == nil {
		return src, true, false
	}
	return ctx.RawFileURL(renderedFileFilename(rendered)), true, true
}
