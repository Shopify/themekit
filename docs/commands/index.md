---
layout: default
---
# Theme Kit Commands

## Get help from the command line
If you want this information while you are working you can get more help on each
command by running:

```bash
# Get a list of all commands and flags
> theme help
# Get a list of all the flags for a command
> theme [command] --help
```

## General Global Flags

|`-c` |`--config            `| path to config.yml
|`-d` |`--dir               `| directory that command will take effect. (default current directory)
|`-e` |`--env               `| environment to run the command
|`-h` |`--help              `| help for themekit
|`  ` |`--ignored-file      `| A single file to ignore, use the flag multiple times to add multiple.
|`  ` |`--ignores           `| A path to a file that contains ignore patterns.
|`  ` |`--no-ignore         `| Will disable config ignores so that all files can be changed
|`  ` |`--no-update-notifier`| Stop theme kit from notifying about updates.
|`-p` |`--password          `| theme password. This will override what is in your config.yml
|`  ` |`--proxy             `| proxy for all theme requests. This will override what is in your config.yml
|`-s` |`--store             `| your shopify domain. This will override what is in your config.yml
|`-t` |`--themeid           `| theme id. This will override what is in your config.yml
|`  ` |`--timeout           `| the timeout to kill any stalled processes. This will override what is in your config.yml
|`-v` |`--verbose           `| Enable more verbose output from the running command.

## Bootstrap

The bootstrap command has been renamed to `new`, please see the corresponding docs.

## Configure

Use this command to create or update configuration files. If you run the following
command:

```bash
theme configure --password=[your-api-password] --store=[your-store.myshopify.com] --themeid=[your-theme-id]
```

This will output a `config.yml` file in the current directory with the following contents

```yaml
development:
  password: [your-api-password]
  theme_id: "[your-theme-id]"
  store: [your-store].myshopify.com
  timeout: 30s
```

|**Required Flags**||
|`-p`|`--password`| Password for access to your Shopify account.
|`-s`|`--store   `| Your store's domain for changes to take effect
|`-t`|`--themeid `| The ID of the theme that you want changes to take effect

## Deploy
Deploy will completely replace what is on Shopify with what is in your current
project directory. This means that any files that are on Shopify but are not on
your local disk will be removed from Shopify. Any files that are both on your local
disk and Shopify will be updated. Lastly any files that are only on your local
disk will be upload to Shopify.

Deploy can be used without any filenames and it will replace the whole theme. If
some filenames are provided to replace then only those files will be replaced.

|**Optional Flags**||
|`-a`|`--allenvs`| Will run this command for each environment in your config file.
|`-n`|`--nodelete`| will run deploy without removing files from shopify.

## Download
If called without any arguments, it will download the entire theme, otherwise if
you specify the files you want to download, then only those files will be retrieved.
For example if you wanted to download the 404 and article liquid templates you
could enter the following command:

```bash
theme download templates/404.liquid templates/article.liquid
```

## Get
Get can be used to setup your theme on your local machine. It will both create
a config file and download the theme you request. If you have existing
themes on shopify, you can see what is available by running:

```
theme get --list -p=[your-api-password] -s=[your-store.myshopify.com]
```

Then once you have a theme id for the theme you want to setup on your local machine,
you can run:

```bash
theme get --password=[your-api-password] --store=[your-store.myshopify.com] --themeid=[your-theme-id]
```

This will output a `config.yml` file in the current directory with the following contents

```yaml
development:
  password: [your-api-password]
  theme_id: "[your-theme-id]"
  store: [your-store].myshopify.com
```

|**Required Flags**||
|`-p`|`--password`| Password for access to your Shopify account.
|`-s`|`--store   `| Your store's domain for changes to take effect
|**Optional Flags**||
|`-t`|`--themeid `| The ID of the theme that you want changes to take effect, if no theme id is passed, your live theme will be fetched

## new

If you are starting a new theme and want to have some sane defaults, you can use
the new command (formerly the bootstrap command). It will create a new theme based on Timber, update your
configuration file for that theme and download it to your computer. You will
need to provide your API password and your store domain to the command. You can
run the command like the following:

```bash
theme new --password=[your-api-password] --store=[your-store.myshopify.com]
```

**To get your credentials setup please refer to [the setup docs]({{ '/#get-api-access' | prepend: site.baseurl }})**

|**Required Flags**||
|`-p`|`--password`| Password for access to your Shopify account.
|`-s`|`--store   `| Your store's domain for changes to take effect
|**Optional Flags**||
|    |`--name   ` | a name to define your theme on your shopify admin
|    |`--prefix ` | prefix to the Timber theme being created
|    |`--url    ` | a url to pull a project theme zip file from.
|    |`--version` | version of Shopify Timber to use (default "latest")

## Open
Open will open the preview page for your theme in your browser as well as print
out the URL for your reference.

```bash
theme open --env=production # will open http://your-store.myshopify.com?preview_theme_id=<your-theme-id>
```

|**Optional Flags**||
|`-a`|`--allenvs`| run command with all environments
|`-b`|`--browser`| name of the browser to open the url, matching the name of browser on your system.
|`-E`|`--edit   `| open the web editor for the theme.

## Remove
Remove will delete theme files both locally and on Shopify. Unlike the other file
operation commands, this command requires filenames. This is done so that you cannot
accidentally delete your entire theme. For example if you wanted to delete the 404
and article liquid templates you could enter the following command:

```bash
theme remove templates/404.liquid templates/article.liquid
```

**Please note** you may not be able to remove some files from Shopify because they
are required to serve a valid theme.

|**Optional Flags**||
|`-a`|`--allenvs`| Will run this command for each environment in your config file.

## Replace
Replace has been renamed to `deploy` and has been deprecated, please see corresponding docs.

## Update
Update will update the Theme Kit command to the newest version. Update can also be
used to roll back to previous versions by providing it with a `--version` argument.
If there is a beta or prerelease version available you may also specify it with
the version flag, otherwise it will not be installed.

```bash
theme update --version=v0.5.0
```

|**Optional Flags**||
||`--version`  | Specifies what version Theme Kit should install.

## Upload
Upload has been renamed to `deploy` with the --nodelete flag and has been deprecated, please see corresponding docs.

## Version
Version will print out the current version of the library.

## Watch
Watch will start a process that will watch your directory for changes and
upload them to Shopify. Any changes will be logged to the terminal and the status
of the upload will be logged as well. The program can be stopped by simply typing
ctrl+C.

To ease integrating the watcher with tools such as LiveReload, you can provide
an the optional `--notify` argument with a file path  that you want to have updated
when the workers have gone idle. For example, if you had LiveReload watching for
update made to a file at /tmp/theme.update you would enter the following command:

```
theme watch --notify=/tmp/theme.update
```

|**Optional Flags**||
|`-a`|`--allenvs`| Will run this command for each environment in your config file.
|`-n`|`--notify` | File path to a file that you want updated on idle.
