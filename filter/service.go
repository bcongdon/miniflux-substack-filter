package filter

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	lru "github.com/hashicorp/golang-lru"
	miniflux "miniflux.app/client"
)

// Service represents a substack filter service.
type Service struct {
	client *miniflux.Client
	l      log.Logger
	dryRun bool
	cache  *lru.Cache
}

// New creates a new Service.
func New(client *miniflux.Client, logger log.Logger, dryRun bool) (*Service, error) {
	cache, err := lru.New(1024)
	if err != nil {
		return nil, err
	}
	return &Service{
		client: client,
		l:      logger,
		dryRun: dryRun,
		cache:  cache, // Contains list of fetched entries.
	}, nil
}

// RunFilterJob runs the filtering job, which marks paywalled entries as read.
func (s *Service) RunFilterJob() error {
	f, err := s.client.Entries(&miniflux.Filter{Status: miniflux.EntryStatusUnread})
	if err != nil {
		return err
	}

	var paywalledEntries []int64
	for _, entry := range f.Entries {
		if !strings.Contains(entry.Feed.FeedURL, "substack") {
			continue
		} else if s.cache.Contains(entry.ID) {
			level.Debug(s.l).Log("msg", "skipping cached entry", "entry_id", entry.ID)
			continue
		}
		res, err := http.Get(entry.URL)
		if err != nil || res.StatusCode != 200 {
			level.Error(s.l).Log("msg", "unable to get entry body", "err", err, "status_code", res.StatusCode)
			continue
		}
		defer res.Body.Close()

		doc, err := goquery.NewDocumentFromResponse(res)
		if err != nil {
			level.Error(s.l).Log("msg", "unable to parse entry body", "err", err)
			continue
		}
		paywalled := doc.Find("article.post .paywall").Length() > 0
		level.Debug(s.l).Log("msg", "fetched substack article", "url", entry.URL, "paywalled", paywalled)

		s.cache.Add(entry.ID, true)
		if !paywalled {
			continue
		}
		paywalledEntries = append(paywalledEntries, entry.ID)
	}

	if len(paywalledEntries) == 0 {
		return nil
	}
	if s.dryRun {
		level.Info(s.l).Log("msg", "would have marked entries as read", "entry_ids", fmt.Sprintf("%v", paywalledEntries))
		return nil
	}
	return s.client.UpdateEntries(paywalledEntries, miniflux.EntryStatusRead)
}
