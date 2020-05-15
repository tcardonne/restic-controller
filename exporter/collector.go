package exporter

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/tcardonne/restic-controller/conf"
	"github.com/tcardonne/restic-controller/controller"
	"github.com/tcardonne/restic-controller/restic"
)

type repositoryCollector struct {
	ctx                 context.Context
	repositories        []*conf.Repository
	integrityController *controller.IntegrityController
	retentionController *controller.RetentionController
	logger              *log.Entry

	errorMetric          *prometheus.Desc
	scrapeDurationMetric *prometheus.Desc

	repoSnapshotsTotalMetric           *prometheus.Desc
	repoSnapshotTimestampMetric        *prometheus.Desc
	groupSnapshotsTotalMetric          *prometheus.Desc
	groupLatestSnapshotTimestampMetric *prometheus.Desc

	repoIntegrityStatusMetric       *prometheus.Desc
	repoIntegrityStatusLatestMetric *prometheus.Desc

	repoRetentionForgetKeptMetric    *prometheus.Desc
	repoRetentionForgetRemovedMetric *prometheus.Desc
	repoRetentionForgetLatestMetric  *prometheus.Desc
}

// Initializes every descriptor and returns a pointer to the collector
func newRepositoryCollector(ctx context.Context,
	repositories []*conf.Repository,
	integrityController *controller.IntegrityController,
	retentionController *controller.RetentionController,
) *repositoryCollector {
	return &repositoryCollector{
		ctx:                 ctx,
		repositories:        repositories,
		integrityController: integrityController,
		retentionController: retentionController,
		logger:              log.WithFields(log.Fields{"component": "exporter/collector"}),

		errorMetric: prometheus.NewDesc("restic_error", "Error occurred when trying to collect metrics", nil, nil),
		scrapeDurationMetric: prometheus.NewDesc("restic_scrape_duration_seconds",
			"Total time in seconds spent to collect all metrics",
			nil, nil,
		),

		repoSnapshotsTotalMetric: prometheus.NewDesc("restic_repo_snapshots_total",
			"Total count of snapshots in the repository",
			[]string{"repository"}, nil,
		),
		repoSnapshotTimestampMetric: prometheus.NewDesc("restic_repo_snapshot_datetime_seconds",
			"Number of seconds since 1970 of snapshot's datetime",
			[]string{"repository", "host", "paths", "tags", "short_id"}, nil,
		),
		groupSnapshotsTotalMetric: prometheus.NewDesc("restic_group_snapshots_total",
			"Total count of snapshots in a group",
			[]string{"repository", "host", "paths", "tags"}, nil,
		),
		groupLatestSnapshotTimestampMetric: prometheus.NewDesc("restic_group_snapshot_latest_seconds",
			"Number of seconds since 1970 of last snapshot",
			[]string{"repository", "host", "paths", "tags"}, nil,
		),

		repoIntegrityStatusMetric: prometheus.NewDesc("restic_repo_integrity_status",
			"Status of the repository (healthy or unhealthy)",
			[]string{"repository"}, nil,
		),
		repoIntegrityStatusLatestMetric: prometheus.NewDesc("restic_repo_integrity_status_latest_seconds",
			"Number of seconds since 1970 of last integrity check",
			[]string{"repository"}, nil,
		),

		repoRetentionForgetKeptMetric: prometheus.NewDesc("restic_repo_retention_forget_kept_total",
			"Number of snapshots kept after last forget action",
			[]string{"repository"}, nil,
		),
		repoRetentionForgetRemovedMetric: prometheus.NewDesc("restic_repo_retention_forget_removed_total",
			"Number of snapshots deleted after last forget action",
			[]string{"repository"}, nil,
		),
		repoRetentionForgetLatestMetric: prometheus.NewDesc("restic_repo_retention_forget_latest_seconds",
			"Number of seconds since 1970 of last forget action",
			[]string{"repository"}, nil,
		),
	}
}

//Each and every collector must implement the Describe function.
//It essentially writes all descriptors to the prometheus desc channel.
func (c *repositoryCollector) Describe(ch chan<- *prometheus.Desc) {
	//Update this section with the each metric you create for a given collector
	ch <- c.repoSnapshotsTotalMetric
}

//Collect implements required collect function for all promehteus collectors
func (c *repositoryCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()

	var wg sync.WaitGroup
	for _, repository := range c.repositories {
		wg.Add(1)
		go c.CollectRepository(repository, ch, &wg)
	}
	wg.Wait()

	elapsted := time.Since(start)
	ch <- prometheus.MustNewConstMetric(c.scrapeDurationMetric, prometheus.GaugeValue, elapsted.Seconds())
}

func (c *repositoryCollector) CollectRepository(repository *conf.Repository, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	c.logger.WithField("repository", repository.Name).Info("Starting collection")

	// Snapshot metrics
	groups, err := restic.GetSnapshotGroups(c.ctx, repository.URL, repository.Password)
	if err != nil {
		c.logger.WithFields(log.Fields{"repository": repository.Name, "err": err}).Error("Error occurred when fetching restic snapshot list")
		ch <- prometheus.NewInvalidMetric(c.errorMetric, err)
		wg.Done()
		return
	}

	ch <- prometheus.MustNewConstMetric(c.repoSnapshotsTotalMetric, prometheus.GaugeValue,
		float64(groups.TotalSnapshotsCount()),
		repository.Name,
	)

	for key, snapshots := range groups {
		ch <- prometheus.MustNewConstMetric(c.groupSnapshotsTotalMetric, prometheus.GaugeValue,
			float64(len(snapshots)),
			repository.Name, key.Hostname, key.Paths, key.Tags,
		)

		if len(snapshots) > 0 {
			ch <- prometheus.MustNewConstMetric(c.groupLatestSnapshotTimestampMetric, prometheus.CounterValue,
				float64(snapshots[0].Time.Unix()),
				repository.Name, key.Hostname, key.Paths, key.Tags,
			)
		}

		for _, snapshot := range snapshots {
			ch <- prometheus.MustNewConstMetric(c.repoSnapshotTimestampMetric, prometheus.CounterValue,
				float64(snapshot.Time.Unix()),
				repository.Name, key.Hostname, key.Paths, key.Tags, snapshot.ShortID,
			)
		}
	}

	// Integrity metrics
	report, err := c.integrityController.GetIntegrityReport(repository.Name)
	if err != nil {
		c.logger.WithFields(log.Fields{"repository": repository.Name, "err": err}).Error("Error occurred whe fetching integrity status from controller")
		ch <- prometheus.NewInvalidMetric(c.errorMetric, err)
		wg.Done()
		return
	}
	if report.Time != nil {
		ch <- prometheus.MustNewConstMetric(c.repoIntegrityStatusLatestMetric, prometheus.CounterValue,
			float64(report.Time.Unix()),
			repository.Name,
		)
		var healthy float64
		if report.Healthy {
			healthy = 1.0
		} else {
			healthy = 0.0
		}
		ch <- prometheus.MustNewConstMetric(c.repoIntegrityStatusMetric, prometheus.GaugeValue,
			healthy,
			repository.Name,
		)
	}

	// Retention forget metrics
	retentionReport, err := c.retentionController.GetRetentionReport(repository.Name)
	if err != nil {
		c.logger.WithFields(log.Fields{"repository": repository.Name, "err": err}).Error("Error occurred whe fetching retention report from controller")
		ch <- prometheus.NewInvalidMetric(c.errorMetric, err)
		wg.Done()
		return
	}
	if retentionReport.Time != nil {
		ch <- prometheus.MustNewConstMetric(c.repoRetentionForgetLatestMetric, prometheus.CounterValue,
			float64(retentionReport.Time.Unix()),
			repository.Name,
		)
		ch <- prometheus.MustNewConstMetric(c.repoRetentionForgetKeptMetric, prometheus.GaugeValue,
			float64(retentionReport.Kept),
			repository.Name,
		)
		ch <- prometheus.MustNewConstMetric(c.repoRetentionForgetRemovedMetric, prometheus.GaugeValue,
			float64(retentionReport.Removed),
			repository.Name,
		)
	}

	c.logger.WithField("repository", repository.Name).Info("Finished collection")
	wg.Done()
}
