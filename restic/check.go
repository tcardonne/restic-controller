package restic

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// RunIntegrityCheck checks the integrity of a given repository.
// Returns a boolean indicating if the repository is healthy.
func RunIntegrityCheck(repository string, password string, env *map[string]string) (bool, error) {
	ctx := context.TODO()

	cmd := execCommandContext(ctx, "restic", "-r", repository, "check", "-q", "--no-lock")
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, buildCmdEnv(password, env)...)

	log.WithFields(log.Fields{"component": "restic", "cmd": strings.Join(cmd.Args, " ")}).Debug("Running restic check command")
	_, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			return false, fmt.Errorf("restic command returned with code %d : %s", exiterr.ExitCode(), exiterr.Stderr)
		}

		return false, err
	}

	return true, nil
}
