# miniflux-substack-filter

Filter paywalled substack posts from miniflux

## Usage

### CLI Usage

```
Usage of miniflux-substack-filter:
  -api-endpoint string
    	the api of your miniflux instance (default "https://rss.notmyhostna.me")
  -api-key string
    	api key used for authentication
  -dry-run
    	whether to start in dry run mode
  -log-level string
    	the level to filter logs at eg. debug, info, warn, error
  -password string
    	the password used to log into miniflux
  -refresh-interval string
    	interval defining how often we check for new entries in miniflux
  -username string
    	the username used to log into miniflux
```

### Enabling for a feed

Feeds that have `substack.com` in the URL are automatically filtered for
paywalled articles. You can also enable filtering on non-substack URLs (e.g.
substack publications that have a custom domain) by adding `substack_paywall` to
the Miniflux [rewrite rules](https://miniflux.app/docs/rules.html#rewrite-rules)
of the feed.

## Attribution

Inspired by
[dewey/miniflux-sidekick](https://github.com/dewey/miniflux-sidekick/)
