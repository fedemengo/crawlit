# Web crawler

At the moment the web crawler is configurable only on its construction. All crawling is performed concurrently, so `c.Crawl(handler)` is **non-blocking** and it calls the `handler` function for every valid url discovered. `c.Result()`, on the other hand, is blocking and once called it consume the result of the crawling.

Consider a stateless version, configurable by specifying parameter to `c.Crawl()` so that crawling with different configuration can be performed by the same object. Crawler before [954b692](https://github.com/fedemengo/crawlit/commit/954b692e173b169b7c54547e7dc2d661ad09a1b5) can be a good starting point

## Web crawler

At the moment the web crawler is configurable only on its construction. All crawling is performed concurrently, so `c.Crawl(handler)` is **non-blocking** and it calls the `handler` function for every valid url discovered. `c.Result()`, on the other hand, is blocking and once called it consume the result of the crawling.

Consider a stateless version, configurable by specifying parameter to `c.Crawl()` so that crawling with different configuration can be performed by the same object. Crawler before [954b692](https://github.com/fedemengo/search-engine/commit/954b692e173b169b7c54547e7dc2d661ad09a1b5) can be a good starting point

