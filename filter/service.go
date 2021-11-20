package filter

import (
	"fmt"
	"strings"

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

const (
	rewriteRule = "substack_paywall"
)

// RunFilterJob runs the filtering job, which marks paywalled entries as read.
func (s *Service) RunFilterJob() error {
	entries, err := s.client.Entries(&miniflux.Filter{Status: miniflux.EntryStatusUnread})
	if err != nil {
		return err
	}

	feedsResp, err := s.client.Feeds()
	if err != nil {
		return err
	}
	feeds := make(map[int64]*miniflux.Feed)
	for _, f := range feedsResp {
		feeds[f.ID] = f
	}

	var paywalledEntries []int64
	for _, entry := range entries.Entries {
		feed, ok := feeds[entry.FeedID]
		feedOptIn := ok && strings.Contains(feed.RewriteRules, rewriteRule)
		if !feedOptIn && !strings.Contains(entry.Feed.FeedURL, "substack.com") {
			continue
		} else if s.cache.Contains(entry.ID) {
			level.Debug(s.l).Log("msg", "skipping cached entry", "entry_id", entry.ID)
			continue
		}
		level.Debug(s.l).Log("msg", "scraping entry", "url", entry.URL, "entry_id", entry.ID)

		paywalled, err := IsURLPaywalled(s.l, entry.URL)
		if err != nil {
			level.Error(s.l).Log("msg", "checking paywall", "err", err)
			continue
		}

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
