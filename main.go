package main

import (
	"flag"
	"os"

	"github.com/opensourceways/community-robot-lib/logrusutil"
	liboptions "github.com/opensourceways/community-robot-lib/options"
	"github.com/opensourceways/xihe-inference-evaluate/client"
	"github.com/opensourceways/xihe-inference-evaluate/config"
	"github.com/opensourceways/xihe-inference-evaluate/controller"
	"github.com/opensourceways/xihe-inference-evaluate/infrastructure/watchimpl"
	"github.com/opensourceways/xihe-inference-evaluate/server"
	"github.com/sirupsen/logrus"
)

type options struct {
	service     liboptions.ServiceOptions
	enableDebug bool
}

func (o *options) Validate() error {
	return o.service.Validate()
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options

	o.service.AddFlags(fs)

	fs.BoolVar(
		&o.enableDebug, "enable_debug", false,
		"whether to enable debug model.",
	)

	fs.Parse(args)
	return o
}

func main() {
	logrusutil.ComponentInit("xihe")
	log := logrus.NewEntry(logrus.StandardLogger())

	err := client.Init()
	if err != nil {
		logrus.Fatalf("k8s client init, err:%s", err.Error())
	}

	o := gatherOptions(
		flag.NewFlagSet(os.Args[0], flag.ExitOnError),
		os.Args[1:]...,
	)
	if err := o.Validate(); err != nil {
		logrus.Fatalf("Invalid options, err:%s", err.Error())
	}

	if o.enableDebug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("debug enabled.")
	}

	// cfg
	cfg := new(config.Config)
	if err := config.LoadConfig(o.service.ConfigFile, cfg); err != nil {
		log.Fatalf("load config, err:%s", err.Error())
	}

	// controller
	controller.Init(log)

	// watcher
	w := watchimpl.NewWatcher(&cfg.Watch)
	go w.Run()
	defer w.Exit()

	// run
	server.StartWebServer(o.service.Port, o.service.GracePeriod, cfg)
}
