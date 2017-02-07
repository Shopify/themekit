---
layout: default
---
# Theme Kit Commands

If you want this information while you are working you can get more help on each
command by running:

```bash
theme [command] --help
```

## General Global Flags

- `--env`, -e : specify an environment to run your command. The default environment is `development`.
- `--config`, -c : specify a path for your config file.
- `--no-update-notifier` : this will supress the update notifier that lets the
  user know when there is an applicable update available.
- `--verbose, -v` : Enable more verbose output from the running command. If this is not provided a progress bar will be displayed instead.

## Bootstrap

If you are starting a new theme and want to have some sane defaults, you can use
the bootstrap command. It will create a new theme based on Timber, update your
configuration file for that theme and download it to your computer. You will
need to provide your API password and your store domain to the command. You can
run the command like the following:

```bash
theme bootstrap --password=[your-api-password] --store=[your-store.myshopify.com]
```
**Required Flags**

* `--password` : Password for access to your Shopify account.
* `--store` : Your store's domain for changes to take effect

**Optional Flags**

* `--version` : With the version flag you can specify the version of Timber to use.
* `--prefix` : You can add a prefix to the theme name on Shopify. (i.e. `--prefix=mine`
  will create a theme called `mine-Timber-latest`)
* `--url` : You can provide a url to a zip file of a theme that you have access to.
  This allows you to create an initial theme from your own template.
* `--name` : Defines the name of the theme on the shopify admin. This flag is recommended
  because otherwise the theme name will be pretty generic.

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

**Required Flags**

* `--password` : Password for access to your Shopify account.
* `--store` : Your store's domain for changes to take effect
* `--themeid` : The ID of the theme that you want changes to take effect

## Download
If called without any arguments, it will download the entire theme, otherwise if
you specify the files you want to download, then only those files will be retrieved.
For example if you wanted to download the 404 and article liquid templates you
could enter the following command:

```bash
theme download templates/404.liquid templates/article.liquid
```

## Open
Open will open the preview page for your theme in your browser as well as print
out the URL for your reference.

```bash
theme open --env=production # will open http://your-store.myshopify.com?preview_theme_id=<your-theme-id>
```

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

**Optional Flags**

* `--allenvs` : Will run this command for each environment in your config file.

## Replace
Replace will completely replace what is on Shopify with what is in your current
project directory. This means that any files that are on Shopify but are not on
your local disk will be removed from Shopify. Any files that are both on your local
disk and Shopify will be updated. Lastly any files that are only on your local
disk will be upload to Shopify.

Replace can be used without any filenames and it will replace the whole theme. If
some filenames are provided to replace then only those files will be replaced.

**Optional Flags**

* `--allenvs` : Will run this command for each environment in your config file.

## Update
Update will update the Theme Kit command to the newest version. Update can also be
used to roll back to previous versions by providing it with a `--version` argument.
If there is a beta or prerelease version available you may also specify it with
the version flag, otherwise it will not be installed.

```bash
theme update --version=v0.5.0
```

**Optional Flags**

* `--version` : Specifies what version Theme Kit should install.

## Upload
Upload will update all file states to Shopify and create any files that are not
on Shopify's servers. If upload is not provided with filename arguments it will do
the whole project, otherwise you can upload individual files like this:

```bash
theme upload templates/404.liquid templates/article.liquid
```
**Optional Flags**

* `--allenvs` : Will run this command for each environment in your config file.

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

**Optional Flags**

* `--allenvs` : Will run this command for each environment in your config file.
* `--notify` : File path to a file that you want updated on idle.
