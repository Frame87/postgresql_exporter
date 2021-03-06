package gauges

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var databaseWritingUsageQuery = `
	SELECT coalesce(tup_inserted, 0) as tup_inserted
		 , coalesce(tup_updated, 0) as tup_updated
		 , coalesce(tup_deleted, 0) as tup_deleted
	  FROM pg_stat_database 
	 WHERE datname = current_database()
`

type writingUsage struct {
	TuplesInserted float64 `db:"tup_inserted"`
	TuplesUpdated  float64 `db:"tup_updated"`
	TuplesDeleted  float64 `db:"tup_deleted"`
}

func (g *Gauges) DatabaseWritingUsage() *prometheus.GaugeVec {
	var gauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "postgresql_database_writing_usage",
			Help:        "Number of inserted, updated and deleted rows per database",
			ConstLabels: g.labels,
		},
		[]string{"stat"},
	)
	go func() {
		for {
			var writingUsage []writingUsage
			if err := g.query(databaseWritingUsageQuery, &writingUsage, emptyParams); err == nil {
				for _, database := range writingUsage {
					gauge.With(prometheus.Labels{
						"stat": "tup_inserted",
					}).Set(database.TuplesInserted)
					gauge.With(prometheus.Labels{
						"stat": "tup_updated",
					}).Set(database.TuplesUpdated)
					gauge.With(prometheus.Labels{
						"stat": "tup_deleted",
					}).Set(database.TuplesDeleted)
				}
			}
			time.Sleep(g.interval)
		}
	}()
	return gauge
}
