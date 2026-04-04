package crons

import (
	"context"
	"github.com/efucloud/cloud-claw-manager/pkg/config"
	"github.com/efucloud/cloud-claw-manager/pkg/openclaw"
	"github.com/robfig/cron/v3"
	"time"
)

func StartCronJob(ctx context.Context) {
	config.Logger.Info("start cron job")
	c := cron.New()
	if _, err := c.AddFunc("@every 45s", func() {
		runCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := openclaw.RefreshRuntimeInsightsCache(runCtx); err != nil {
			config.Logger.Warnf("refresh runtime insights cache failed: %v", err)
		}
	}); err != nil {
		config.Logger.Warnf("add runtime insights cron failed: %v", err)
	}
	c.Start()
	go func() {
		runCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := openclaw.RefreshRuntimeInsightsCache(runCtx); err != nil {
			config.Logger.Warnf("initial runtime insights refresh failed: %v", err)
		}
	}()
	go func() {
		<-ctx.Done()
		stopCtx := c.Stop()
		<-stopCtx.Done()
	}()

}
