# Install Theme Kit manually

> Note:
> The recommended way to [install Theme Kit](https://shopify.dev/tools/theme-kit/getting-started) is by using the command line. You should only complete the following steps if you know what you're doing.

To manually install Theme Kit, [download the file](https://github.com/Shopify/themekit/releases/latest) that matches your operating system (OS) and architecture. When the Theme Kit download finishes, complete the installation steps for macOS and Linux or Windows.

## macOS and Linux steps

1. Compare checksums of the binary file by running `md5 theme`.
2. Set execute permissions with `chmod +x theme`. You might need to prefix with `sudo`.
3. Put the binary on your path. We recommend somewhere like `/usr/local/bin`.
4. Ensure that Theme Kit works as expected by running `theme version`.

## Windows steps

1. Create a folder inside `C:\Program Files` called `Theme Kit`.
2. Copy the extracted program into `C:\Program Files\Theme Kit`.
3. Add `C:\Program Files\Theme Kit` to your `PATH` environment variable.
4. Open `cmd.exe` and type `theme` to verify that Theme Kit is installed.

## Community packages

Theme Kit is available through other package manager distributions. However, Shopify doesn't support or maintain these packages.

Community packages might not contain the latest Theme Kit release, but running `theme update` will [update Theme Kit to the latest version](https://shopify.dev/tools/theme-kit/troubleshooting/#update-theme-kit).

- [AUR](https://aur.archlinux.org/packages/shopify-themekit-bin) ([@rmcfadzean](https://github.com/rmcfadzean))
