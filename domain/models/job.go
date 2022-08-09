package models

import "net/url"

// CrawlJob represents an url to work on.
type CrawlJob struct {
	SeedURL   *url.URL // url that generated this job.
	URL       *url.URL
	Depth     uint64
	Completed bool // zero value is false
	Result    CrawlResult
}

type CrawlResult struct {
	NewJobs []CrawlJob // A crawl job will produce new jobs.
}
