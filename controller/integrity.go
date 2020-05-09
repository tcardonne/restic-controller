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

// IntegrityReport represents an integrity report
type IntegrityReport struct {
	Time    *time.Time
	Healthy bool
}

// IntegrityController represents an instance of the integrity controller
type IntegrityController struct {
	mu               sync.RWMutex
	logger           *log.Entry
	repositories     []*conf.Repository
	integrityReports map[string]IntegrityReport
}

// NewIntegrityController creates a new IntegrityController
func NewIntegrityController(repositories []*conf.Repository) *IntegrityController {
	reports := make(map[string]IntegrityReport)

	for _, repo := range repositories {
		reports[repo.Name] = IntegrityReport{}
	}

	return &IntegrityController{
		logger:           log.WithFields(log.Fields{"component": "controller/integrity"}),
		repositories:     repositories,
		integrityReports: reports,
	}
}

// Start runs integrity checks in the background
func (c *IntegrityController) Start() error {
	schedules := cron.New()
	for _, repository := range c.repositories {
		if repository.Check.Schedule == "" {
			continue
		}

		_, err := schedules.AddFunc(repository.Check.Schedule, c.RunCheck(repository))
		if err != nil {
			return fmt.Errorf(`Failed to add cron for repository "%s" with schedule "%s" : "%s"`, repository.Name, repository.Check.Schedule, err)
		}

		if repository.Check.RunOnStartup {
			go c.RunCheck(repository)()
		}
	}
	schedules.Start()

	return nil
}

// RunCheck runs an integrity check
func (c *IntegrityController) RunCheck(repository *conf.Repository) func() {
	return func() {
		c.logger.WithField("repository", repository.Name).Info("Running integrity check")

		healthy, err := restic.RunIntegrityCheck(repository.URL, repository.Password)
		c.setIntegrityReport(repository.Name, healthy)
		if err != nil {
			c.logger.WithFields(log.Fields{
				"repository": repository.Name,
				"healthy":    healthy,
				"err":        err,
			}).Error("Integrity check reported unhealthy")
		} else {
			c.logger.WithFields(log.Fields{
				"repository": repository.Name,
				"healthy":    healthy,
			}).Info("Finished integrity check")
		}
	}
}

// GetIntegrityReport returns an integrity report
func (c *IntegrityController) GetIntegrityReport(repositoryName string) (IntegrityReport, error) {
	c.mu.Lock()
	report, ok := c.integrityReports[repositoryName]
	c.mu.Unlock()
	if !ok {
		return IntegrityReport{}, fmt.Errorf("No repository for name %s", repositoryName)
	}

	return report, nil
}

func (c *IntegrityController) setIntegrityReport(repositoryName string, healthy bool) error {
	c.mu.Lock()
	report, ok := c.integrityReports[repositoryName]
	if !ok {
		c.mu.Unlock()
		return fmt.Errorf("No repository for name %s", repositoryName)
	}
	time := time.Now()
	report.Time = &time
	report.Healthy = healthy
	c.integrityReports[repositoryName] = report
	c.mu.Unlock()

	return nil
}
