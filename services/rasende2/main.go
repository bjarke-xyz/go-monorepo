package main

import (
	"log"

	"github.com/bjarke-xyz/go-monorepo/libs/common"
	"github.com/bjarke-xyz/go-monorepo/libs/common/config"
	"github.com/bjarke-xyz/go-monorepo/libs/common/db"
	"github.com/bjarke-xyz/go-monorepo/libs/common/jobs"
	"github.com/bjarke-xyz/rasende2/pkg"
	"github.com/bjarke-xyz/rasende2/rss"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg, err := config.NewConfig()
	if err != nil {
		log.Panicf("failed to load config: %v", err)
	}

	err = db.Migrate("up", cfg.ConnectionString())
	if err != nil {
		log.Printf("failed to migrate: %v", err)
	}

	context := &pkg.AppContext{
		Cache:      db.NewRedisCache(cfg),
		Config:     cfg,
		JobManager: *jobs.NewJobManager(),
	}

	rssRepository := rss.NewRssRepository(context)
	rssService := rss.NewRssService(context, rssRepository)

	defer context.JobManager.Stop()
	context.JobManager.Cron("1 * * * *", rss.JobIdentifierIngestion, func() error {
		job := rss.NewIngestionJob(rssService)
		return job.ExecuteJob()
	}, cfg.AppEnv == config.AppEnvProduction)
	go context.JobManager.Start()

	rssHttpHandlers := rss.NewHttpHandlers(context, rssService)

	r := common.GinRouter(cfg)
	r.GET("/search", rssHttpHandlers.HandleSearch)
	r.GET("/charts", rssHttpHandlers.HandleCharts)
	r.POST("/job", rssHttpHandlers.RunJob(cfg.JobKey))

	r.Run()

}
