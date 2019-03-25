package exporter

import (
	"context"
	"fmt"

	"github.com/vlamug/accesslog-exporter/cache"
	"github.com/vlamug/accesslog-exporter/config"
	"github.com/vlamug/accesslog-exporter/exposer"
	"github.com/vlamug/accesslog-exporter/input"
	"github.com/vlamug/accesslog-exporter/parser"
)

type Exporter struct {
	syslogInput input.Input
	exposeFunc  exposer.Exposer

	workersPool pool
	lines       chan *input.LogLine
}

func NewExporter(
	cfg *config.Config,
	syslogInput input.Input,
	logLinePsr parser.LogLineParser,
	userAgentPsr parser.UserAgentParser,
	cc cache.Cache,
	exposeFunc exposer.Exposer,
) (*Exporter, error) {

	if len(cfg.Sources) == 0 {
		return nil, fmt.Errorf("nothing to parse and export, specify at least one source\n")
	}

	// init workers
	workersPool := make(pool, cfg.Global.ExportWorkers)
	for i := 0; i < cfg.Global.ExportWorkers; i++ {
		workersPool <- NewExportWorker(
			logLinePsr,
			userAgentPsr,
			cc,
			exposeFunc,
			&cfg.Global.InternalSubnets,
			&cfg.Sources,
			&cfg.Global.UserAgentReplacementSettings,
			&cfg.Global.RequestURIReplacementSettings,
			&cfg.Global.Hosts,
		)
	}

	return &Exporter{
		syslogInput: syslogInput,
		workersPool: workersPool,
		exposeFunc:  exposeFunc,
		lines:       make(chan *input.LogLine),
	}, nil
}

// Run runs exporting metrics
func (s *Exporter) Run(ctx context.Context) {
	// run syslog server
	go s.syslogInput.Run(s.lines, ctx)

	go func() {
		for line := range s.lines {
			select {
			case <-ctx.Done():
				return
			case w := <-s.workersPool: // tries to get worker from pool
				go func() {
					// runs worker
					w.Process(line, ctx)

					// returns worker to pool
					s.workersPool <- w
				}()
			default:
				s.exposeFunc(exposer.LogsDroppedTotalName, []string{line.NginxHost}, float64(0))
			}

			s.exposeFunc(exposer.LogsTotal, []string{line.NginxHost}, float64(0))
		}
	}()
}
