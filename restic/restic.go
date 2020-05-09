package restic

import "os/exec"

// Making possible to mock exec.CommandContext
var execCommandContext = exec.CommandContext

