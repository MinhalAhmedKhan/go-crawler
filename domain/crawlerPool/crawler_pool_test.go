package crawlerPool_test

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"io"
	"log"
	"monzoCrawler/domain/adapters/FIFOqueue"
	"monzoCrawler/domain/model"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"monzoCrawler/domain/crawlerPool"
	"monzoCrawler/domain/crawlerPool/internal/mocks"
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
		err := jobQueue.Push(model.CrawlJob{SeedURL: &url.URL{Scheme: "http", Host: siteToCrawl}})
		assert.NoError(t, err, "expected no error when pushing a job to the queue")

		crawlerStartedFetching := make(chan struct{}, 1)

		mockFetcherExtractor := &mocks.FetcherExtractorMock{
			FetchFunc: func(ctx context.Context, urlMoqParam url.URL) (io.ReadCloser, error) {
				assert.Equal(t, siteToCrawl, urlMoqParam.Host, "fetch called with wrong url")
				crawlerStartedFetching <- struct{}{}

				return io.NopCloser(strings.NewReader("Monzo")), nil
			},
			ExtractFunc: func(io.Reader) (model.CrawlResult, error) {
				return model.CrawlResult{}, nil
			},
		}

		cp := crawlerPool.New(logger, 2, jobQueue, time.Second, mockFetcherExtractor, 1)

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
			FetchFunc: func(ctx context.Context, urlMoqParam url.URL) (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader("Monzo")), nil
			},
			ExtractFunc: func(io.Reader) (model.CrawlResult, error) {
				return model.CrawlResult{}, nil
			},
		}

		cp := crawlerPool.New(logger, 5, FIFOqueue.New(), time.Second, mockFetcherExtractor, 10)

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
		doneChan := make(chan struct{}, 1)

		mockFetcherExtractor := &mocks.FetcherExtractorMock{
			FetchFunc: func(ctx context.Context, urlMoqParam url.URL) (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader("Monzo")), nil
			},
			ExtractFunc: func(io.Reader) (model.CrawlResult, error) {
				return model.CrawlResult{}, nil
			}
		}

		crawlerPool := crawlerPool.New(logger, 5, FIFOqueue.New(), time.Second, mockFetcherExtractor, 1)

		go crawlerPool.Start(ctx, doneChan)


	})
}
