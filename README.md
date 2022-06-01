# LetterRest

Provide a RESTful API against letterboxd.com

TODO: Collections [Example](https://letterboxd.com/films/in/halloween-collection/)

## CLI Usage

### Server

Use `letterrestd server` to start a restful API server. Hit up the swagger docs
[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

### Scrape Client

This is mainy useful for testing out the scrape capabilities. Use `letterrestd
scrape -h` to see the options here. These commands interact directly with the
letterboxd.com website, not through a legit API.

Found in the [cli/](cli/) directory.

### API Client Library

This should be more useful than the scraper. Interacts directly with the restful
API from the command line. Check it out with `letterrestd api -h'.

Found in the [client/](client/) directory.
