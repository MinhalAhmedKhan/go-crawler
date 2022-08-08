package sameDomainFilter

import (
	"monzoCrawler/domain/model"
)

type Filter struct {
}

func New() *Filter {
	return &Filter{}
}

func (f *Filter) ShouldCrawl(job model.CrawlJob) bool {
	return job.SeedURL.Host == job.URL.Host
}
