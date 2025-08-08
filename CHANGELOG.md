# Changelog

## v0.0.1

- youtube music support (gets the top music by scraping the youtube music
  website)

## v0.1.0

- attempt to use the gitlab/github cicd (broken because youtube does not allow
  gitlab/github cicd to scrape their website) so its mostly useless

## v1.0.0

- add detailed documentation on how to run, how to use cicd, etc...
- make it more robust to re-runs by deleting the folders it creates to prevent
  errors

## v1.1.0

- add support for last.fm and listenbrainz (suggestion request very cool)\
  you can use multiple providers (lastfm, listenbrainz, youtube) to get the
  favorite music from a custom order (set by --fallback) where we try to get the
  first one, and if it fails we try the other ones

## v1.1.1

- fix: dont fail if .env cannot be loaded

## v1.1.2

- fix: images getting generated incorrectly in markdown
