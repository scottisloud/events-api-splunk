# Events Reporting for Splunk

This repository contains code to integrate Splunk with 1Password's Events API. It includes a Splunk Add-on that meets Splunk's AppInspect requirements, binary source code, and make commands to build the project.

This fork adds **multi-tenant** support: connect multiple 1Password accounts to one Splunk deployment, ingest into shared indexes, and filter events with a `tenant_id` field. See [docs/events-reporting-splunk-multi-tenant.md](docs/events-reporting-splunk-multi-tenant.md) for setup steps. The standard single-tenant flow is documented on the [1Password Support site](https://support.1password.com/events-reporting-splunk/).

## Directory Structure

The top level directory only contains two files, this `README.md` and a `Makefile` which has all the commands to build the Splunk Add-on for various distributions as well as build the linux specific version for running the application locally in docker.

### src/

This folder contains the go source and dependency code used in the Splunk Add-on. Changing this source code will not be reflected in the Splunk Add-on until you recompile the source. Use the `make compile_app_binary` to accomplish this.

### app/

This folder contains the necessary Splunk configuration files and compiled go source code. See this folder's `README.md` to learn about running the Splunk add-on locally.

### builds/

This folder will contain the OS specific Add-ons, compressed to Splunk's distribution requirements as well as installation steps.

## Requirements

### Go

If you do not have `go` locally installed, you can find installation steps [here](https://golang.org/doc/install).

## Commands

- `make compile_app_binary`
  This command will update the Splunk Add-on, found in `app`, with any changes made from `src`.

- `make new_version`
  This command will update the JS portion of the Splunk Add-on to `Makefile VERSION` and build a release bundle for the web app.

- `make build_all_apps` (or `make build`) compiles binaries for all platforms, bundles the Splunk app, and writes tarballs under `builds/bin`. This runs `build_all_binaries` automatically; you do not need to invoke it separately first.
- `make build_all_binaries` only cross-compiles Go binaries and runs the setup UI webpack build (creates `builds/bin`). Installs [gox](https://github.com/mitchellh/gox) to `$(go env GOPATH)/bin` if it is missing.
