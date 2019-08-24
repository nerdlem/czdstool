[![GoDoc](https://godoc.org/github.com/nerdlem/czdstool?status.svg)](https://godoc.org/github.com/nerdlem/czdstool)
[![Go Report Card](https://goreportcard.com/badge/github.com/nerdlem/czdstool)](https://goreportcard.com/report/github.com/nerdlem/czdstool)
[![Build Status](https://travis-ci.org/nerdlem/czdstool.svg?branch=master)](https://travis-ci.org/nerdlem/czdstool)

# czdstool, command line interface to the ICANN CZDS REST API

[ICANN](https://icann.org) provides public access to thousands of TLD zone files through the [Centralized Zone Data System — CZDS](https://czds.icann.org) using an [HTTP REST API](https://github.com/icann/czds-api-client-java/blob/master/docs/ICANN_CZDS_api.pdf).

In order to use this tool, you'll need to access the CZDS system and complete your registration to get your access credentials.

## Installation

This should be a matter of

```
go get github.com/nerdlem/czdstool
go install github.com/nerdlem/czdstool
```

## Configuration

The shipped [sample configuration](https://github.com/nerdlem/czdstool/blob/master/czds.toml-example) only needs you to put your username and password. Keep that file safe — mind your file permissions — to prevent misuse.

## Logging in

The REST API uses JWT for authentication. JWTs are cryptographically signed _blobs_ that include authorization information. To prevent brute forcing attacks, the authentication API that provides the JWT tokens is subjected to aggressive rte limiting.

`czdstool` can store the JWTs in a local file, so you can authenticate once and keep issuing commands for up to 24 hours. To have `czdstool` authenticate itself and save the JWT to a file, issue the following command:

```
czdstool --config my-config.toml save ./my-token.jwt
```

If all went well, the file `my-token.jwt` should now contain a JWT token that allows you to access the CZDS via the REST API.

To use the saved JWT for authentication, remember to add the `--auth-file` command line option.

## Help

To get help about the tool, use either the `help` or `--help` commands as follows:

```
$ czdstool --help

Utility program to use the ICANN CZDS REST API to download authorized
TLD zone files.

Usage:
  czdstool [flags]
  czdstool [command]

Available Commands:
  fetch       Download TLD zone files using the ICANN CZDS REST API
  help        Help about any command
  ls          List available TLD zone URLs in the ICANN CZDS
  save        Persist authorization ICANN CZDS authorization token for later reuse

Flags:
  -A, --auth-file string   auth file previously created with save
      --config string      config file (default is /etc/czds.toml)
  -h, --help               help for czdstool
  -T, --tlds string        comma-separated list of TLDs to download
  -v, --verbose            verbose output format

Use "czdstool [command] --help" for more information about a command.
```

## Listing TLDs you have access to

Use the `ls` sub-command, as below. I'm also using the `-l` option to get some additional information about the TLD file:

```
$ czdstool ls --auth-file ./czds.ticket -l link
{"name":"link","last_modified":"2019-05-18T00:57:45Z","content_length":11766449,"content_type":"application/x-gzip"}
```

Issuing the `ls` commands with no additional arguments will simply list _all_ the TLDs you have available. This can be useful to keep track on access authorizations as they are revoked or expire automatically.

## Retrieving TLD zone files

This is done via the `fetch` command. The following example downloads a TLD file to the `./samples/` directory:

```
$ czdstool fetch --destination-dir ./samples/ --verbose sexy
Requesting auth token using credentials
Request is authorized
processing list of TLDs provided
looking up data for https://czds-api.icann.org/czds/downloads/sexy.zone
stat() ./samples//sexy.gz: stat ./samples//sexy.gz: no such file or directory
will fetch ./samples//sexy.gz via https://czds-api.icann.org/czds/downloads/sexy.zone
```

Since the `--auth-file` was not used, this command actually authenticated to the API and used a brand new JWT — which was discarded after use — for this transaction. The `--verbose` option causes additional output detailing what's happening.

If no TLDs are given to the `fetch` command, then all available TLD zones will be downloaded.

By default, only zone files that are older than the last-modified time provided by the REST API will be downloaded. The `--force` cancels this behavior and forces downloading the zone file regardless of local state.

As zone files are downloaded, their final size is compared with the length reported by the REST API, with an error being reported otherwise and the in-progress download being deleted. This behavior can be overridden with the `--keep-anyway` option.

Downloads happen atomically _ala_ `rsync`. Files are downloaded to a temporary location and, if the size check passes, are then renamed to the final intended name.

## Parallel access to the CZDS REST API

The option `--info-workers` and `--fetch-workers` control the number of concurrent `HEAD` and `GET` requests used to access the zone file data.

## Requesting access to tlds

The `request` command automatically requests access to the TLDs provided in the command line.
