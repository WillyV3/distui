package workflow

const workflowTemplate = `name: Release

on:
  push:
    tags: ['v*']
  workflow_dispatch:

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

{{- if .IncludeTests}}
      - name: Run tests
        run: go test ./...
{{- end}}

      - uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{"{{"}} secrets.GITHUB_TOKEN {{"}}"}}
{{- if .NPMEnabled}}
          NPM_TOKEN: ${{"{{"}} secrets.NPM_TOKEN {{"}}"}}
{{- end}}
`
