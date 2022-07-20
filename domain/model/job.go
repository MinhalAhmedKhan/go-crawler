package model

import "net/url"

// CrawlJob represents an url to work on.
type CrawlJob struct {
	SeedURL *url.URL
	Depth   int
}

type CrawlResult struct {
	NewJobs []CrawlJob // A crawl job will produce new jobs.
}

type JobDAO interface {
	// AddJob adds a job to store.
	AddJob(job CrawlJob) error

	// JobExists checks if a job exists in store.
	JobExists(job CrawlJob) (bool, error)
}
