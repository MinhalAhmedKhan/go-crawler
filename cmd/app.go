package main

import (
	"context"
	"io"
	"log"
	"monzoCrawler/domain/crawlerPool"
	"monzoCrawler/domain/model"
	"net/url"
	"os"
)

type (
	FetcherExtractor interface {
		Fetch(ctx context.Context, url url.URL) (io.ReadCloser, error)
		Extract(io.Reader) (model.CrawlResult, error)
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
	logger := log.New(os.Stdout, "[CrawlerPool]", log.LstdFlags)

	// TODO: Takes in a config instead of parameters?
	cPool := crawlerPool.New(logger, cfg.CrawlerPoolSize, cfg.IngressJobQueue, cfg.CrawlerPoolShutDownTimeout, cfg.FetcherExtractor, cfg.CrawlerDepth, cfg.JobFilters, cfg.CompletionHook)
	return &App{
		crawlerPool:       cPool,
		crawlerPoolConfig: cfg.CrawlerPoolConfig,
	}
}

func (a *App) StartCrawlerPool(ctx context.Context) {
	go a.crawlerPool.Start(ctx, a.crawlerPoolConfig.DoneChan)
}
