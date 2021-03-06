package gauges

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Metric struct {
	Value float64
	Name  string
}

var labels = prometheus.Labels{
	"testing": "true",
}

func evaluate(t *testing.T, c prometheus.Collector) (result []Metric) {
	var require = require.New(t)
	var reg = prometheus.NewRegistry()
	require.NoError(reg.Register(c))
	time.Sleep(100 * time.Millisecond)
	metrics, err := reg.Gather()
	require.NoError(err)
	for _, metric := range metrics {
		for _, m := range metric.GetMetric() {
			result = append(
				result,
				Metric{
					Value: m.GetGauge().GetValue(),
					Name:  metric.GetName(),
				},
			)
		}
	}
	return
}

func assertNoErrs(t *testing.T, gauges *Gauges) {
	var assert = assert.New(t)
	var errs = evaluate(t, gauges.Errs)
	assert.Len(errs, 1)
	assert.Equal(0.0, errs[0].Value)
}

func assertErrs(t *testing.T, gauges *Gauges, errors int) {
	var assert = assert.New(t)
	var errs = evaluate(t, gauges.Errs)
	assert.Len(errs, 1)
	assert.Equal(float64(errors), errs[0].Value)
}

func assertGreaterThan(t *testing.T, expected float64, m Metric) {
	var assert = assert.New(t)
	assert.True(
		m.Value > expected,
		"%s should be > %v: %v", m.Name, expected, m.Value,
	)
}

func assertEqual(t *testing.T, expected float64, m Metric) {
	var assert = assert.New(t)
	assert.Equal(
		expected,
		m.Value,
		"%s should be equal to %v: %v", m.Name, expected, m.Value,
	)
}

func prepare(t *testing.T) (*sql.DB, *Gauges, func()) {
	var db = connect(t)
	var gauges = New("test", db, 1*time.Minute, 1*time.Second)
	return db, gauges, func() {
		assert.NoError(t, db.Close())
	}
}

func connect(t *testing.T) *sql.DB {
	var require = require.New(t)
	var url = os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://localhost:5432/postgres?sslmode=disable"
	}
	db, err := sql.Open("postgres", url)
	require.NoError(err, "failed to open connection to the database")
	require.NoError(db.Ping(), "failed to ping database")
	db.SetMaxOpenConns(1)
	return db
}

func createTestTable(t *testing.T, db *sql.DB) func() {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS testtable(id bigint PRIMARY KEY)")
	require.NoError(t, err)
	return func() {
		_, err := db.Exec("DROP TABLE IF EXISTS testtable")
		assert.New(t).NoError(err)
	}
}

func TestVersion(t *testing.T) {
	var assert = assert.New(t)
	_, gauges, close := prepare(t)
	defer close()
	assert.NotEmpty(gauges.version())
	assertNoErrs(t, gauges)
}
