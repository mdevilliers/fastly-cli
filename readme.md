### fastly-cli

A highly opinionated CLI to aid me in my day-to-day tasks with Fastly.
```./fastly-cli 
Usage:
  fastly-cli [command]

Available Commands:
  create      Create a new Fastly service
  help        Help about any command
  launch      Fuzzy search for a service and launch in browser.
  tokens      Manage API tokens

Flags:
  -a, --fastlyAPIKey string   Fastly API Key to use
  -h, --help                  help for fastly-cli
```

Install

```
go get github.com/mdevilliers/fastly-cli/cmd/fastly-cli
```

Fuzzy search on service names and launch web UI

```
export FASTLY_API_KEY=xxxxxxxxxx
fastly-cli launch images 
```

In the above example 'fastly-cli' will :
- search all of your services containing the term 'images'
- if only one result then launch the Fastly web UI for that service
- if some results will allow you select from a short list
- if none will display all of your services

