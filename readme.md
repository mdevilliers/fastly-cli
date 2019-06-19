### fastly-cli

A highly opinionated CLI to aid me in my day-to-day tasks with Fastly.

[![CircleCI](https://circleci.com/gh/mdevilliers/fastly-cli.svg?style=svg)](https://circleci.com/gh/mdevilliers/fastly-cli)
[![ReportCard](https://goreportcard.com/badge/github.com/mdevilliers/fastly-cli)](https://goreportcard.com/report/github.com/mdevilliers/fastly-cli)

```
./fastly-cli

Usage:
  fastly-cli [command]

Available Commands:
  create      Create a new Fastly service
  eavesdrop   Listen in to your Fastly instance.
  help        Help about any command
  launch      Fuzzy search for a service and launch in browser.
  sync        Sync local CSV files with Fastly edge dictionaries.
  tokens      Manage API tokens

Flags:
      --fastly-api-key string         Fastly API Key (export FASTLY_API_KEY=xxxx)
      --fastly-user-name string       Fastly user name (export FASTLY_USER_NAME=xxxx)
      --fastly-user-password string   Fastly user password (export FASTLY_USER_PASSWORD=xxxx)
  -h, --help                          help for fastly-cli

Use "fastly-cli [command] --help" for more information about a command.
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

#### eavesdrop

Add a syslog logger to a service and stream the output as a series of JSON lines via a local TCP server.

The original service is cloned, a syslog listener added and made active. On shutdown the syslog listener is removed.

``````
./fastly-cli --fastly-api-key=xxxxxxxxxxxxx --endpoint=my.external.com --port=10089 eavesdrop servicename

{ "type": "req","service_id": "foo","request_id": "(null)","start_time": "1559726730","fastly_info": "MISS", "datacenter": "LCY","client_ip": "88.202.148.160", "req_method": "GET", "req_uri": "/h", "req_h_host": "www.bar.com", "req_h_referer": "", "req_h_user_agent": "curl/7.58.0", "req_h_accept_encoding": "", "req_header_bytes": "107", "req_body_bytes": "0", "resp_status": "404", "resp_bytes": "71044", "resp_header_bytes": "681", "resp_body_bytes": "70363" }
{ "type": "req","service_id": "foo","request_id": "(null)","start_time": "1559726745","fastly_info": "MISS", "datacenter": "LCY","client_ip": "88.202.148.160", "req_method": "GET", "req_uri": "/hi", "req_h_host": "www.bar.com", "req_h_referer": "", "req_h_user_agent": "curl/7.58.0", "req_h_accept_encoding": "", "req_header_bytes": "108", "req_body_bytes": "0", "resp_status": "404", "resp_bytes": "71064", "resp_header_bytes": "699", "resp_body_bytes": "70365" }
{ "type": "req","service_id": "foo","request_id": "(null)","start_time": "1559726746","fastly_info": "MISS", "datacenter": "LCY","client_ip": "88.202.148.160", "req_method": "GET", "req_uri": "/hi", "req_h_host": "www.bar.com", "req_h_referer": "", "req_h_user_agent": "curl/7.58.0", "req_h_accept_encoding": "", "req_header_bytes": "108", "req_body_bytes": "0", "resp_status": "404", "resp_bytes": "71046", "resp_header_bytes": "681", "resp_body_bytes": "70365" }
{ "type": "req","service_id": "foo","request_id": "(null)","start_time": "1559726748","fastly_info": "MISS", "datacenter": "LCY","client_ip": "88.202.148.160", "req_method": "GET", "req_uri": "/hij", "req_h_host": "www.bar.com", "req_h_referer": "", "req_h_user_agent": "curl/7.58.0", "req_h_accept_encoding": "", "req_header_bytes": "109", "req_body_bytes": "0", "resp_status": "404", "resp_bytes": "71057", "resp_header_bytes": "690", "resp_body_bytes": "70367" }
{ "type": "req","service_id": "foo","request_id": "(null)","start_time": "1559726750","fastly_info": "MISS", "datacenter": "LCY","client_ip": "88.202.148.160", "req_method": "GET", "req_uri": "/hijk", "req_h_host": "www.bar.com", "req_h_referer": "", "req_h_user_agent": "curl/7.58.0", "req_h_accept_encoding": "", "req_header_bytes": "110", "req_body_bytes": "0", "resp_status": "404", "resp_bytes": "71076", "resp_header_bytes": "707", "resp_body_bytes": "70369" }
{ "type": "req","service_id": "foo","request_id": "(null)","start_time": "1559726753","fastly_info": "MISS", "datacenter": "LCY","client_ip": "88.202.148.160", "req_method": "GET", "req_uri": "/hijkl", "req_h_host": "www.bar.com", "req_h_referer": "", "req_h_user_agent": "curl/7.58.0", "req_h_accept_encoding": "", "req_header_bytes": "111", "req_body_bytes": "0", "resp_status": "404", "resp_bytes": "71071", "resp_header_bytes": "700", "resp_body_bytes": "70371" }
{ "type": "req","service_id": "foo","request_id": "(null)","start_time": "1559726930","fastly_info": "MISS", "datacenter": "LHR","client_ip": "18.130.227.222", "req_method": "GET", "req_uri": "/favicon.ico", "req_h_host": "www.bar.com", "req_h_referer": "", "req_h_user_agent": "Slack-ImgProxy (+https://api.slack.com/robots)", "req_h_accept_encoding": "gzip", "req_header_bytes": "184", "req_body_bytes": "0", "resp_status": "200", "resp_bytes": "2508", "resp_header_bytes": "659", "resp_body_bytes": "1849" }``


``````
#### sync

Sync local CSV files with an existing edge dictionary.

CSV files are of the format KEY,VALUE (see ./fixtures)
```
./fastly-cli sync --dict={{DICTIONARY_NAME}} --path={{PATH TO CSV FILE}} --service={{SERVICE_NAME}}
```
Updates are batched as a series of creates, deletes and updates.

#### create

Create a new Fastly service and an optional API key scoped to that service.

#### tokens

View or create API tokens

