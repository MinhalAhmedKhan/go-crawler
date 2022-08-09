package main

import (
	"context"
	"io"
	"monzoCrawler/domain/crawlerPool"
	"monzoCrawler/domain/models"
	"net/url"
)

type (
	FetcherExtractor interface {
		Fetch(ctx context.Context, url *url.URL) (io.ReadCloser, error)
		Extract(url *url.URL, contents io.Reader) (models.CrawlResult, error)
	}

	Queue interface {
		Push(val interface{}) error
		Pop() (interface{}, error)
	}
)

type App struct {
	crawlerPool       *crawlerPool.CrawlerPool
	crawlerPoolConfig CrawlerPoolConfig
}

func NewApp(cfg AppConfig) *App {
	// TODO: Takes in a config instead of parameters?
	cPool := crawlerPool.New(cfg.Logger, cfg.CrawlerPoolSize, cfg.IngressJobQueue, cfg.CrawlerPoolShutDownTimeout, cfg.FetcherExtractor, cfg.CrawlerDepth, cfg.JobPrinter, cfg.JobFilters, cfg.CompletionHook)
	return &App{
		crawlerPool:       cPool,
		crawlerPoolConfig: cfg.CrawlerPoolConfig,
	}
}

func (a *App) StartCrawlerPool(ctx context.Context) {
	go a.crawlerPool.Start(ctx, a.crawlerPoolConfig.DoneChan)
}
