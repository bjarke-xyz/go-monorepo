package main

type AppContext struct {
	PriceRepository *PriceRepository
	Storage         *StorageClient
	JobManager      *JobManager
}

func NewAppContext(config *Config) *AppContext {
	return &AppContext{
		PriceRepository: NewPriceRepository(config),
		Storage:         NewStorageClient(config),
		JobManager:      NewJobManager(),
	}
}
