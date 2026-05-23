package cli

import (
	"fmt"
	"io"
	"runtime/debug"

	"github.com/w-h-a/assay/internal/service"
)

// Handler is the CLI handler for assay commands.
type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// Version writes version information to out.
func (h *Handler) Version(out io.Writer) error {
	v, c, d := "dev", "unknown", "unknown"

	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			v = info.Main.Version
		}
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				c = s.Value
			case "vcs.time":
				d = s.Value
			}
		}
	}

	short := c
	if len(short) > 7 {
		short = short[:7]
	}

	fmt.Fprintf(out, "assay %s (commit: %s, built: %s)\n", v, short, d)
	return nil
}

// Check validates a spec file. Diagnostic errors are written to stderr.
func (h *Handler) Check(stderr io.Writer, specPath string) error {
	errs := h.svc.Check(specPath)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(stderr, e)
		}
		return fmt.Errorf("check failed: %d error(s)", len(errs))
	}

	return nil
}
