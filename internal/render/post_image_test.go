package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseImageAttributes(t *testing.T) {
	t.Run("WidthPxDefault", func(t *testing.T) {
		attrs := parseImageAttributes("width=400")
		assert.Equal(t, "400px", attrs.width)
		assert.Empty(t, attrs.align)
	})

	t.Run("WidthPxExplicit", func(t *testing.T) {
		attrs := parseImageAttributes("width=400px")
		assert.Equal(t, "400px", attrs.width)
	})

	t.Run("WidthPercent", func(t *testing.T) {
		attrs := parseImageAttributes("width=50%")
		assert.Equal(t, "50%", attrs.width)
	})

	t.Run("AlignCenter", func(t *testing.T) {
		attrs := parseImageAttributes("align=center")
		assert.Equal(t, "center", attrs.align)
	})

	t.Run("Multiple", func(t *testing.T) {
		attrs := parseImageAttributes("width=50% align=right")
		assert.Equal(t, "50%", attrs.width)
		assert.Equal(t, "right", attrs.align)
	})

	t.Run("IgnoresInvalidWidth", func(t *testing.T) {
		attrs := parseImageAttributes("width=bad align=center")
		assert.Empty(t, attrs.width)
		assert.Equal(t, "center", attrs.align)
	})

	t.Run("FloatRight", func(t *testing.T) {
		attrs := parseImageAttributes("float=right")
		assert.Equal(t, "right", attrs.float)
	})

	t.Run("MarginShorthand", func(t *testing.T) {
		attrs := parseImageAttributes("margin=0 1rem 1rem 0")
		assert.Equal(t, "0 1rem 1rem 0", attrs.margin)
	})

	t.Run("MarginSingleRem", func(t *testing.T) {
		attrs := parseImageAttributes("margin=1rem")
		assert.Equal(t, "1rem", attrs.margin)
	})

	t.Run("FloatAndMargin", func(t *testing.T) {
		attrs := parseImageAttributes("float=left margin=8px width=200")
		assert.Equal(t, "left", attrs.float)
		assert.Equal(t, "8px", attrs.margin)
		assert.Equal(t, "200px", attrs.width)
	})
}

func TestBuildImageStyleDefaultsToCenter(t *testing.T) {
	style := buildImageStyle(imageAttributes{width: "400px"})
	assert.Contains(t, style, "display:block")
	assert.Contains(t, style, "margin-left:auto")
	assert.Contains(t, style, "margin-right:auto")
}

func TestBuildImageStyle(t *testing.T) {
	style := buildImageStyle(imageAttributes{width: "400px", align: "center"})
	assert.Contains(t, style, "width:400px")
	assert.Contains(t, style, "margin-left:auto")
	assert.Contains(t, style, "margin-right:auto")
}

func TestBuildImageStyleFloatWithoutDefaultMargin(t *testing.T) {
	style := buildImageStyle(imageAttributes{float: "left", width: "200px"})
	assert.Contains(t, style, "float:left")
	assert.NotContains(t, style, "margin:")
}

func TestBuildImageStyleFloatWithMargin(t *testing.T) {
	style := buildImageStyle(imageAttributes{float: "right", margin: "0 0 1rem 1rem"})
	assert.Contains(t, style, "float:right")
	assert.Contains(t, style, "margin:0 0 1rem 1rem")
}
