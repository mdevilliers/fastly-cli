### fastly-cli

A highly opinionated CLI to aid me in my day-to-day tasks with Fastly.

[![CircleCI](https://circleci.com/gh/mdevilliers/fastly-cli.svg?style=svg)](https://circleci.com/gh/mdevilliers/fastly-cli)

```
./fastly-cli
Usage:
  fastly-cli [command]

Available Commands:
  create      Create a new Fastly service
  help        Help about any command
  launch      Fuzzy search for a service and launch in browser.
  tokens      Manage API tokens

Flags:
      --fastly-api-key string         Fastly API Key
      --fastly-user-name string       Fastly user name
      --fastly-user-password string   Fastly user password
  -h, --help                  help for fastly-cli
```

#### Install

fastly-cli requires Golang 1.12.1+
```
go get github.com/mdevilliers/fastly-cli/cmd/fastly-cli
```

### Commands

#### launch

Fuzzy search on service names and launch web UI

Example

```
export FASTLY_API_KEY=xxxxxxxxxx
fastly-cli launch images 
```

In the above example 'fastly-cli' will :
- search all of your services containing the term 'images'
- if only one result then launch the Fastly web UI for that service
- if some results will allow you select from a short list
- if none will display all of your services


#### create

Create a new Fastly service and an optional API key scoped to that service.

#### tokens

View or create API tokens

