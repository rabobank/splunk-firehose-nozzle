package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"

	"code.cloudfoundry.org/lager/lagerflags"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/splunknozzle"
)

var (
	version string
	branch  string
	commit  string
	buildos string
)

func main() {
	lagerflags.AddFlags(flag.CommandLine)

	logger, _ := lagerflags.New("splunk-nozzle-logger")
	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true}) // Disable the `time` field, it is already included in the cf logging
	logger.Info("Running splunk-firehose-nozzle")

	shutdownChan := make(chan os.Signal, 2)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	config := splunknozzle.NewConfigFromCmdFlags(version, branch, commit, buildos)
	if config.MemoryBallastSize > 0 {
		ballast := make([]byte, config.MemoryBallastSize<<20)
		_ = ballast
	}

	if config.AppCacheTTL == 0 && config.OrgSpaceCacheTTL > 0 {
		logger.Info("Apps are not being cached. When apps are not cached, the org and space caching TTL is ineffective")
	}

	splunkNozzle := splunknozzle.NewSplunkFirehoseNozzle(config, logger)
	err := splunkNozzle.Run(shutdownChan)
	if err != nil {
		logger.Error("Failed to run splunk-firehose-nozzle", err)
	}
}
