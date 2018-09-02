## Web crawler

Stateless web crawler configurable on each crawling call. All crawling is performed concurrently, so `c.Crawl(config, handler)` is **non-blocking** and it calls the `handler` function for every valid url discovered. `c.Result()`, on the other hand, is blocking and once called it consume the result of the **one** crawling.
