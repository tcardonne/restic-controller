package restic

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSnapshotGroups(t *testing.T) {
	execCommandContext = mockExecOutputFile("testdata/snapshots.json", 0)
	defer func() { execCommandContext = exec.CommandContext }()

	ctx := context.TODO()
	groups, err := GetSnapshotGroups(ctx, testResticRepository, testResticPassword, &testEnvMap)
	assert.NoError(t, err)

	assert.Len(t, groups, 7)

	// Snapshot groups present in test fixture
	wantGroups := []struct {
		Key                   SnapshotGroupKey
		ExpectedSnapshotCount int
	}{
		{SnapshotGroupKey{Hostname: "host1", Paths: "/", Tags: "tag1"}, 5},
		{SnapshotGroupKey{Hostname: "host2", Paths: "/", Tags: "tag2"}, 7},
		{SnapshotGroupKey{Hostname: "host2", Paths: "/,/backup", Tags: "tag1,tag2"}, 1},
		{SnapshotGroupKey{Hostname: "host1", Paths: "/,/backup", Tags: ""}, 1},
		{SnapshotGroupKey{Hostname: "host1", Paths: "/", Tags: "tag1,tag2"}, 1},
		{SnapshotGroupKey{Hostname: "host1", Paths: "/backup", Tags: "tag1"}, 1},
		{SnapshotGroupKey{Hostname: "host2", Paths: "/", Tags: "tag1"}, 2},
	}

	for _, wg := range wantGroups {
		assert.Containsf(t, groups, wg.Key, `Key "%+v" not found`, wg.Key)
		assert.Lenf(t, groups[wg.Key], wg.ExpectedSnapshotCount, `Group "%+v" should have %d snapshots`, wg.Key, wg.ExpectedSnapshotCount)

		// Test sort. Index 0 should be most recent snapshot
		snapshots := groups[wg.Key]
		for i, sn := range snapshots[1:] {
			prevTime := snapshots[i].Time
			if prevTime.Before(sn.Time) {
				t.Errorf("Snapshot %s with time %s is more recent than previous one with time %s", sn.ShortID, sn.Time, prevTime)
			}
		}
	}
}

func TestGetSnapshotGroups_EmptyRepository(t *testing.T) {
	execCommandContext = mockExecOutputString(`[]`, 0)
	defer func() { execCommandContext = exec.CommandContext }()

	ctx := context.TODO()
	groups, err := GetSnapshotGroups(ctx, testResticRepository, testResticPassword, &testEnvMap)
	assert.NoError(t, err)
	assert.Empty(t, groups)
}

func TestGetSnapshotGroups_InvalidJSON(t *testing.T) {
	execCommandContext = mockExecOutputString(`invalid json`, 0)
	defer func() { execCommandContext = exec.CommandContext }()

	ctx := context.TODO()
	out, err := GetSnapshotGroups(ctx, testResticRepository, testResticPassword, &testEnvMap)
	assert.Error(t, err)
	assert.Nil(t, out)
}

func TestGetSnapshotGroups_ExitError(t *testing.T) {
	execCommandContext = mockExecOutputString("[]", 1)
	defer func() { execCommandContext = exec.CommandContext }()

	ctx := context.TODO()
	out, err := GetSnapshotGroups(ctx, testResticRepository, testResticPassword, &testEnvMap)
	assert.Error(t, err)
	assert.Nil(t, out)
}
