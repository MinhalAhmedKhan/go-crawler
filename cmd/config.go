package main

import (
	"monzoCrawler/domain/crawlerPool"
	"time"
)

type AppConfig struct {
	FetcherExtractor FetcherExtractor
	IngressJobQueue  Queue
	CrawlerPoolConfig
}

type CrawlerPoolConfig struct {
	CrawlerPoolShutDownTimeout time.Duration
	CrawlerPoolSize            uint64
	CrawlerDepth               uint64
	JobFilters                 []crawlerPool.JobFilter
	CompletionHook             crawlerPool.CrawlerCompletedHook
	DoneChan                   chan struct{}
}
