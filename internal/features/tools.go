package features

import (
	"strings"

	"github.com/apollogeddon/forgego/internal/templates"
)

// toolVar is the Taskfile var that runs a tool: GOLANGCI_LINT for golangci-lint.
func toolVar(tool templates.Tool) string {
	return strings.ToUpper(strings.ReplaceAll(tool.Name, "-", "_"))
}

// ref is how a Taskfile command refers to a tool's var.
func ref(tool templates.Tool) string {
	return "{{." + toolVar(tool) + "}}"
}

// useTool pins a tool in .forgego/ and gives the Taskfile a var that runs it.
func useTool(ctx *Context, tool templates.Tool) bool {
	ok := WriteManaged(ctx, tool.ModPath(), tool.ModFile())
	ok = WriteManaged(ctx, tool.SumPath(), tool.SumFile()) && ok
	ctx.Tasks.Var(toolVar(tool), tool.Command())
	ctx.Tools = append(ctx.Tools, tool)
	return ok
}

// dropTool removes a tool's pinned module when the feature that used it is turned off.
func dropTool(ctx *Context, tool templates.Tool) {
	RemoveFile(ctx, tool.ModPath())
	RemoveFile(ctx, tool.SumPath())
}
