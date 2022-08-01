package main

import "time"

type AppConfig struct {
	FetcherExtractor FetcherExtractor
	IngressJobQueue  Queue
	CrawlerPoolConfig
}

type CrawlerPoolConfig struct {
	CrawlerPoolShutDownTimeout time.Duration
	CrawlerPoolSize            uint64
	CrawlerDepth               uint64
}
