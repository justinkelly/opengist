package gist

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/thomiceli/opengist/internal/db"
	"github.com/thomiceli/opengist/internal/git"
	"github.com/thomiceli/opengist/internal/render"
	"github.com/thomiceli/opengist/internal/web/context"
)

func GistRoot(ctx *context.Context) error {
	if ctx.GetData("gistpage") == "js" {
		return GistJs(ctx)
	} else if ctx.GetData("gistpage") == "json" {
		return GistJson(ctx)
	}

	if hasPostTab, ok := ctx.GetData("hasPostTab").(bool); ok && hasPostTab {
		gist := ctx.GetData("gist").(*db.Gist)
		return ctx.Redirect(302, "/"+gist.User.Username+"/"+gist.Identifier()+"/post")
	}

	return GistIndex(ctx)
}

func GistIndex(ctx *context.Context) error {
	if ctx.GetData("gistpage") == "js" {
		return GistJs(ctx)
	} else if ctx.GetData("gistpage") == "json" {
		return GistJson(ctx)
	}

	gist := ctx.GetData("gist").(*db.Gist)
	revision := ctx.Param("revision")

	if revision == "" {
		revision = "HEAD"
	}

	files, hasMoreFiles, err := gist.Files(revision, true)
	if _, ok := err.(*git.RevisionNotFoundError); ok {
		return ctx.NotFound("Revision not found")
	} else if err != nil {
		return ctx.ErrorRes(500, "Error fetching files", err)
	}

	renderedFiles := render.RenderFiles(files)

	ctx.SetData("page", "code")
	ctx.SetData("commit", revision)
	ctx.SetData("files", renderedFiles)
	ctx.SetData("hasMoreFiles", hasMoreFiles)
	ctx.SetData("revision", revision)
	ctx.SetData("htmlTitle", gist.Title)
	return ctx.Html("gist.html")
}

func Post(ctx *context.Context) error {
	gist := ctx.GetData("gist").(*db.Gist)
	revision := ctx.Param("revision")

	if revision == "" {
		revision = "HEAD"
	}

	files, hasMoreFiles, err := gist.Files(revision, true)
	if _, ok := err.(*git.RevisionNotFoundError); ok {
		return ctx.NotFound("Revision not found")
	} else if err != nil {
		return ctx.ErrorRes(500, "Error fetching files", err)
	}

	if len(files) == 0 {
		return ctx.Redirect(302, "/"+gist.User.Username+"/"+gist.Identifier())
	}

	readme := render.FindReadmeFile(files)
	if readme != nil {
		files, _, err = gist.Files(revision, false)
		if err != nil {
			return ctx.ErrorRes(500, "Error fetching files", err)
		}
		readme = render.FindReadmeFile(files)
	}

	renderedFiles := render.RenderFiles(files)

	ctx.SetData("commit", revision)
	ctx.SetData("revision", revision)
	ctx.SetData("page", "post")
	ctx.SetData("htmlTitle", gist.Title)

	if readme != nil {
		renderedMap := render.BuildRenderedFileMap(renderedFiles)

		postHTML, err := render.PostMarkdown(readme.Content, render.PostContext{
			Rendered: renderedMap,
			RenderFile: func(file render.RenderedFile) string {
				html, err := renderGistFileHTML(ctx, file)
				if err != nil {
					return `<div class="gist-embed-error rounded-md border border-red-300 dark:border-red-700 bg-red-50 dark:bg-red-900/30 px-4 py-2 text-sm text-red-900 dark:text-red-200">Failed to render file</div>`
				}
				return html
			},
		})
		if err != nil {
			return ctx.ErrorRes(500, "Error rendering post", err)
		}

		ctx.SetData("postHTML", postHTML)
		return ctx.Html("post.html")
	}

	var postHTML string
	if gist.Description != "" {
		postHTML, err = render.PostMarkdownString(gist.Description)
		if err != nil {
			return ctx.ErrorRes(500, "Error rendering post", err)
		}
	}

	ctx.SetData("postHTML", postHTML)
	ctx.SetData("postFiles", renderedFiles)
	ctx.SetData("hasMoreFiles", hasMoreFiles)
	return ctx.Html("post.html")
}

func renderGistFileHTML(ctx *context.Context, file render.RenderedFile) (string, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	data := echo.Map{
		"file": file,
		"root": ctx.DataMap(),
	}
	if err := ctx.Echo().Renderer.Render(w, "_gist_file", data, ctx); err != nil {
		return "", fmt.Errorf("render gist file partial: %w", err)
	}
	if err := w.Flush(); err != nil {
		return "", err
	}
	return buf.String(), nil
}
