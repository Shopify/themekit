<p align="center">
  <a href="https://shopify.github.io/themekit/"><h3 align="center">Theme Kit</h3></a>
  <p align="center">Shopify Theme Manipulation CLI</p>
  <p align="center">
    <a href="https://goreportcard.com/report/github.com/shopify/themekit"><img src="https://goreportcard.com/badge/github.com/shopify/themekit"></a>
    <a href="https://circleci.com/gh/Shopify/themekit"><img src="https://circleci.com/gh/Shopify/themekit.png?circle-token=ac951910873cafaaf9c1be6049d2b9d3276eb2d4"></a>
    <a href="http://godoc.org/github.com/Shopify/themekit"><img src="https://godoc.org/github.com/Shopify/themekit?status.svg"></a>
    <a href="https://github.com/Shopify/themekit/releases/latest"><img src="http://github-release-version.herokuapp.com/github/Shopify/themekit/release.svg?style=flat"></a>
  </p>
</p>

---

Theme Kit is a cross-platform tool for building Shopify Themes. Theme Kit is a single binary that has no dependencies.

## Features
- Upload Themes to Multiple Environments
- Fast Uploads and Downloads
- Watch for local changes and upload automatically to Shopify
- Prevent overwriting changes made in the online editor.
- Works on Windows, Linux and macOS

## Installation

You can find installation instructions for each platform on the [Docs Website](https://shopify.github.io/themekit/#installation)

## Setup, Usage and Commands

Please find further usage instructions on the [theme kit website](https://shopify.github.io/themekit/)

## Contribution & Development

Please see the [contributing guidlines](https://github.com/Shopify/themekit/blob/master/.github/CONTRIBUTING.md)

## Authors

[Tim Anema](https://github.com/tanema), [Chris Saunders](https://github.com/csaunders),
[Chris Butcher](https://github.com/chrisbutcher), [Jakob Külzer](https://github.com/ilikeorangutans)

refactor todo
- [x] testing packages 10/10
- [x] fixing commands tests 8/8
- [x] add back in warnings 2/2
- [ ] better credentials errors
  - [ ] get shop info to verify shop domain
  - [ ] get all themes to verify password
  - [ ] with all themes, set id of live theme in running config so that actions are more explicit
- [ ] better timeout errors
  - net/http: request canceled (Client.Timeout exceeded while awaiting headers)
  - if this error is experienced, recommend increasing timeout
  - possibly add a timeout value to their config
- [ ] better performance
  - [ ] replace should use a channel to stream assets and not load them all at once
  - [ ] upload should stream assets in a channel and not load all assets at once
- [ ] Better onboarding
  - [ ] instead of bootstrap/config have new command that bootstraps or downloads

