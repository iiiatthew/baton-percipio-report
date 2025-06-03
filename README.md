![Baton Logo](./docs/images/baton-logo.png)

# `baton-percipio-report` [![Go Reference](https://pkg.go.dev/badge/github.com/iiiatthew/baton-percipio-report.svg)](https://pkg.go.dev/github.com/iiiatthew/baton-percipio-report) ![main ci](https://github.com/iiiatthew/baton-percipio-report/actions/workflows/main.yaml/badge.svg)

`baton-percipio-report` is a connector for built using the [Baton SDK](https://github.com/conductorone/baton-sdk).

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

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
      --log-format string         The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string          The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
      --organization-id string    required: The Percipio Organization ID ($BATON_ORGANIZATION_ID)
  -v, --version                   version for baton-percipio-report

Use "baton-percipio-report[command] --help" for more information about a command.
```
