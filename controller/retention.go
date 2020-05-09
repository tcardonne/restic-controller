package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/tcardonne/restic-controller/conf"
	"github.com/tcardonne/restic-controller/restic"
)

// RetentionReport represents a report about a forget action
type RetentionReport struct {
	Time    *time.Time
	Kept    int
	Removed int
}

// RetentionController represents an instance of the retention controller
type RetentionController struct {
	mu               sync.RWMutex
	logger           *log.Entry
	repositories     []*conf.Repository
	retentionReports map[string]RetentionReport
}

// NewRetentionController creates a new retention controller
func NewRetentionController(repositories []*conf.Repository) *RetentionController {
	reports := make(map[string]RetentionReport)

	for _, repo := range repositories {
		reports[repo.Name] = RetentionReport{}
	}

	return &RetentionController{
		logger:           log.WithFields(log.Fields{"component": "controller/retention"}),
		repositories:     repositories,
		retentionReports: reports,
	}
}

// Start applies retention policy periodically checks in the background
func (c *RetentionController) Start() error {
	schedules := cron.New()
	for _, repository := range c.repositories {
		if repository.Retention.Schedule == "" {
			continue
		}

		_, err := schedules.AddFunc(repository.Retention.Schedule, c.RunForget(repository))
		if err != nil {
			return fmt.Errorf(`Failed to add cron for repository "%s" with schedule "%s" : "%s"`, repository.Name, repository.Retention.Schedule, err)
		}

		if repository.Retention.RunOnStartup {
			go c.RunForget(repository)()
		}
	}
	schedules.Start()

	return nil
}

// RunForget runs a forget action
func (c *RetentionController) RunForget(repository *conf.Repository) func() {
	return func() {
		c.logger.WithField("repository", repository.Name).Info("Running forget")

		forgetResult, err := restic.RunForget(repository.URL, repository.Password, repository.Retention.Policy)
		if err != nil {
			c.logger.WithFields(log.Fields{
				"repository": repository.Name,
				"err":        err,
			}).Error("Forget failed")
		} else {
			c.setRetentionReport(repository.Name, forgetResult.TotalKeep(), forgetResult.TotalRemove())
			c.logger.WithFields(log.Fields{
				"repository": repository.Name,
				"removed":    forgetResult.TotalRemove(),
				"kept":       forgetResult.TotalKeep(),
			}).Info("Finished forget")
		}
	}
}

// GetRetentionReport returns a retention report
func (c *RetentionController) GetRetentionReport(repositoryName string) (RetentionReport, error) {
	c.mu.Lock()
	report, ok := c.retentionReports[repositoryName]
	c.mu.Unlock()
	if !ok {
		return RetentionReport{}, fmt.Errorf("No repository for name %s", repositoryName)
	}

	return report, nil
}

func (c *RetentionController) setRetentionReport(repositoryName string, kept int, removed int) error {
	c.mu.Lock()
	report, ok := c.retentionReports[repositoryName]
	if !ok {
		c.mu.Unlock()
		return fmt.Errorf("No repository for name %s", repositoryName)
	}
	time := time.Now()
	report.Time = &time
	report.Kept = kept
	report.Removed = removed
	c.retentionReports[repositoryName] = report
	c.mu.Unlock()

	return nil
}
