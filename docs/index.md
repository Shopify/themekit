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

You will need to set up an API key to add to our configuration and create a connection
between your store and Theme Kit. The API key allows Theme Kit to talk to and access
your store, as well as its theme files.

To do so, log into the Shopify store, and create a private app. In the Shopify
Admin, go to Apps and click on `Manage private apps`. From there, click `Create a new private app`, to create your private app. Fill out the information at the top
and set the permissions of **Theme templates and theme assets** to have **_Read and write_**
access. Press `Save` and you will be presented with the next screen. In it you will
see your access credentials. Please copy the password. You will need it later.

<img src="{{ "/assets/images/shopify-local-theme-development-generate-api.gif" | prepend: site.baseurl }}" />

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
theme get --list -p=[your-password] -s=[your-store.myshopify.com]
```

Then once you have noted your theme ID, run the following command to generate a
config and download the theme from shopify:

```bash
theme get -p=[your-password] -s=[your-store.myshopify.com] -t=[your-theme-id]
```
