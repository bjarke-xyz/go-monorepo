package pkg

import (
	"github.com/bjarke-xyz/go-monorepo/libs/common/config"
	"github.com/bjarke-xyz/go-monorepo/libs/common/db"
	"github.com/bjarke-xyz/go-monorepo/libs/common/jobs"
)

type AppContext struct {
	Cache      *db.RedisCache
	Config     *config.Config
	JobManager jobs.JobManager
}
