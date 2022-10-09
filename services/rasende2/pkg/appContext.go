package pkg

import (
	"github.com/bjarke-xyz/go-monorepo/libs/common/config"
	"github.com/bjarke-xyz/go-monorepo/libs/common/jobs"
)

type AppContext struct {
	Cache      *Cache
	Config     *config.Config
	JobManager jobs.JobManager
}
