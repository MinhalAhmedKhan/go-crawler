package main

import (
	"monzoCrawler/domain/model"
	"net/url"
	"time"

	"monzoCrawler/domain/adapters/FIFOqueue"
	"monzoCrawler/domain/adapters/urlFetcherExtractor"
)

func main() {
	ctx, cancel := listenForCancellationAndAddToContext()
	defer cancel()

	jobQueue := FIFOqueue.New()
	fetcherExtractor := urlFetcherExtractor.NewHTTPFetcherExtractor(time.Minute)

	jobQueue.Push(model.CrawlJob{SeedURL: &url.URL{Scheme: "http", Host: "monzo.com"}})

	done := make(chan struct{}, 1)

	app := NewApp(
		AppConfig{
			FetcherExtractor: fetcherExtractor,
			IngressJobQueue:  jobQueue,
			CrawlerPoolConfig: CrawlerPoolConfig{
				CrawlerPoolShutDownTimeout: time.Second * 10,
				CrawlerPoolSize:            100,
				CrawlerDepth:               3,
				DoneChan:                   done,
			},
		})

	app.StartCrawlerPool(ctx)

	<-done
}
