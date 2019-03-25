package input

import (
	"context"

	"github.com/vlamug/accesslog-exporter/pkg/logging"

	"gopkg.in/mcuadros/go-syslog.v2"
)

// Input is an interface for input that received log lines
type Input interface {
	Run(lines chan<- *LogLine, ctx context.Context)
}

// Syslog is input that works as syslog server.
type Syslog struct {
	listenAddr string
}

// NewSyslog creates new syslog input.
func NewSyslog(listenAddr string) *Syslog {
	return &Syslog{listenAddr: listenAddr}
}

// Run runs syslog server ans sending log lines to 'lines' channel.
func (i *Syslog) Run(lines chan<- *LogLine, ctx context.Context) {
	syslogChannel := make(syslog.LogPartsChannel)
	logHandler := syslog.NewChannelHandler(syslogChannel)

	server := syslog.NewServer()

	// listen on upd protocol
	err := server.ListenUDP(i.listenAddr)
	if err != nil {
		logging.WithContext(ctx).With().Sugar().Fatalf("could not listen syslog server on udp protocol: %s", err)
	}

	server.SetHandler(logHandler)
	server.SetFormat(syslog.Automatic)

	err = server.Boot()
	if err != nil {
		logging.WithContext(ctx).Sugar().Fatalf("could nod boot syslog server: %s", err)
	}

	// send logs to channel
	go func(logsChannel syslog.LogPartsChannel) {
		for logLine := range logsChannel {
			lines <- NewLogLine(logLine["hostname"].(string), logLine["content"].(string))
		}
	}(syslogChannel)

	server.Wait()
}
