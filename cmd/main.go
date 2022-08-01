package main

import (
	"monzoCrawler/domain/model"
	"net/url"
	"time"

	"monzoCrawler/domain/adapters/FIFOqueue"
	"monzoCrawler/domain/adapters/urlFetcherExtractor"
)

func main() {
	ctx, done := listenForCancellationAndAddToContext()
	defer done()

	jobQueue := FIFOqueue.New()
	fetcherExtractor := urlFetcherExtractor.NewHTTPFetcherExtractor(time.Minute)

	targetURL, err := url.Parse("https://google.com")
	if err != nil {
		return
	}
	jobQueue.Push(model.CrawlJob{SeedURL: targetURL})

	app := NewApp(
		AppConfig{
			FetcherExtractor: fetcherExtractor,
			IngressJobQueue:  jobQueue,
			CrawlerPoolConfig: CrawlerPoolConfig{
				CrawlerPoolShutDownTimeout: time.Second * 10,
				CrawlerPoolSize:            1,
				CrawlerDepth:               2,
			},
		})

	app.StartCrawlerPool(ctx)

	app.WaitForCrawlerPoolToComplete()
}
