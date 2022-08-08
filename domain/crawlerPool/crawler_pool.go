//go:generate moq -out internal/mocks/queue_moq.go -pkg mocks . Queue
//go:generate moq -out internal/mocks/fetcherextractor_moq.go -pkg mocks . FetcherExtractor

package crawlerPool

import (
	"context"
	"io"

	"net/url"
	"sync/atomic"
	"time"

	"monzoCrawler/domain/crawler"
	"monzoCrawler/domain/model"
)

type (
	Logger interface {
		Printf(format string, args ...interface{})
	}

	Queue interface {
		Push(val interface{}) error
		Pop() (interface{}, error)
	}

	FetcherExtractor interface {
		Fetch(ctx context.Context, url url.URL) (io.ReadCloser, error)
		Extract(io.Reader) (model.CrawlResult, error)
	}

	// JobFilter is a function that returns false if the job should be filtered out.
	JobFilter interface {
		ShouldCrawl(job model.CrawlJob) bool
	}
)

// CrawlerPool manages crawlers running.
type CrawlerPool struct {
	logger Logger

	size         uint64 // Number of crawlers
	maxDepth     uint64 // Max depth to crawl
	depthReached uint64 // Depth reached by this pool.
	jobQueue     Queue  // Jobs to be processed.

	fetcherExtractor FetcherExtractor

	jobFilters []JobFilter // Filters to apply to jobs.

	activeCrawlers      uint64              // Number of active crawlers.
	updatedCrawlerCount chan struct{}       // Channel to signal that crawler count has been updated.
	crawlerDone         chan model.CrawlJob // Channel to signal that a crawler is done with this job.

	completionHook CrawlerCompletedHook // Hook to call when a crawler is done.

	shutdownTimeout time.Duration    // Timeout for shutdown.
	forceShutdown   chan interface{} // Channel to signal that the pool should forcefully shut down.

}

// CrawlerCompletedHook is a function that is called when a crawler is done with a job.
type CrawlerCompletedHook func(context.Context, model.CrawlJob)

func NoOpCompletedHook(ctx context.Context, model model.CrawlJob) {
	return
}

// New creates a new CrawlerPool.
// Filters are applied in the order they are specified.
// CompletedHook is called when a crawler is done with a job and is a non-blocking call.
func New(logger Logger, size uint64, jobQueue Queue, shutdownTimeout time.Duration, fetcherExtractor FetcherExtractor, maxDepth uint64, jobFilters []JobFilter, completionHook CrawlerCompletedHook) *CrawlerPool {
	// TODO: Add validation for fields
	return &CrawlerPool{
		logger: logger,

		size:         size,
		maxDepth:     maxDepth,
		depthReached: 0,
		jobQueue:     jobQueue,

		fetcherExtractor: fetcherExtractor,

		jobFilters: jobFilters,

		activeCrawlers:      0,
		updatedCrawlerCount: make(chan struct{}),
		crawlerDone:         make(chan model.CrawlJob),

		completionHook: completionHook,

		shutdownTimeout: shutdownTimeout,
		forceShutdown:   make(chan interface{}, 1),
	}
}

func (cp *CrawlerPool) listenForCompletedJobs(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			go func() {
				cp.forceShutdown <- <-time.After(cp.shutdownTimeout)
			}()
			// set new ctx done channel to prevent spawning shutdown go routines.
			ctx = context.Background()
		case job := <-cp.crawlerDone:
			// A crawler completed its job.
			cp.decrementCrawlerCount()

			if job.Completed {
				go cp.completionHook(ctx, job)
			}
		}
	}
}

// Wait blocks until all crawlers have exited or shutdown timed out forcing a shutdown.
func (cp *CrawlerPool) wait(ctx context.Context, doneChan chan struct{}) {
	cancelled := false
	// signal crawler pool exited gracefully
	defer func() {
		doneChan <- struct{}{}
	}()

	for {
		select {
		case <-ctx.Done():
			cancelled = true
			if atomic.LoadUint64(&cp.activeCrawlers) == 0 {
				return
			}
			ctx = context.Background()
		case <-cp.forceShutdown:
			cp.logger.Printf("Shutdown forced")
			return
		case <-cp.updatedCrawlerCount:
			if cancelled && atomic.LoadUint64(&cp.activeCrawlers) == 0 {
				cp.logger.Printf("all crawlers successfully shutdown")
				return
			}
			// TODO: what if the depth is < maxdepth? and there is no more urls?
			if atomic.LoadUint64(&cp.activeCrawlers) == 0 && cp.getDepthCount() > cp.maxDepth {
				cp.logger.Printf("max depth of %d crawled, shutting down", cp.maxDepth)
				return
			}
		}
	}
}

// Start crawling up to the specified depth.
func (cp *CrawlerPool) Start(ctx context.Context, doneChan chan struct{}) {
	go cp.listenForCompletedJobs(ctx)
	go cp.wait(ctx, doneChan)

	for {
		select {
		case <-ctx.Done():
			// graceful shutdown
			return
		default:
			if atomic.LoadUint64(&cp.activeCrawlers) == cp.size {
				continue
			}

			jobPickedUp, err := cp.jobQueue.Pop()

			if err != nil {
				//TODO: log error, continue
				continue
			}

			if jobPickedUp == nil {
				continue
			}

			job, ok := jobPickedUp.(model.CrawlJob)

			if !ok {
				cp.logger.Printf("Invalid job type found, got job of type %T", jobPickedUp)
				continue
			}

			if cp.getDepthCount() > cp.maxDepth {
				// picked up first job that is too deep.
				// given a breadth first search, the first job that is too deep is the start of the new depth.
				// therefore, crawl is completed for the specified depth.
				return
			}

			if job.Depth > cp.getDepthCount() {
				cp.incrementDepthCount()
			}

			// run filters on job
			for _, filter := range cp.jobFilters {
				if !filter.ShouldCrawl(job) {
					continue
				}
			}

			cp.incrementCrawlerCount()
			go crawler.
				New(cp.fetcherExtractor, job, cp.jobQueue, cp.crawlerDone).
				Crawl(ctx)

		}
	}
}

func (cp *CrawlerPool) incrementCrawlerCount() {
	atomic.AddUint64(&cp.activeCrawlers, 1)
	cp.updatedCrawlerCount <- struct{}{}
}

func (cp *CrawlerPool) decrementCrawlerCount() {
	atomic.AddUint64(&cp.activeCrawlers, ^uint64(0))
	cp.updatedCrawlerCount <- struct{}{}
}

func (cp *CrawlerPool) incrementDepthCount() {
	atomic.AddUint64(&cp.depthReached, 1)
}

func (cp CrawlerPool) getDepthCount() uint64 {
	return atomic.LoadUint64(&cp.depthReached)
}
