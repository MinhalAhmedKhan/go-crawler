package crawlerPool

import (
	"context"
	"io"
	"sync/atomic"
	"time"

	"monzoCrawler/dao"
	"monzoCrawler/domain/crawler"
)

type (
	Logger interface {
		Printf(format string, args ...interface{})
	}

	Queue interface {
		Push(val interface{}) error
		Pop() (interface{}, error)
		Size() uint64
	}

	FetcherExtractor interface {
		Fetch(ctx context.Context, url string) (io.ReadCloser, error)
		Extract(io.Reader) (dao.CrawlResult, error)
	}
)

// CrawlerPool manages crawlers running.
type CrawlerPool struct {
	logger Logger

	size     uint64 // Number of crawlers
	jobQueue Queue  // Jobs to be processed.

	fetcherExtractor FetcherExtractor

	activeCrawlers      uint64        // Number of active crawlers.
	updatedCrawlerCount chan struct{} // Channel to signal that crawler count has been updated.
	crawlerDone         chan struct{} // Channel to signal that a crawler is done.

	shutdownTimeout time.Duration    // Timeout for shutdown.
	forceShutdown   chan interface{} // Channel to signal that the pool should forcefully shut down.
}

func NewCrawlerPool(logger Logger, size uint64, jobQueue Queue, shutdownTimeout time.Duration, fetcherExtractor FetcherExtractor) *CrawlerPool {
	return &CrawlerPool{
		logger: logger,

		size:     size,
		jobQueue: jobQueue,

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
			//TODO: Fix waiting logic. doesnt work well with 1 active crawler as it exits
			if atomic.LoadUint64(&cp.activeCrawlers) == 0 {
				return
			}
		}
	}
}

// Start crawling up to the specified depth.
func (cp *CrawlerPool) Start(ctx context.Context, depth int) {
	go cp.listenForCompletedJobs(ctx)

	for {
		select {
		case <-ctx.Done():
			// graceful shutdown
			cp.Wait()
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

			job, ok := jobPickedUp.(dao.CrawlJob)

			if !ok {
				//TODO: log Invalid job type.
				cp.logger.Printf("Invalid job type found, got job of type %T", jobPickedUp)
				continue
			}

			if job.Depth > depth {
				// picked up first job that is too deep.
				// given a FIFO queue, the first job that is too deep is the start of the new depth.
				// therefore crawl is compeleted.
				return
			}

			cp.incrementCrawlerCount()
			go crawler.
				NewCrawler(cp.fetcherExtractor, job, cp.jobQueue, cp.crawlerDone).
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

func (cp *CrawlerPool) AddJobToQueue() {
	cp.jobQueue.Push(dao.CrawlJob{SeedURL: "https://google.com", Depth: 0})
}
