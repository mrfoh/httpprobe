# Setup httpprobe

A GitHub Action that installs the [httpprobe](https://github.com/mrfoh/httpprobe) CLI on a runner and adds it to `PATH`, so subsequent steps can run `httpprobe`.

## Usage

```yaml
- uses: mrfoh/httpprobe/.github/actions/setup@v1.2.0
  with:
    version: latest
- run: httpprobe run -p ./tests
```

Pinning to a specific version is recommended for reproducible builds:

```yaml
- uses: mrfoh/httpprobe/.github/actions/setup@v1.2.0
  with:
    version: v1.2.0
```

## Inputs

| Name           | Required | Default                | Description                                                              |
| -------------- | -------- | ---------------------- | ------------------------------------------------------------------------ |
| `version`      | no       | `latest`               | Version to install — `latest`, `v1.2.0`, or `1.2.0`.                     |
| `github-token` | no       | `${{ github.token }}`  | Token used to query the releases API. Avoids anonymous rate limits.      |

## Outputs

| Name      | Description                                                |
| --------- | ---------------------------------------------------------- |
| `version` | The resolved version that was installed (e.g. `v1.2.0`).   |

## Supported runners

Linux, macOS, and Windows GitHub-hosted runners on `amd64` and `arm64`. The action also maps `X86` runners to the 386 binary if you use self-hosted 32-bit runners.

## Marketplace

This action is intentionally not published to the GitHub Marketplace — Marketplace requires `action.yml` at the repo root, which would conflict with this repo's primary purpose as the CLI itself. Reference the action by its full subdirectory path as shown above.
