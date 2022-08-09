package jobPrinter

import (
	"monzoCrawler/domain/models"
	"strings"
)

type Logger interface {
	Printf(format string, args ...interface{})
}

type JobPrinter struct {
	logger Logger
}

func New(logger Logger) *JobPrinter {
	return &JobPrinter{logger: logger}
}

func (p *JobPrinter) Print(job models.CrawlJob) {
	generatedUrls := make([]string, 0, len(job.Result.NewJobs))
	for _, newJob := range job.Result.NewJobs {
		generatedUrls = append(generatedUrls, newJob.URL.String())
	}

	p.logger.Printf("Url: %s Links: %s", job.URL.String(), strings.Join(generatedUrls, "\n"))
	p.logger.Printf("----------------------------------------------------")
}
