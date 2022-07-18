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
	ctx, done := listenForCancellationAndAddToContext()
	defer done()

	logger := log.New(os.Stdout, "", log.LstdFlags)

	jobQueue := FIFOqueue.NewFIFOQueue()

	fetcherExtractor := urlFetcherExtractor.NewHTTPFetcherExtractor(time.Minute)
	shutdownTimeout := 10 * time.Second

	pool := crawlerPool.NewCrawlerPool(logger, 2, jobQueue, shutdownTimeout, fetcherExtractor)

	pool.AddJobToQueue()

	go pool.Start(ctx, 3)

	pool.Wait()
}
