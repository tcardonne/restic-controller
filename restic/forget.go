package restic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// ForgetPolicy specifies to restic retention policy rules
type ForgetPolicy struct {
	KeepLast    int      `mapstructure:"keep_last"`
	KeepDaily   int      `mapstructure:"keep_daily"`
	KeepHourly  int      `mapstructure:"keep_hourly"`
	KeepWeekly  int      `mapstructure:"keep_weekly"`
	KeepMonthly int      `mapstructure:"keep_monthly"`
	KeepYearly  int      `mapstructure:"keep_yearly"`
	KeepTags    []string `mapstructure:"keep_tags"`
	KeepWithin  string   `mapstructure:"keep_within"`
}

// ForgetResult is the aggregate of result by group of the forget action
type ForgetResult struct {
	GroupResults []ForgetGroupResult
}

// TotalRemove returns the total of snapshots deleted during the forget action
func (r *ForgetResult) TotalRemove() int {
	var total int
	for _, v := range r.GroupResults {
		total += len(v.Remove)
	}
	return total
}

// TotalKeep returns the total of snapshots kept during the forget action
func (r *ForgetResult) TotalKeep() int {
	var total int
	for _, v := range r.GroupResults {
		total += len(v.Keep)
	}
	return total
}

// ForgetGroupResult contains group output for the forget action
type ForgetGroupResult struct {
	Tags   string     `json:"tags"`
	Host   string     `json:"host"`
	Paths  []string   `json:"paths"`
	Keep   []Snapshot `json:"keep"`
	Remove []Snapshot `json:"remove"`
}

// RunForget calls restic to run the forget command according to the given policy
func RunForget(repository string, password string, env *map[string]string, policy *ForgetPolicy) (*ForgetResult, error) {
	ctx := context.TODO()

	args := []string{
		"-r", repository,
		"forget",
		"--prune",
		"--json",
		"-q",
	}
	policyArgs := getForgetPolicyArgs(policy)
	args = append(args, policyArgs...)

	cmd := execCommandContext(ctx, "restic", args...)
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, buildCmdEnv(password, env)...)

	log.WithFields(log.Fields{"component": "restic", "cmd": strings.Join(cmd.Args, " ")}).Debug("Running restic forget command")
	output, err := cmd.Output()

	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("restic command returned with code %d : %s", exiterr.ExitCode(), exiterr.Stderr)
		}

		return nil, err
	}

	groupResults := []ForgetGroupResult{}
	if err := json.Unmarshal(output, &groupResults); err != nil {
		return nil, err
	}

	result := ForgetResult{GroupResults: groupResults}
	return &result, nil
}

func getForgetPolicyArgs(policy *ForgetPolicy) []string {
	var args []string

	if policy.KeepLast != 0 {
		args = append(args, "--keep-last="+strconv.Itoa(policy.KeepLast))
	}

	if policy.KeepDaily != 0 {
		args = append(args, "--keep-daily="+strconv.Itoa(policy.KeepDaily))
	}

	if policy.KeepHourly != 0 {
		args = append(args, "--keep-hourly="+strconv.Itoa(policy.KeepHourly))
	}

	if policy.KeepWeekly != 0 {
		args = append(args, "--keep-weekly="+strconv.Itoa(policy.KeepWeekly))
	}

	if policy.KeepMonthly != 0 {
		args = append(args, "--keep-monthly="+strconv.Itoa(policy.KeepMonthly))
	}

	if policy.KeepYearly != 0 {
		args = append(args, "--keep-yearly="+strconv.Itoa(policy.KeepYearly))
	}

	if len(policy.KeepTags) > 0 {
		for _, v := range policy.KeepTags {
			args = append(args, "--keep-tags="+v)
		}
	}

	if len(policy.KeepWithin) > 0 {
		args = append(args, "--keep-within="+policy.KeepWithin)
	}

	return args
}
