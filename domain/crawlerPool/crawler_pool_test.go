package crawlerPool_test

import (
	"io"
	"log"
	"monzoCrawler/domain/adapters/FIFOqueue"
	"monzoCrawler/domain/crawlerPool"
	"monzoCrawler/domain/crawlerPool/internal/mocks"
	"monzoCrawler/domain/models"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestCrawlerPool_Start(t *testing.T) {
	logger := log.New(os.Stdout, "[CrawlerPoolTest]", log.LstdFlags)

	t.Run("start crawler when a job is picked up", func(t *testing.T) {
		const siteToCrawl = "monzo.com"

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		doneChan := make(chan struct{}, 1)

		// Given a job in the queue
		jobQueue := FIFOqueue.New()
		err := jobQueue.Push(models.CrawlJob{URL: &url.URL{Scheme: "http", Host: siteToCrawl}})
		assert.NoError(t, err, "expected no error when pushing a job to the queue")

		crawlerStartedFetching := make(chan struct{}, 1)

		mockFetcherExtractor := &mocks.FetcherExtractorMock{
			FetchFunc: func(ctx context.Context, urlMoqParam *url.URL) (io.ReadCloser, error) {
				assert.Equal(t, siteToCrawl, urlMoqParam.Host, "fetch called with wrong url")
				crawlerStartedFetching <- struct{}{}

				return io.NopCloser(strings.NewReader("Monzo")), nil
			},
			ExtractFunc: func(urlMoqParam *url.URL, contents io.Reader) (models.CrawlResult, error) {
				return models.CrawlResult{}, nil
			},
		}

		cp := crawlerPool.New(logger, 2, jobQueue, time.Second, mockFetcherExtractor, 1, emptyJobPrinter(), nil, crawlerPool.NoOpCompletedHook)

		// When the crawler pool is started
		go cp.Start(ctx, doneChan)

		select {
		case <-crawlerStartedFetching:
			// Then the crawler should start and fetch the site
			return
		case <-time.After(time.Second):
			t.Error("crawler did not start and finish")
		}
	})
	t.Run("stop pool when context is cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		doneChan := make(chan struct{}, 1)

		mockFetcherExtractor := &mocks.FetcherExtractorMock{
			FetchFunc: func(ctx context.Context, urlMoqParam *url.URL) (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader("Monzo")), nil
			},
			ExtractFunc: func(urlMoqParam *url.URL, contents io.Reader) (models.CrawlResult, error) {
				return models.CrawlResult{}, nil
			},
		}

		cp := crawlerPool.New(logger, 5, FIFOqueue.New(), time.Second, mockFetcherExtractor, 10, emptyJobPrinter(), nil, crawlerPool.NoOpCompletedHook)

		go cp.Start(ctx, doneChan)

		// when we cancel the context, the pool should stop
		cancel()

		select {
		case <-doneChan:
			t.Log("crawler pool stopped")
			return
		case <-time.After(time.Second * 5):
			t.Fatal("crawler pool did not stop when context was cancelled")
		}
	})
	t.Run("stop pool when max depth reached", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		const maxDepth = 5
		doneChan := make(chan struct{}, 1)

		mockFetcherExtractor := &mocks.FetcherExtractorMock{
			FetchFunc: func(ctx context.Context, urlMoqParam *url.URL) (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader("Monzo")), nil
			},
			ExtractFunc: func(urlMoqParam *url.URL, contents io.Reader) (models.CrawlResult, error) {
				return models.CrawlResult{}, nil
			},
		}

		jobQueue := FIFOqueue.New()
		// Given a job in the queue with a max depth of greater than the max depth we want to crawl
		for i := 1; i <= maxDepth+1; i++ {
			jobQueue.Push(models.CrawlJob{URL: &url.URL{Scheme: "http", Host: "monzo.com"}, Depth: uint64(i)})
		}

		cp := crawlerPool.New(logger, 5, jobQueue, time.Second, mockFetcherExtractor, maxDepth, emptyJobPrinter(), nil, crawlerPool.NoOpCompletedHook)

		// When the crawler pool is started
		go cp.Start(ctx, doneChan)

		select {
		case <-doneChan:
			// Then the crawler should stop and not crawl the site
			return
		case <-time.After(time.Second):
			t.Fatal("crawler pool did not stop when max depth was reached")
		}
	})
	t.Run("call completion hook when crawler is done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		doneChan := make(chan struct{}, 1)
		comletionHookCalled := make(chan struct{}, 1)

		mockFetcherExtractor := &mocks.FetcherExtractorMock{
			FetchFunc: func(ctx context.Context, urlMoqParam *url.URL) (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader("Monzo")), nil
			},
			ExtractFunc: func(urlMoqParam *url.URL, contents io.Reader) (models.CrawlResult, error) {
				return models.CrawlResult{}, nil
			},
		}

		jobQueue := FIFOqueue.New()
		jobQueue.Push(models.CrawlJob{URL: &url.URL{Scheme: "http", Host: "monzo.com"}})

		cp := crawlerPool.New(logger, 5, jobQueue, time.Second, mockFetcherExtractor, 10, emptyJobPrinter(), nil, func(ctx context.Context, job models.CrawlJob) {
			comletionHookCalled <- struct{}{}
		})

		go cp.Start(ctx, doneChan)

		select {
		case <-comletionHookCalled:
			return
		case <-time.After(time.Second):
			t.Fatal("crawler pool did not call completion hook when crawler is done")
		}
	})
}

func emptyJobPrinter() *mocks.JobPrinterMock {
	return &mocks.JobPrinterMock{
		PrintFunc: func(job models.CrawlJob) {},
	}
}
