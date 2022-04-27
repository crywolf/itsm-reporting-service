# ITSM Reporting service

Requires Go ver. >= 1.16

`make test` runs unit tests

`make run` starts application for local use/testing

`make docs` starts API documentation server on default port 3001;
you can specify different port: `make docs PORT=3002`

`make swagger` regenerates swagger.yaml file from source code (usually no need to use unless API changes)
