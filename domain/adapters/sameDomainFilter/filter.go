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
	return f.hostWithoutWWW(job.SeedURL.Host) == f.hostWithoutWWW(job.URL.Host)
}

func (f Filter) hostWithoutWWW(host string) string {
	if len(host) > 4 && host[:4] == "www." {
		return host[4:]
	}
	return host
}
