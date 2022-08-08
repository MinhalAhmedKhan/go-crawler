package storeHook

import (
	"context"
	"log"
	"monzoCrawler/domain/model"
)

type Store interface {
	Add(url string) error
}

type StoreHook struct {
	store Store
}

func New(store Store) *StoreHook {
	return &StoreHook{
		store: store,
	}
}

func (h *StoreHook) Store(ctx context.Context, job model.CrawlJob) {
	if err := h.store.Add(job.URL.String()); err != nil {
		log.Printf("Error storing job: %s", err)
	}
}
