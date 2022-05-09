# ITSM Reporting service

Requires Go ver. >= 1.16

`make test` runs unit tests

to regenerate SQL calls to be used in tests (using mocking library `copyist`) use command 
`COPYIST_RECORD=1 go test -v -count=1 -race -timeout 10s ./internal/...`
with connection to the running test database correctly established

`make run` starts application for local use/testing

`make docs` starts API documentation server on default port 3001;
you can specify different port: `make docs PORT=3002`

`make swagger` regenerates swagger.yaml file from the source code (usually no need to use unless API changes)
