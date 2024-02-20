### Example of web-crowler

## how to 

## How to run

There are two parameters on start:
 1. URL to start the web-crowler
 2. Configuration file (optional)

### Run with Makefile
Run "make run <http://example.com>" to start the web-crowler.
### Build
Run "make build" to build and run the artifact from ./build directory.
### Configuration
The default configuration file is located in `./configs/config.yaml` by default
```yaml
parallelism: 10 # number of parallel requests
acceptable_mime_types: # acceptable mime types on resolving response
  - text/html
  - application/json
  - application/xml
  - text/css
database_file: ./crawler.db # path for the database file
api_addr: localhost:8080 # address for the API
```

The stataistics API is available on `http://localhost:8080/` (just a few counters which are barely useful)

## How to test
Run "make test" to run the tests (TWO tests). 


