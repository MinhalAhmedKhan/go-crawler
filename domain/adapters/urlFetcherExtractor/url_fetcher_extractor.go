package urlFetcherExtractor

import (
	"context"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"time"

	"monzoCrawler/domain/model"
)

type HTTPFetcherExtractor struct {
	client *http.Client
}

func NewHTTPFetcherExtractor(timeout time.Duration) HTTPFetcherExtractor {
	return HTTPFetcherExtractor{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (fe HTTPFetcherExtractor) Fetch(ctx context.Context, url url.URL) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	resp, err := fe.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (fe HTTPFetcherExtractor) Extract(contents io.Reader) (model.CrawlResult, error) {
	return fe.getLinks(contents), nil
}

//Collect all links from response body and return it as an array of strings
func (fe *HTTPFetcherExtractor) getLinks(body io.Reader) model.CrawlResult {
	crawlResult := model.CrawlResult{NewJobs: []model.CrawlJob{}}

	z := html.NewTokenizer(body)
	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			//todo: links list shoudn't contain duplicates
			return crawlResult
		case html.StartTagToken, html.EndTagToken:
			token := z.Token()
			if "a" == token.Data {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						targetURL, err := url.Parse(attr.Val)
						if err != nil {
							continue
						}
						//targetURL.ResolveReference()
						//fmt.Println(targetURL.String(), "is a link", attr.Val)
						crawlResult.NewJobs = append(crawlResult.NewJobs, model.CrawlJob{SeedURL: targetURL})
					}
				}
			}

		}
	}
}
