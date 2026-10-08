// Package config holds the resolved options for one `forgego init` run.
package config

// Mode is the kind of project being scaffolded.
type Mode string

const (
	Backend Mode = "backend"
	Library Mode = "library"
	Website Mode = "website"
)

// DefaultGo is the Go version new projects target.
const DefaultGo = "1.27"

// Init is the resolved configuration for `forgego init`.
type Init struct {
	Mode       Mode
	Force      bool
	DryRun     bool
	Testing    bool
	Linting    bool
	Versioning bool
	Docker     bool
	Debian     bool
	Go         string
	Target     string
}

// Default returns a backend configuration with every standard feature on.
func Default(target string) Init {
	return Init{
		Mode:       Backend,
		Testing:    true,
		Linting:    true,
		Versioning: true,
		Go:         DefaultGo,
		Target:     target,
	}
}

func (c Init) IsBackend() bool { return c.Mode == Backend }
func (c Init) IsLibrary() bool { return c.Mode == Library }
func (c Init) IsWebsite() bool { return c.Mode == Website }

// Validate returns every reason the combination of options can't be scaffolded.
func (c Init) Validate() []string {
	var errs []string
	if c.Docker && c.IsLibrary() {
		errs = append(errs, "--docker is not available in library mode")
	}
	if c.Debian && !c.IsBackend() {
		errs = append(errs, "--debian is only available in backend mode")
	}
	if !validGoVersion(c.Go) {
		errs = append(errs, "--go must be a Go release such as 1.27 or 1.27.1, got "+c.Go)
	}
	return errs
}

func validGoVersion(v string) bool {
	parts := 0
	digits := 0
	for _, r := range v {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case r == '.' && digits > 0:
			parts++
			digits = 0
		default:
			return false
		}
	}
	return digits > 0 && (parts == 1 || parts == 2) && v[0] == '1'
}

// GoDirective is the version for go.mod's go line: 1.27 becomes 1.27.0.
func (c Init) GoDirective() string {
	dots := 0
	for _, r := range c.Go {
		if r == '.' {
			dots++
		}
	}
	if dots == 1 {
		return c.Go + ".0"
	}
	return c.Go
}
