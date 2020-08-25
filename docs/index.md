---
layout: default
---

# Shopify Theme Kit

Theme Kit is a command line tool for shopify themes. Download the application
and with a tiny bit of setup you’re off to the theme creation races.

Using Theme Kit will enable you to

- Upload Themes to Multiple Environments
- Fast Uploads and Downloads
- Watch for local changes and upload automatically to Shopify
- Works on Windows, Linux and macOS

## Installation

### macOS Installation

Use [homebrew](http://brew.sh/) to install Theme Kit by running the following commands.

```bash
brew tap shopify/shopify
brew install themekit
```

### Windows Chocolatey Installation

If you have [chocolatey](https://chocolatey.org/) installed you can install themekit by running the following commands.

```
choco install themekit
```

### Linux Installation

If you are on linux based system, you can use the following installation script to automatically
download and install the latest Theme Kit for you.

```bash
curl -s https://shopify.github.io/themekit/scripts/install.py | sudo python
```

### Arch installation

There is a package available for install on the [AUR](https://aur.archlinux.org/packages/shopify-themekit-bin)

### Manual Installation

| OS      | Architecture | md5 checksums              |                                                                                                          |
| :------ | :----------- | :------------------------- | :------------------------------------------------------------------------------------------------------- |
| macOS   | 64-bit       | {{ site.darwinamd64sum }}  | [download](https://shopify-themekit.s3.amazonaws.com/{{ site.themekitversion }}/darwin-amd64/theme)      |
| Windows | 64-bit       | {{ site.windowsamd64sum }} | [download](https://shopify-themekit.s3.amazonaws.com/{{ site.themekitversion }}/windows-amd64/theme.exe) |
| Windows | 32-bit       | {{ site.windows386sum }}   | [download](https://shopify-themekit.s3.amazonaws.com/{{ site.themekitversion }}/windows-386/theme.exe)   |
| Linux   | 64-bit       | {{ site.linuxamd64sum }}   | [download](https://shopify-themekit.s3.amazonaws.com/{{ site.themekitversion }}/linux-amd64/theme)       |
| Linux   | 32-bit       | {{ site.linux386sum }}     | [download](https://shopify-themekit.s3.amazonaws.com/{{ site.themekitversion }}/linux-386/theme)         |
| FreeBSD | 64-bit       | {{ site.freebsdamd64sum }} | [download](https://shopify-themekit.s3.amazonaws.com/{{ site.themekitversion }}/freebsd-amd64/theme)     |
| FreeBSD | 32-bit       | {{ site.freebsd386sum }}   | [download](https://shopify-themekit.s3.amazonaws.com/{{ site.themekitversion }}/freebsd-386/theme)       |

#### Linux & macOS

- Download the themekit binary that works for your system.
- Compare checksums of the binary by running `md5 theme`
- Put the binary on your path. We recommend somewhere like `/usr/local/bin`
- Ensure that it works as expected by running `theme version`

#### Windows Installation

- Create a folder inside `C:\Program Files` called `Theme Kit`
- Download themekit (below) and copy the extracted program into `C:\Program Files\Theme Kit`
- You will then need to add `C:\Program Files\Theme Kit` to your `PATH` environment variable. You can find really [in-depth instructions here](https://helpdeskgeek.com/windows-10/add-windows-path-environment-variable/)
- To verify that Theme Kit has been installed, open `cmd.exe` and type in `theme`.

#### Build from source

Please refer to the [contributing docs](https://github.com/Shopify/themekit/blob/master/.github/CONTRIBUTING.md#developing-themekit)

## Get API Access

You'll need to set up new Shopify API credentials to connect Theme Kit to your store and manage your template files. Theme Kit manages its connection using a [private app](https://shopify.dev/concepts/apps#private-apps).

#### Steps:

1. In the store's Shopify admin, click **Apps**.
1. Near the bottom of the page, click on **Manage private apps**.
1. If you see a notice that "Private app development is disabled", then click "Enable private app development".
1. Click **Create new private app**.
1. In the **App details** section, fill out the app name and your email address.
1. In the **Admin API** section, click **Show inactive Admin API permissions**.
1. Scroll to the "Themes" section and select **Read and write** from the dropdown.
1. Click **Save**.
1. Read the private app confirmation dialog, then click **Create app**.

You'll return to the app detail page. Your new, unique access credentials are visible in the **Admin API** section. Copy the password. You'll use it in the next step.

<video src="https://screenshot.click/themekit-private-app-setup-1000p15-192kbps.mp4" style="max-width: 100%" loop autoplay>Sorry, your browser doesn't support embedded video.</video>

## Create a new theme.

If you are starting from scratch and want to get a quick start, run the following:

```bash
theme new --password=[your-password] --store=[your-store.myshopify.com] --name=[theme name]
```

This will:

- generate a basic theme template locally
- create a theme on shopify
- upload the new files to shopify
- Create/update your config file with the configuration for your new theme.

## Configure an existing theme.

To connect an existing theme, you need the theme’s ID number. The easiest way to
get your theme’s ID number is to use the get command like this:

```bash
theme get --list -p=[your-password] -s=[you-store.myshopify.com]
```

Then once you have noted your theme ID, run the following command to generate a
config and download the theme from shopify:

```bash
theme get -p=[your-password] -s=[you-store.myshopify.com] -t=[your-theme-id]
```
