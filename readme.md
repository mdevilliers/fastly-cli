### fastly-cli

A highly opinionated CLI to aid me in my day-to-day tasks with Fastly.

[![CircleCI](https://circleci.com/gh/mdevilliers/fastly-cli.svg?style=svg)](https://circleci.com/gh/mdevilliers/fastly-cli)

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

#### Install

fastly-cli requires Golang 1.12.1+
```
go get github.com/mdevilliers/fastly-cli/cmd/fastly-cli
```

#### launch

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


#### create

Create a new Fastly service and an API key scoped to that service.
``````
./fastly-cli create -h
Create a new Fastly service

Usage:
  fastly-cli create [flags]

Flags:
      --create-api-token      create an API token (default true)
      --enable-2FA            use 2FA. If enabled you will be asked to provide a token when creating an API user (default true)
  -h, --help                  help for create
      --service-name string   name of service to create
      --token-name string     name of the API token to create. Defaults to the service-name if not supplied

Global Flags:
  -a, --fastlyAPIKey string   Fastly API Key to use
``````

#### tokens

View API tokens
``````
./fastly-cli tokens -h
Manage API tokens

Usage:
  fastly-cli tokens [flags]
  fastly-cli tokens [command]

Available Commands:
  all         List all API tokens

Flags:
  -h, --help   help for tokens

Global Flags:
  -a, --fastlyAPIKey string   Fastly API Key to use
``````

