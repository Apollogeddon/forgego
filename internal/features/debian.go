package features

import (
	"github.com/apollogeddon/forgego/internal/config"
	"github.com/apollogeddon/forgego/internal/templates"
)

// Debian adds the systemd unit and install scripts the .goreleaser.yaml nfpms section
// packages. The package runs the service as its own system user, never root.
type Debian struct{}

func (Debian) Name() string                   { return "debian" }
func (Debian) ShouldRun(cfg config.Init) bool { return cfg.Debian }

func (Debian) Apply(ctx *Context) bool {
	values := map[string]string{"NAME": ctx.Project.Name}
	ok := CreateFile(ctx, "packaging/"+ctx.Project.Name+".service", templates.Render(templates.SystemdUnit, values))
	ok = CreateFile(ctx, "packaging/postinstall.sh", templates.Render(templates.Postinstall, values)) && ok
	return CreateFile(ctx, "packaging/preremove.sh", templates.Render(templates.Preremove, values)) && ok
}

func (Debian) Cleanup(ctx *Context) {
	if !ctx.Cfg.Debian {
		RemoveFile(ctx, "packaging/"+ctx.Project.Name+".service")
		RemoveFile(ctx, "packaging/postinstall.sh")
		RemoveFile(ctx, "packaging/preremove.sh")
	}
}
