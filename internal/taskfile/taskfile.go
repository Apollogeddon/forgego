// Package taskfile collects the tasks each feature contributes and writes them into the
// project's Taskfile.yml, keeping any task or var the project already defines.
package taskfile

import (
	"go.yaml.in/yaml/v3"

	"github.com/apollogeddon/forgego/internal/yamlx"
)

// Task is one entry under tasks:.
type Task struct {
	Name string
	Desc string
	Deps []string
	Cmds []string
}

// Var is one entry under vars:.
type Var struct {
	Name  string
	Value string
}

// Builder accumulates vars and tasks in the order features add them, and the names of
// tasks and vars to remove.
type Builder struct {
	vars    []Var
	tasks   []Task
	removed []string
}

// Remove drops tasks or vars of these names from the Taskfile when it's rendered.
func (b *Builder) Remove(names ...string) { b.removed = append(b.removed, names...) }

// NewBuilder returns an empty Builder.
func NewBuilder() *Builder { return &Builder{} }

// Var adds a variable, keeping the first value when a name is added twice.
func (b *Builder) Var(name, value string) {
	for _, v := range b.vars {
		if v.Name == name {
			return
		}
	}
	b.vars = append(b.vars, Var{name, value})
}

// Add adds a task, replacing an earlier one of the same name.
func (b *Builder) Add(t Task) {
	for i, existing := range b.tasks {
		if existing.Name == t.Name {
			b.tasks[i] = t
			return
		}
	}
	b.tasks = append(b.tasks, t)
}

// Tasks returns the tasks added so far.
func (b *Builder) Tasks() []Task { return b.tasks }

// Render merges the builder into an existing Taskfile (empty for a new one). A task or
// var the file already has is kept unless force is set; it reports what changed.
func (b *Builder) Render(existing string, force bool) (string, []string, error) {
	root, err := yamlx.Parse(existing)
	if err != nil {
		return "", nil, err
	}
	yamlx.SetIfAbsent(root, "version", quoted("3"), false)

	var changed []string
	if len(b.vars) > 0 {
		vars := yamlx.Child(root, "vars")
		for _, v := range b.vars {
			if yamlx.SetIfAbsent(vars, v.Name, yamlx.Str(v.Value), force) {
				changed = append(changed, "var "+v.Name)
			}
		}
	}
	tasks := yamlx.Child(root, "tasks")
	for _, t := range b.tasks {
		if yamlx.SetIfAbsent(tasks, t.Name, taskNode(t), force) {
			changed = append(changed, "task "+t.Name)
		}
	}
	for _, name := range b.removed {
		for _, section := range []string{"tasks", "vars"} {
			if m := yamlx.Get(root, section); m != nil && yamlx.Delete(m, name) {
				changed = append(changed, "removed "+name)
			}
		}
	}
	out, err := yamlx.Encode(root)
	return out, changed, err
}

func taskNode(t Task) *yaml.Node {
	node := yamlx.Map()
	if t.Desc != "" {
		yamlx.Set(node, "desc", yamlx.Str(t.Desc))
	}
	if len(t.Deps) > 0 {
		yamlx.Set(node, "deps", yamlx.Seq(t.Deps...))
	}
	yamlx.Set(node, "cmds", yamlx.Seq(t.Cmds...))
	return node
}

func quoted(value string) *yaml.Node {
	node := yamlx.Str(value)
	node.Style = yaml.SingleQuotedStyle
	return node
}
