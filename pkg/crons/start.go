package crons

import (
	"github.com/efucloud/cloud-claw-manager/pkg/config"
	"github.com/robfig/cron/v3"
)

func StartCronJob() {
	config.Logger.Info("start cron job")
	c := cron.New()
	c.Start()

}
