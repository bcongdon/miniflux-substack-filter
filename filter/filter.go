package filter

import (
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

const (
	paywalledThreadToken = "This thread is only visible to paying subscribers of"
)

func IsURLPaywalled(logger log.Logger, url string) (bool, error) {
	res, err := http.Get(url)
	if err != nil || res.StatusCode != 200 {
		var status int
		if res != nil {
			status = res.StatusCode
		}
		level.Error(logger).Log("msg", "unable to get entry body", "err", err, "status_code", status)
		return false, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		level.Error(logger).Log("msg", "unable to parse entry body", "err", err)
		return false, err
	}
	articlePaywall := doc.Find(".paywall").Length() > 0
	threadPaywall := strings.Contains(doc.Find(".thread-head").Text(), paywalledThreadToken)
	paywalled := articlePaywall || threadPaywall
	level.Debug(logger).Log("msg", "fetched substack article", "url", url, "article_paywall", articlePaywall, "thread_paywall", threadPaywall)
	return paywalled, nil
}
