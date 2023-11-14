package restic

import (
	"fmt"
	"os/exec"
)

// Making possible to mock exec.CommandContext
var execCommandContext = exec.CommandContext

func buildCmdEnv(repositoryPassword string, env *map[string]string) []string {
	var cmdEnv []string
	cmdEnv = append(cmdEnv, "RESTIC_PASSWORD="+repositoryPassword)
	for k, v := range *env {
		cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", k, v))
	}

	return cmdEnv
}
