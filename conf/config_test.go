package conf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfiguration_Base(t *testing.T) {
	config, err := LoadConfiguration("testdata/config_base.yml")

	assert.NoError(t, err)

	assert.Equal(t, ":8080", config.Exporter.BindAddress)
	assert.Equal(t, "info", config.Log.Level)
	assert.Equal(t, "backtothefuture", config.Repositories[0].Name)
	assert.Equal(t, "rest:https://user:password@repositories.restic.example/backtothefuture", config.Repositories[0].URL)
	assert.Equal(t, "testtest", config.Repositories[0].Password)
	assert.Equal(t, "* * * * *", config.Repositories[0].Check.Schedule)
	assert.Equal(t, "* * * * *", config.Repositories[0].Retention.Schedule)
	assert.Equal(t, 1, config.Repositories[0].Retention.Policy.KeepLast)
}

func TestLoadConfiguration_FromFile(t *testing.T) {
	filename := "./tmp-test-loadconfig-envfromfile"
	err := os.WriteFile(filename, []byte("someSecretValue"), 0644)
	defer os.Remove(filename)

	if err != nil {
		t.Logf("Failed to write tmp file: %s", err)
		t.FailNow()
	}
	config, err := LoadConfiguration("testdata/config_from_file.yml")

	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, ":8080", config.Exporter.BindAddress)
	assert.Equal(t, "info", config.Log.Level)
	assert.Equal(t, "backtothefuture", config.Repositories[0].Name)
	assert.Equal(t, "rest:https://repositories.restic.example/backtothefuture", config.Repositories[0].URL)
	assert.Equal(t, "someSecretValue", config.Repositories[0].Env["RESTIC_REST_PASSWORD"])
	assert.Equal(t, "someSecretValue", config.Repositories[0].Password)
	assert.Equal(t, "* * * * *", config.Repositories[0].Check.Schedule)
	assert.Equal(t, "* * * * *", config.Repositories[0].Retention.Schedule)
	assert.Equal(t, 1, config.Repositories[0].Retention.Policy.KeepLast)
}

func TestLoadConfiguration_Invalid(t *testing.T) {
	_, err := LoadConfiguration("testdata/config_invalid.yml")

	assert.Error(t, err)
}
