package gauges

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeadTuples(t *testing.T) {
	var assert = assert.New(t)
	db, gauges, close := prepare(t)
	defer close()

	_, err := db.Exec("CREATE TABLE IF NOT EXISTS testtable(id bigint)")
	assert.NoError(err)
	_, err = db.Exec("CREATE EXTENSION IF NOT EXISTS pgstattuple")
	assert.NoError(err)
	defer func() {
		_, err := db.Exec("DROP TABLE IF EXISTS testtable")
		assert.NoError(err)
	}()

	var metrics = evaluate(t, gauges.DeadTuples())
	assert.True(len(metrics) > 0)
	for _, m := range metrics {
		assert.Equal(0.0, m.Value)
	}
	assertNoErrs(t, gauges)
}

func TestDeadTuplesWithoutPgstatTuple(t *testing.T) {
	var assert = assert.New(t)
	db, gauges, close := prepare(t)
	defer close()

	_, err := db.Exec("CREATE TABLE IF NOT EXISTS testtable(id bigint)")
	assert.NoError(err)
	_, err = db.Exec("DROP EXTENSION IF EXISTS pgstattuple")
	assert.NoError(err)
	defer func() {
		_, err := db.Exec("DROP TABLE IF EXISTS testtable")
		assert.NoError(err)
	}()

	var metrics = evaluate(t, gauges.DeadTuples())
	assert.Len(metrics, 0)
	assertErrs(t, gauges, 0)
}
