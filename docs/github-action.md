---
layout: default
title: GitHub Action
nav_order: 9
description: Install and run httpprobe in your GitHub Actions workflows
---

# GitHub Action
{: .no_toc }

Install the httpprobe CLI on a GitHub Actions runner with a single step, then use it in any subsequent step.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Quick start

Add the setup action to any workflow:

```yaml
- uses: mrfoh/httpprobe/.github/actions/setup@v1.2.0
  with:
    version: latest
- run: httpprobe run -p ./tests
```

The action downloads the matching binary for the runner's OS and architecture, adds it to `PATH`, and exits. Subsequent steps can invoke `httpprobe` directly.

For reproducible builds, pin both the action ref and the `version` input to an explicit release:

```yaml
- uses: mrfoh/httpprobe/.github/actions/setup@v1.2.0
  with:
    version: v1.2.0
```

## Inputs

| Name           | Required | Default                | Description                                                         |
| -------------- | -------- | ---------------------- | ------------------------------------------------------------------- |
| `version`      | no       | `latest`               | Version to install — `latest`, `v1.2.0`, or `1.2.0`.                |
| `github-token` | no       | `${{ github.token }}`  | Token used to query the releases API. Avoids anonymous rate limits. |

## Outputs

| Name      | Description                                                  |
| --------- | ------------------------------------------------------------ |
| `version` | The resolved version that was installed (e.g. `v1.2.0`).     |

## Matrix testing

The action works on Linux, macOS, and Windows runners on both `amd64` and `arm64`:

```yaml
jobs:
  api-tests:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: mrfoh/httpprobe/.github/actions/setup@v1.2.0
      - run: httpprobe run -p ./tests
```

## Pinning to a major tag

If the repository maintains a floating major tag (e.g. `v1`), you can use it to auto-receive patch releases of the action:

```yaml
- uses: mrfoh/httpprobe/.github/actions/setup@v1
```

For production workflows where reproducibility matters, prefer pinning to a specific release tag instead.

## Supported platforms

| OS      | Architectures        |
| ------- | -------------------- |
| Linux   | amd64, arm64, 386    |
| macOS   | amd64, arm64         |
| Windows | amd64, arm64, 386    |

## Why no Marketplace listing?

GitHub Marketplace requires `action.yml` to live at the repo root, which would conflict with this repo's identity as the CLI project. Reference the action by its full subdirectory path (`mrfoh/httpprobe/.github/actions/setup@<ref>`) instead.
