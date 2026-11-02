package features

import (
	"strings"

	"github.com/apollogeddon/forgego/v2/internal/gomod"

	"github.com/apollogeddon/forgego/v2/internal/templates"
)

// toolVar is the Taskfile var that runs a tool: GOLANGCI_LINT for golangci-lint.
func toolVar(tool templates.Tool) string {
	return strings.ToUpper(strings.ReplaceAll(tool.Name, "-", "_"))
}

// ref is how a Taskfile command refers to a tool's var.
func ref(tool templates.Tool) string {
	return "{{." + toolVar(tool) + "}}"
}

// useTool pins a tool in .forgego/ and gives the Taskfile a var that runs it. A pin the
// project raised past forgego's, as a Dependabot security update does, is kept.
func useTool(ctx *Context, tool templates.Tool) bool {
	pin := gomod.Pinned(ctx.FS, ctx.Cfg.Target, tool)
	ok := WriteManaged(ctx, tool.ModPath(), pin.Mod)
	ok = WriteManaged(ctx, tool.SumPath(), pin.Sum) && ok
	ctx.Tasks.Var(toolVar(tool), tool.Command())
	ctx.Tools = append(ctx.Tools, pin)
	return ok
}

// dropTool removes a tool's pinned module when the feature that used it is turned off.
func dropTool(ctx *Context, tool templates.Tool) {
	RemoveFile(ctx, tool.ModPath())
	RemoveFile(ctx, tool.SumPath())
}
