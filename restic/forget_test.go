package restic

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunForget(t *testing.T) {
	execCommandContext = mockExecOutputFile("testdata/forget_output.json", 0)
	defer func() { execCommandContext = exec.CommandContext }()

	policy := ForgetPolicy{KeepLast: 1}
	result, err := RunForget(testResticRepository, testResticPassword, &testEnvMap, &policy)
	assert.NoError(t, err)

	assert.Equal(t, 1, result.TotalKeep())
	assert.Equal(t, 3, result.TotalRemove())
}

func TestRunForget_InvalidJSON(t *testing.T) {
	execCommandContext = mockExecOutputString("invalidjson", 0)
	defer func() { execCommandContext = exec.CommandContext }()

	policy := ForgetPolicy{KeepLast: 1}
	result, err := RunForget(testResticRepository, testResticPassword, &testEnvMap, &policy)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestRunForget_ExitError(t *testing.T) {
	execCommandContext = mockExecOutputString("", 1)
	defer func() { execCommandContext = exec.CommandContext }()

	policy := ForgetPolicy{KeepLast: 1}
	result, err := RunForget(testResticRepository, testResticPassword, &testEnvMap, &policy)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetForgetPolicyArgs(t *testing.T) {
	testCases := []struct {
		policy ForgetPolicy
		want   []string
	}{
		// Simple
		{ForgetPolicy{KeepLast: 1}, []string{"--keep-last=1"}},
		{ForgetPolicy{KeepDaily: 1}, []string{"--keep-daily=1"}},
		{ForgetPolicy{KeepHourly: 1}, []string{"--keep-hourly=1"}},
		{ForgetPolicy{KeepWeekly: 1}, []string{"--keep-weekly=1"}},
		{ForgetPolicy{KeepMonthly: 1}, []string{"--keep-monthly=1"}},
		{ForgetPolicy{KeepYearly: 1}, []string{"--keep-yearly=1"}},
		{ForgetPolicy{KeepTags: []string{"tag1,tag2"}}, []string{"--keep-tags=tag1,tag2"}},
		{ForgetPolicy{KeepTags: []string{"tag1", "tag2"}}, []string{"--keep-tags=tag1", "--keep-tags=tag2"}},
		{ForgetPolicy{KeepWithin: "2y5m7d3h"}, []string{"--keep-within=2y5m7d3h"}},
		// Complex
		{ForgetPolicy{KeepLast: 0, KeepDaily: 1}, []string{"--keep-daily=1"}},
		{ForgetPolicy{KeepDaily: 10, KeepLast: 1}, []string{"--keep-last=1", "--keep-daily=10"}},
	}

	for _, tc := range testCases {
		t.Run(strings.Join(tc.want, " "), func(t *testing.T) {
			result := getForgetPolicyArgs(&tc.policy)

			assert.Equal(t, tc.want, result)
		})
	}
}
