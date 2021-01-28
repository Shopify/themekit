<p align="center">
  <a href="https://shopify.dev/tools/theme-kit"><h3 align="center">Theme Kit</h3></a>
  <p align="center">Shopify Theme Manipulation CLI</p>
  <p align="center">
    <a href="https://goreportcard.com/report/github.com/shopify/themekit"><img src="https://goreportcard.com/badge/github.com/shopify/themekit"></a>
    <a href="https://travis-ci.com/Shopify/themekit"><img src="https://travis-ci.com/Shopify/themekit.svg?branch=master"></a>
    <a href="https://codecov.io/gh/Shopify/themekit"><img src="https://codecov.io/gh/Shopify/themekit/branch/master/graph/badge.svg" /></a>
    <a href="http://godoc.org/github.com/Shopify/themekit"><img src="https://godoc.org/github.com/Shopify/themekit?status.svg"></a>
    <a href="https://github.com/Shopify/themekit/releases/latest"><img src="https://img.shields.io/github/release/Shopify/themekit.svg"></a>
  </p>
</p>

---

Theme Kit is a cross-platform command line tool that you can use to build Shopify themes.

## Features

With Theme Kit, you can use your own development tools to interact with the Shopify platform in the following ways:

- Use workflow tools like Git to work with a team of theme developers.
- Upload themes to multiple environments.
- Watch for local changes and upload them automatically to Shopify.
- Work on Linux, macOS, and Windows.

## Install Theme Kit

You can install Theme Kit using the command line on the following operating systems:

- [Linux](https://shopify.dev/tools/theme-kit/getting-started#linux)
- [macOS](https://shopify.dev/tools/theme-kit/getting-started#macos)
- [Windows](https://shopify.dev/tools/theme-kit/getting-started#windows)

### Install Theme Kit manually

To manually install Theme Kit, download the file that matches your operating system (OS) and architecture. When the Theme Kit download finishes, complete the installation steps for [macOS and Linux](#macos-and-linux-steps) or [Windows](#windows-steps).

| OS | Architecture | md5 checksums | Download link |
|---|---|---|---|
| macOS | 64-bit | f85765e969256dec9a365112f230d37c | [download](https://shopify-themekit.s3.amazonaws.com/v1.1.5/darwin-amd64/theme) |
| Windows | 64-bit | fb45f717e502f376444ee44c65e04df6 | [download](https://shopify-themekit.s3.amazonaws.com/v1.1.5/windows-amd64/theme.exe) |
| Windows | 32-bit | ea53990984e61f774f2c52d390d84b0a | [download](https://shopify-themekit.s3.amazonaws.com/v1.1.5/windows-386/theme.exe) |
| Linux | 64-bit | b0b134b084c780a4054a3c47971351fb | [download](https://shopify-themekit.s3.amazonaws.com/v1.1.5/linux-amd64/theme) |
| Linux | 32-bit | ed7812adbacbc79f6d5d8ac4fc1e368f | [download](https://shopify-themekit.s3.amazonaws.com/v1.1.5/linux-386/theme) |
| FreeBSD | 64-bit | 8d043fe5116e09099a5821f5c6ce0200	| [download](https://shopify-themekit.s3.amazonaws.com/v1.1.5/freebsd-amd64/theme) |
| FreeBSD | 32-bit | 4743786ec3cb000a03ebf696d0e403c8 | [download](https://shopify-themekit.s3.amazonaws.com/v1.1.5/freebsd-386/theme) |

#### macOS and Linux steps

1. Compare checksums of the binary file by running `md5 theme`.
2. Set execute permissions with `chmod +x theme`. You might need to prefix with `sudo`.
3. Put the binary on your path. We recommend somewhere like `/usr/local/bin`.
4. Ensure that Theme Kit works as expected by running `theme version`.

#### Windows steps

1. Create a folder inside `C:\Program Files` called `Theme Kit`.
2. Copy the extracted program into `C:\Program Files\Theme Kit`.
3. Add `C:\Program Files\Theme Kit` to your `PATH` environment variable.
4. Open `cmd.exe` and type `theme` to verify that Theme Kit is installed.

### Node package

If you want to integrate Theme Kit into your build process, then you can run the following `npm` command to install the [Node wrapper](https://github.com/Shopify/node-themekit) around Theme Kit:

``` terminal
$ npm install themekit
```

### Community packages

Theme Kit is available through other package manager distributions. However, Shopify doesn't support or maintain these packages.

Community packages might not contain the latest Theme Kit release, but running `theme update` will [update Theme Kit to the latest version](https://shopify.dev/tools/theme-kit/troubleshooting/#update-theme-kit).

- [AUR](https://aur.archlinux.org/packages/shopify-themekit-bin) ([@rmcfadzean](https://github.com/rmcfadzean))

## Reference guides

- **[Theme Kit command reference](https://shopify.dev/tools/theme-kit/command-reference)** - Learn about the different commands that you can use in Theme Kit to execute key operations.
- **[Theme Kit configuration reference](https://shopify.dev/tools/theme-kit/configuration-reference)** - Familiarize yourself with the configuration variables available and the accepted values.

## Troubleshooting

Refer to [*Troubleshooting Theme Kit*](https://shopify.dev/tools/theme-kit/troubleshooting) to learn how to identify and resolve common issues in Theme Kit.

## Contributing to Theme Kit

Theme Kit is open source and you can help [contribute to the GitHub repository](https://github.com/Shopify/themekit/blob/master/CONTRIBUTING.md).

## Where to get help

- **[Open a GitHub issue](https://github.com/Shopify/themekit/issues)** - To report bugs or request new features, open an issue in the Theme Kit GitHub repository.
- **[Shopify Community Forums](https://community.shopify.com/)** - Visit our forums to connect with the community and learn more about Theme Kit development.
