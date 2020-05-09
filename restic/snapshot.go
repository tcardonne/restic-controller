package restic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Snapshot represents a snapshot returned by restic
type Snapshot struct {
	ID       string    `json:"id"`
	ShortID  string    `json:"short_id"`
	Hostname string    `json:"hostname"`
	Paths    []string  `json:"paths"`
	Tags     []string  `json:"tags"`
	Time     time.Time `json:"time"`
	Username string    `json:"username"`
}

// Snapshots is a list of Snapshots
type Snapshots []*Snapshot

// Len returns the number of snapshots in sn.
func (sn Snapshots) Len() int {
	return len(sn)
}

// Less returns true iff the ith snapshot has been made after the jth.
func (sn Snapshots) Less(i, j int) bool {
	return sn[i].Time.After(sn[j].Time)
}

// Swap exchanges the two snapshots.
func (sn Snapshots) Swap(i, j int) {
	sn[i], sn[j] = sn[j], sn[i]
}

// SnapshotGroupKey is used as a key when grouping snapshots
type SnapshotGroupKey struct {
	Hostname string `json:"hostname"`
	Paths    string `json:"paths"`
	Tags     string `json:"tags"`
}

// SnapshotGroups is used a as a map of Snapshots grouped by SnapshotGroupKey
type SnapshotGroups map[SnapshotGroupKey]Snapshots

// TotalSnapshotsCount returns the total count of snapshots across all groups
func (sg SnapshotGroups) TotalSnapshotsCount() int {
	var count int
	for _, g := range sg {
		count += len(g)
	}
	return count
}

// Sort will order snapshots in each group by date, newest to oldest.
// Latest snapshots will then be on index 0.
func (sg SnapshotGroups) Sort() {
	for _, g := range sg {
		sort.Sort(g)
	}
}

// GetSnapshotGroups returns the list of snapshots
func GetSnapshotGroups(ctx context.Context, repository string, password string) (SnapshotGroups, error) {
	cmd := execCommandContext(ctx, "restic", "-r", repository, "snapshots", "--json", "--no-lock")
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, "RESTIC_PASSWORD="+password)

	log.WithFields(log.Fields{"component": "restic", "cmd": strings.Join(cmd.Args, " ")}).Debug("Running restic snapshots command")
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if exiterr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("Restic command returned with code %d : %s", exiterr.ExitCode(), exiterr.Stderr)
		}

		return nil, err
	}

	var snapshots Snapshots
	if err := json.Unmarshal(output, &snapshots); err != nil {
		return nil, err
	}

	snapshotGroups := groupSnapshots(snapshots)
	snapshotGroups.Sort()

	return snapshotGroups, nil
}

func groupSnapshots(snapshots Snapshots) SnapshotGroups {
	// group by hostname and dirs
	snapshotGroups := make(SnapshotGroups)

	for _, sn := range snapshots {
		hostname := sn.Hostname
		paths := sn.Paths
		sort.StringSlice(paths).Sort()
		tags := sn.Tags
		sort.StringSlice(tags).Sort()

		groupKey := SnapshotGroupKey{
			Hostname: hostname,
			Paths:    strings.Join(paths, ","),
			Tags:     strings.Join(tags, ","),
		}
		snapshotGroups[groupKey] = append(snapshotGroups[groupKey], sn)
	}

	return snapshotGroups
}
