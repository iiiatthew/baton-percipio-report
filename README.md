![Baton Logo](./docs/images/baton-logo.png)

# `baton-percipio-report` [![Go Reference](https://pkg.go.dev/badge/github.com/iiiatthew/baton-percipio-report.svg)](https://pkg.go.dev/github.com/iiiatthew/baton-percipio-report) ![main ci](https://github.com/iiiatthew/baton-percipio-report/actions/workflows/main.yaml/badge.svg)

`baton-percipio-report` is a connector for built using the [Baton SDK](https://github.com/conductorone/baton-sdk).

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

# Getting Started

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-percipio-report
baton-percipio-report
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_DOMAIN_URL=domain_url -e BATON_API_KEY=apiKey -e BATON_USERNAME=username ghcr.io/iiiatthew/baton-percipio-report:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/iiiatthew/baton-percipio-report/cmd/baton-percipio-report@main

baton-percipio-report

baton resources
```

# Data Model

`baton-percipio-report` will pull down information about the following resources:

- Users

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually
building spreadsheets. We welcome contributions, and ideas, no matter how
small&mdash;our goal is to make identity and permissions sprawl less painful for
everyone. If you have questions, problems, or ideas: Please open a GitHub Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-percipio-report` Command Line Usage

```
baton-percipio-report

Usage:
  baton-percipio-report[flags]
  baton-percipio-report[command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
      --api-token string          required: The Percipio Bearer Token ($BATON_API_TOKEN)
      --client-id string          The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string      The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
  -f, --file string               The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                      help for baton-percipio-report
      --limited-courses strings   Limit imported courses to a specific list by Course ID ($BATON_LIMITED_COURSES)
      --log-format string         The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string          The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
      --organization-id string    required: The Percipio Organization ID ($BATON_ORGANIZATION_ID)
  -p, --provisioning              This must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --skip-full-sync            This must be set to skip a full sync ($BATON_SKIP_FULL_SYNC)
      --ticketing                 This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                   version for baton-percipio-report

Use "baton-percipio-report[command] --help" for more information about a command.
```
