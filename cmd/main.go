package main

import (
	"log"
	"monzoCrawler/domain/adapters/jobPrinter"
	"monzoCrawler/domain/models"
	"net/url"
	"os"
	"time"

	"monzoCrawler/domain/adapters/FIFOqueue"
	"monzoCrawler/domain/adapters/sameDomainFilter"
	"monzoCrawler/domain/adapters/urlFetcherExtractor"
	"monzoCrawler/domain/crawlerPool"
	"monzoCrawler/domain/hooks/storeCrawlJob"
	"monzoCrawler/domain/store"
)

func main() {
	ctx, cancel := listenForCancellationAndAddToContext()
	defer cancel()

	logger := log.New(os.Stdout, "[CrawlerPool]", log.LstdFlags)

	urlStore := store.NewUrlStore()
	jobQueue := FIFOqueue.New()

	fetcherExtractor := urlFetcherExtractor.NewHTTPFetcherExtractor(time.Minute)

	SeedURL := &url.URL{Scheme: "http", Host: "monzo.com"}
	jobQueue.Push(models.CrawlJob{SeedURL: SeedURL, URL: SeedURL})

	done := make(chan struct{}, 1)

	app := NewApp(
		AppConfig{
			FetcherExtractor: fetcherExtractor,
			IngressJobQueue:  jobQueue,
			CrawlerPoolConfig: CrawlerPoolConfig{
				Logger:                     logger,
				CrawlerPoolShutDownTimeout: time.Second * 10,
				CrawlerPoolSize:            1000,
				CrawlerDepth:               3,
				JobPrinter:                 jobPrinter.New(logger),
				JobFilters: []crawlerPool.JobFilter{
					sameDomainFilter.New(),
				},
				CompletionHook: storeHook.New(urlStore).Store,
				DoneChan:       done,
			},
		})

	app.StartCrawlerPool(ctx)

	<-done
}
