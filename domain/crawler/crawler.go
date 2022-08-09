//go:generate moq -out internal/mocks/fetcher_extractor_moq.go -pkg mocks . FetcherExtractor
//go:generate moq -out internal/mocks/queue_moq.go -pkg mocks . Queue

package crawler

import (
	"context"
	"io"
	"monzoCrawler/domain/models"
	"net/url"
)

type (
	// FetcherExtractor focuses on retrieving and extracting from a given url.
	// Fetching may differ e.g. HTTP, Reading from a file, etc.
	// Response formats may differ e.g. csv so its responsibility is to also extract based on the expected format.
	FetcherExtractor interface {
		Fetch(ctx context.Context, url *url.URL) (io.ReadCloser, error)
		Extract(url *url.URL, contents io.Reader) (models.CrawlResult, error)
	}

	// Queue of jobs to push to.
	Queue interface {
		Push(val interface{}) error
	}
)

type Crawler struct {
	job          models.CrawlJob // job to crawl
	jobPushQueue Queue           // queue to send generated jobs to

	result models.CrawlResult // result of the crawl

	fetcherExtractor FetcherExtractor

	done chan<- models.CrawlJob // channel to signal that the crawler is done
}

func New(fetcherExtractor FetcherExtractor, job models.CrawlJob, jobPushQueue Queue, done chan<- models.CrawlJob) *Crawler {
	return &Crawler{
		job:              job,
		jobPushQueue:     jobPushQueue,
		result:           models.CrawlResult{},
		fetcherExtractor: fetcherExtractor,
		done:             done,
	}
}

// Crawl crawls on jobs picked up.
func (c *Crawler) Crawl(ctx context.Context) {
	defer c.signalDone()

	// TODO: Handle error
	// error retryable?
	response, err := c.fetcherExtractor.Fetch(ctx, c.job.URL)
	if err != nil {
		return
	}
	defer response.Close()

	// TODO: Handle error
	// error retryable?
	results, _ := c.fetcherExtractor.Extract(c.job.URL, response)

	for _, job := range results.NewJobs {
		// TODO: Handle error
		// increment depth
		job.Depth = c.job.Depth + 1
		job.SeedURL = c.job.SeedURL
		_ = c.jobPushQueue.Push(job)
	}
	c.job.Completed = true
	c.job.Result = results
}

func (c *Crawler) signalDone() {
	c.done <- c.job
}
