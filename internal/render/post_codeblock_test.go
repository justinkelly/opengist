package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtectFencedCodeBlocks(t *testing.T) {
	input := "# Title\n\n```markdown\n![x](y.png){align=left}\n```\n\n![real](z.png)\n"
	blocks := make([]string, 0)
	protected := protectFencedCodeBlocks(input, &blocks)
	assert.Len(t, blocks, 1)
	assert.Contains(t, blocks[0], "![x](y.png){align=left}")
	assert.Contains(t, protected, "![real](z.png)")
	assert.NotContains(t, protected, "![x](y.png)")
}

func TestProtectInlineCode(t *testing.T) {
	input := "Use `![x](y.png){align=left}` here"
	blocks := make([]string, 0)
	protected := protectInlineCode(input, &blocks)
	assert.Len(t, blocks, 1)
	assert.Equal(t, "`![x](y.png){align=left}`", blocks[0])
	assert.NotContains(t, protected, "![x](y.png)")
}

func TestProtectInlineCodeEmbedSyntax(t *testing.T) {
	input := "Use `{{ \"example.go\" }}` here"
	blocks := make([]string, 0)
	protected := protectInlineCode(input, &blocks)
	assert.Len(t, blocks, 1)
	assert.Equal(t, "`{{ \"example.go\" }}`", blocks[0])
	assert.NotContains(t, protected, "example.go")
}

func TestRestoreProtectedBlocks(t *testing.T) {
	blocks := []string{"```\n![x](y.png)\n```"}
	restored := restoreProtectedBlocks("before\x00OGCODE0\x00after", blocks)
	assert.Equal(t, "before```\n![x](y.png)\n```after", restored)
}
