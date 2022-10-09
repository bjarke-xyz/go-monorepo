package rss

const JobIdentifierIngestion = "RASENDE2_INGESTION_JOB"

type IngestionJob struct {
	service *RssService
}

func NewIngestionJob(service *RssService) *IngestionJob {
	return &IngestionJob{
		service: service,
	}
}

func (i *IngestionJob) ExecuteJob() error {
	return i.service.FetchAndSaveNewItems()
}
