package main

type AppContext struct {
	PriceRepository *PriceRepository
	Storage         *StorageClient
}

func NewAppContext(config *Config) *AppContext {
	return &AppContext{
		PriceRepository: NewPriceRepository(config),
		Storage:         NewStorageClient(config),
	}
}
