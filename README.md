# [Theme Kit](https://shopify.github.io/themekit/) [![Go Report Card](https://goreportcard.com/badge/github.com/shopify/themekit)](https://goreportcard.com/report/github.com/shopify/themekit) [![Build Status](https://circleci.com/gh/Shopify/themekit.png?circle-token=ac951910873cafaaf9c1be6049d2b9d3276eb2d4)](https://circleci.com/gh/Shopify/themekit)[![GoDoc](https://godoc.org/github.com/Shopify/themekit?status.svg)](http://godoc.org/github.com/Shopify/themekit)
## Shopify Theme Manipulation CLI

Theme Kit is a cross-platform tool for building Shopify Themes. Theme Kit is a single binary that has no dependencies.

**Features**
- Upload Themes to Multiple Environments
- Fast Uploads and Downloads
- Watch for local changes and upload automatically to Shopify
- Works on Windows, Linux and macOS

[Read more about it on the website](https://shopify.github.io/themekit/)

## Installation

You can find installation instructions for each platform on the [Docs Website](https://shopify.github.io/themekit/#installation)

## Setup, Usage and Commands

Please find further usage instructions on the [theme kit website](https://shopify.github.io/themekit/)

# Development

Themekit requires go 1.8. You can setup your development environment by running
the following:

```bash
go get -u github.com/Shopify/themekit
cd $GOPATH/src/github.com/Shopify/themekit
make [mac_tools|linux_tools] # install platform specific development tools
glide install # see https://github.com/Masterminds/glide for glide usage
make # build themekit
theme version #should output the current themekit version
```

This will install theme kit into `$GOPATH/bin` and you will now have access to
the theme command.

## Contribution Guidelines

We welcome your contributions to the project. There are a few steps to take when
looking to make a contribution.

- Open an issue to discuss the feature/bug
- If feature/bug is deemed valid then fork repo.
- Implement patch to resolve issue.
- Include tests to prevent regressions and validate the patch.
- Update the docs for any API changes.
- Submit pull request and mention maintainers. Current maintainers are @tanema, @chrisbutcher

## Authors

[Chris Saunders](https://github.com/csaunders), [Tim Anema](https://github.com/tanema),
[Chris Butcher](https://github.com/chrisbutcher), [Jakob Külzer](https://github.com/ilikeorangutans)
