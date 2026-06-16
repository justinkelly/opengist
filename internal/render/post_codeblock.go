package render

import (
	"strconv"
	"strings"
	"unicode"
)

const codePlaceholderPrefix = "\x00OGCODE"

func replaceOutsideMarkdownCode(content string, replace func(string) string) string {
	blocks := make([]string, 0)
	protected := protectFencedCodeBlocks(content, &blocks)
	protected = protectInlineCode(protected, &blocks)
	protected = replace(protected)
	return restoreProtectedBlocks(protected, blocks)
}

func protectFencedCodeBlocks(content string, blocks *[]string) string {
	if content == "" {
		return content
	}

	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	inFence := false
	var fenceChar rune
	fenceLen := 0
	fenceLines := make([]string, 0)

	flushFence := func() {
		segment := strings.Join(fenceLines, "\n")
		idx := len(*blocks)
		*blocks = append(*blocks, segment)
		out = append(out, codePlaceholderPrefix+strconv.Itoa(idx)+"\x00")
		fenceLines = fenceLines[:0]
		inFence = false
		fenceChar = 0
		fenceLen = 0
	}

	for _, line := range lines {
		if !inFence {
			if openChar, openLen, ok := fenceOpenLine(line); ok {
				inFence = true
				fenceChar = openChar
				fenceLen = openLen
				fenceLines = []string{line}
				continue
			}
			out = append(out, line)
			continue
		}

		fenceLines = append(fenceLines, line)
		if fenceCloseLine(line, fenceChar, fenceLen) {
			flushFence()
		}
	}

	if inFence {
		out = append(out, fenceLines...)
	}

	return strings.Join(out, "\n")
}

func fenceOpenLine(line string) (rune, int, bool) {
	trimmed := strings.TrimLeft(line, " \t")
	if trimmed == "" {
		return 0, 0, false
	}

	runes := []rune(trimmed)
	if runes[0] != '`' && runes[0] != '~' {
		return 0, 0, false
	}

	length := 0
	for length < len(runes) && runes[length] == runes[0] {
		length++
	}
	if length < 3 {
		return 0, 0, false
	}

	return runes[0], length, true
}

func fenceCloseLine(line string, fenceChar rune, fenceLen int) bool {
	trimmed := strings.TrimRight(line, " \t")
	if trimmed == "" {
		return false
	}

	runes := []rune(trimmed)
	if runes[0] != fenceChar {
		return false
	}

	length := 0
	for length < len(runes) && runes[length] == fenceChar {
		length++
	}
	if length < fenceLen {
		return false
	}

	for _, r := range runes[length:] {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func protectInlineCode(content string, blocks *[]string) string {
	var out strings.Builder
	runes := []rune(content)
	i := 0

	for i < len(runes) {
		if runes[i] != '`' {
			out.WriteRune(runes[i])
			i++
			continue
		}

		j := i
		for j < len(runes) && runes[j] == '`' {
			j++
		}
		openLen := j - i
		if openLen > 2 {
			out.WriteRune(runes[i])
			i++
			continue
		}

		k := j
		found := false
		for k < len(runes) {
			if runes[k] != '`' {
				k++
				continue
			}
			closeLen := 0
			for closeLen < openLen && k+closeLen < len(runes) && runes[k+closeLen] == '`' {
				closeLen++
			}
			if closeLen == openLen {
				segment := string(runes[i : k+closeLen])
				idx := len(*blocks)
				*blocks = append(*blocks, segment)
				out.WriteString(codePlaceholderPrefix + strconv.Itoa(idx) + "\x00")
				i = k + closeLen
				found = true
				break
			}
			k++
		}
		if !found {
			out.WriteRune(runes[i])
			i++
		}
	}

	return out.String()
}

func restoreProtectedBlocks(content string, blocks []string) string {
	for i, block := range blocks {
		content = strings.ReplaceAll(content, codePlaceholderPrefix+strconv.Itoa(i)+"\x00", block)
	}
	return content
}
