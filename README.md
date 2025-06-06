![Baton Logo](./docs/images/baton-logo.png)

# `baton-percipio-report` [![Go Reference](https://pkg.go.dev/badge/github.com/iiiatthew/baton-percipio-report.svg)](https://pkg.go.dev/github.com/iiiatthew/baton-percipio-report)

`baton-percipio-report` is a connector for built using the [Baton SDK](https://github.com/conductorone/baton-sdk).

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

## Summary

This connector is a refactored version of the standard `baton-percipio` connector which only calls Percipio's Reporting Services endpoints to fetch a Learning Activity Report instead of making separate calls to the Content Discovery and User Management Services endpoints.

### The key differences are:

**Simplified Architecture**: Uses a single learning activity report to extract minimum required user and course data to retrieve course completion statuses instead of making multiple paginated API calls to separate endpoints.

**Testing Optimization**: Introduces `--lookback-days` and `--lookback-years` flags to control how far back to fetch learning activity data for testing purposes. The standard `baton-percipio` connector is coded to request 10 years of data. For development and testing, use `--lookback-days=1` or `--lookback-days=30` to generate reports much faster and speed up connector testing and validation.

**Production Configuration**: For production use, `--lookback-years=10` (default) provides comprehensive historical data for compliance and audit purposes.

## Building the Connector Binary

The repo includes a `Makefile` for building, adding and updating dependencies, and linting

```bash
make build
```

If you don't want to use `make` you can build the binary with:

```bash
go build -o dist/baton-percipio-report ./cmd/baton-percipio-report
```

## Running the Connector

### Running in Local Test Mode (One-Shot)

Running the connector without ConductorOne credentials will trigger one-shot mode which generates a c1z file to use for testing and manual upload. By default this mode will create a sync.c1z file, in the current directory, that can be examined using the [Baton Toolkit](https://github.com/conductorone/baton).

#### Generating the .c1z file

```bash
baton-percipio-report \
  --api-token <PERCIPIO_API_TOKEN> \
  --organization-id <PERCIPIO_ORG_ID> \
  --lookback-days 1 \
  --log-level debug
```

#### Validating the generated c1z file

Requires the [Baton Toolkit](https://github.com/conductorone/baton) to be installed on your local machine.

```bash
baton resources -f sync.c1z
```

```bash
baton grants -f sync.c1z
```

```bash
baton explorer -f sync.c1z
```

### Running in Service Mode / Continuous Sync Mode (Production)

Once you are satified that the local testing mode execution generates the expected resources, entitlements, and grants as expected, you can run the connector in service mode. Service mode runs as a continuous process which ConductorOne calls for sync operations once per hour. To run the connector in service mode, pass your ConductorOne connector credentials as commandline flags, this will automatically trigger service mode.

Run as a service for ConductorOne:

```bash
baton-percipio-report \
  --api-token <PERCIPIO_API_TOKEN> \
  --organization-id <PERCIPIO_ORG_ID> \
  --client-id <BATON_CLIENT_ID> \
  --client-secret <BATON_CLIENT_SECRET> \
```

# `baton-percipio-report` Command Line Usage

```
baton-percipio-report

Usage:
  baton-percipio-report [flags]
  baton-percipio-report [command]

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
  -d, --lookback-days int         How many days back of learning activity data to fetch ($BATON_LOOKBACK_DAYS)
  -y, --lookback-years int        How many years back of learning activity data to fetch (default is 10 years) ($BATON_LOOKBACK_YEARS)
      --log-format string         The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string          The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
      --organization-id string    required: The Percipio Organization ID ($BATON_ORGANIZATION_ID)
  -v, --version                   version for baton-percipio-report

Use "baton-percipio-report [command] --help" for more information about a command.
```
