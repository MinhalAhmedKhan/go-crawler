package main

import (
	"net/url"
	"time"

	"monzoCrawler/domain/adapters/FIFOqueue"
	"monzoCrawler/domain/adapters/sameDomainFilter"
	"monzoCrawler/domain/adapters/urlFetcherExtractor"
	"monzoCrawler/domain/crawlerPool"
	"monzoCrawler/domain/hooks/storeCrawlJob"
	"monzoCrawler/domain/model"
	"monzoCrawler/domain/store"
)

func main() {
	ctx, cancel := listenForCancellationAndAddToContext()
	defer cancel()

	urlStore := store.NewUrlStore()
	jobQueue := FIFOqueue.New()

	fetcherExtractor := urlFetcherExtractor.NewHTTPFetcherExtractor(time.Minute)

	SeedURL := &url.URL{Scheme: "http", Host: "monzo.com"}
	jobQueue.Push(model.CrawlJob{SeedURL: SeedURL, URL: SeedURL})

	done := make(chan struct{}, 1)

	app := NewApp(
		AppConfig{
			FetcherExtractor: fetcherExtractor,
			IngressJobQueue:  jobQueue,
			CrawlerPoolConfig: CrawlerPoolConfig{
				CrawlerPoolShutDownTimeout: time.Second * 10,
				CrawlerPoolSize:            1000,
				CrawlerDepth:               10,
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
