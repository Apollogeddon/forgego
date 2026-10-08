package features

import (
	"github.com/apollogeddon/forgego/internal/config"
	"github.com/apollogeddon/forgego/internal/taskfile"
	"github.com/apollogeddon/forgego/internal/templates"
)

// Docker adds a Dockerfile: a distroless binary for a backend, nginx for a website.
type Docker struct{}

func (Docker) Name() string                   { return "docker" }
func (Docker) ShouldRun(cfg config.Init) bool { return cfg.Docker }

func (Docker) Apply(ctx *Context) bool {
	image := ctx.Project.Slug // image names must be lowercase
	values := map[string]string{"NAME": ctx.Project.Name, "GO_MINOR": goMinor(ctx.Cfg.Go)}
	dockerfile, run := templates.DockerfileBackend, "docker run --rm "+image
	if ctx.Cfg.IsWebsite() {
		dockerfile, run = templates.DockerfileWebsite, "docker run --rm -p 8080:80 "+image
	}
	ok := CreateFile(ctx, "Dockerfile", templates.Render(dockerfile, values))
	ok = CreateFile(ctx, ".dockerignore", templates.Dockerignore) && ok
	ctx.Tasks.Add(taskfile.Task{Name: "docker:build", Desc: "Build the image", Cmds: []string{"docker build -t " + image + " ."}})
	ctx.Tasks.Add(taskfile.Task{Name: "docker:run", Desc: "Run the image", Cmds: []string{run}})
	return ok
}

func (Docker) Cleanup(ctx *Context) {
	if !ctx.Cfg.Docker {
		RemoveTasks(ctx, "docker:build", "docker:run")
		RemoveFile(ctx, "Dockerfile")
		RemoveFile(ctx, ".dockerignore")
	}
}
