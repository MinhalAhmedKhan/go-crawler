//go:generate moq -out internal/mocks/queue_moq.go -pkg mocks . Queue

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
)

// CrawlerPool manages crawlers running.
type CrawlerPool struct {
	logger Logger

	size         uint64 // Number of crawlers
	maxDepth     uint64 // Max depth to crawl
	depthReached uint64 // Depth reached by this pool.
	jobQueue     Queue  // Jobs to be processed.

	fetcherExtractor FetcherExtractor

	activeCrawlers      uint64        // Number of active crawlers.
	updatedCrawlerCount chan struct{} // Channel to signal that crawler count has been updated.
	crawlerDone         chan struct{} // Channel to signal that a crawler is done.

	shutdownTimeout time.Duration    // Timeout for shutdown.
	forceShutdown   chan interface{} // Channel to signal that the pool should forcefully shut down.

}

func New(logger Logger, size uint64, jobQueue Queue, shutdownTimeout time.Duration, fetcherExtractor FetcherExtractor, maxDepth uint64) *CrawlerPool {
	return &CrawlerPool{
		logger: logger,

		size:         size,
		maxDepth:     maxDepth,
		depthReached: 0,
		jobQueue:     jobQueue,

		fetcherExtractor: fetcherExtractor,

		activeCrawlers:      0,
		updatedCrawlerCount: make(chan struct{}),
		crawlerDone:         make(chan struct{}),

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
		case <-cp.crawlerDone:
			// A crawler completed its job.
			cp.decrementCrawlerCount()

		}
	}
}

// Wait blocks until all crawlers have exited or shutdown timed out forcing a shutdown.
func (cp *CrawlerPool) Wait() {
	for {
		select {
		case <-cp.forceShutdown:
			return
		case <-cp.updatedCrawlerCount:
			if atomic.LoadUint64(&cp.activeCrawlers) == 0 && cp.GetDepthCount() == cp.maxDepth {
				return
			}
		}
	}
}

// Start crawling up to the specified depth.
func (cp *CrawlerPool) Start(ctx context.Context) {
	go cp.listenForCompletedJobs(ctx)

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
				//TODO: log Invalid job type.
				cp.logger.Printf("Invalid job type found, got job of type %T", jobPickedUp)
				continue
			}

			if job.Depth > cp.maxDepth {
				// picked up first job that is too deep.
				// given a breadth first search, the first job that is too deep is the start of the new depth.
				// therefore, crawl is completed for the specified depth.
				return
			}

			if job.Depth > cp.GetDepthCount() {
				cp.incrementDepthCount()
			}

			cp.logger.Printf("Starting crawler for %s", job.SeedURL)

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

func (cp CrawlerPool) GetDepthCount() uint64 {
	return atomic.LoadUint64(&cp.depthReached)
}
