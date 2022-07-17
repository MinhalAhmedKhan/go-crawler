package main

import (
	"log"
	"os"
	"time"

	"monzoCrawler/domain/adapters/FIFOqueue"
	"monzoCrawler/domain/adapters/urlFetcherExtractor"
	"monzoCrawler/domain/crawlerPool"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	ctx, done := listenForCancellationAndAddToContext()
	defer done()

	jobQueue := FIFOqueue.NewFIFOQueue()

	fetcherExtractor := urlFetcherExtractor.NewHTTPFetcherExtractor(time.Minute)
	shutdownTimeout := 10 * time.Second

	pool := crawlerPool.NewCrawlerPool(logger, 2, jobQueue, shutdownTimeout, fetcherExtractor)

	pool.AddJobToQueue()

	go pool.Start(ctx)

	pool.Wait()
}
