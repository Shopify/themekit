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

## Using Environments

All of the following commands can be run with a selected environment from the configuration.
This allows multiple configurations to be held in the same config file.
Please see the [configuration documentation]({{ '/configuration/#config-file' | prepend: site.baseurl }}) for reference.

To use these environments you must specify them by name with the `--env` flag or `-e` for short.
The default environment is `development`. For example if you wanted to deploy to your production environment you could run

```bash
theme deploy --env=production
```

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

Theme Kit calculates a checksum for each file, and only updates assets if you've made changes to them locally.

|**Optional Flags**||
|    |`--allow-live`| Will allow Theme Kit to deploy the file changes to the currently live theme.
|`-a`|`--allenvs` | Will run this command for each environment in your config file.
|`-n`|`--nodelete`| will run deploy without removing files from shopify.

## Download
If called without any arguments, it will download the entire theme, otherwise if
you specify the files you want to download, then only those files will be retrieved.
For example if you wanted to download the 404 and article liquid templates you
could enter the following command:

```bash
theme download templates/404.liquid templates/article.liquid
```

Similar to the `deploy` command, Theme Kit will skip downloading any unchanged files.

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

**To get your credentials setup please refer to [the setup docs]({{ '/#get-api-access' | prepend: site.baseurl }})**

|**Required Flags**||
|`-p`|`--password`| Password for access to your Shopify account.
|`-s`|`--store   `| Your store's domain for changes to take effect
|`-t`|`--themeid `| The ID of the theme that you want changes to take effect.

## New

If you are starting a new theme and want to have some sane defaults, you can use
the new command. The command will
- Create a new theme on shopify with the provided name in the current directory you are in.
- Initialize your configuration with your credentials and your new theme id.
- Generate and upload some default templates to make your theme valid.

You will need to provide your API password and your store domain to the command. You can
run the command like the following:

```bash
theme new --password=[your-api-password] --store=[your-store.myshopify.com] --name="Dramatic Theme"
```

**To get your credentials setup please refer to [the setup docs]({{ '/#get-api-access' | prepend: site.baseurl }})**

|**Required Flags**||
|`-p`|`--password`| Password for access to your Shopify account.
|`-s`|`--store   `| Your store's domain for changes to take effect
|`-n`|`--name    `| The name of your new theme.

|**Optional Flags**||
|`--dir`| Directory to place all of the files for the new theme in. **Note:** Directory must exist beforehand.

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
|    |`--hidepb `| hide preview bar when opening the the preview page

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
|    |`--allow-live`| Will allow Theme Kit to remove files from a live theme

## Watch
Watch will start a process that will watch your directory for changes and
upload them to Shopify. Any changes will be logged to the terminal and the status
of the upload will be logged as well. The program can be stopped by simply typing
ctrl+C.

To ease integrating the watcher with tools such as LiveReload, you can provide
the optional `--notify` argument with a file path that you want to have updated
when the workers have gone idle. For example, if you had LiveReload watching for
update made to a file at /tmp/theme.update you would enter the following command:

```
theme watch --notify=/tmp/theme.update
```

|**Optional Flags**||
|`-a`|`--allenvs`| Will run this command for each environment in your config file.
|`-n`|`--notify` | Filepath or URL. Filepath is to a file that you want updated on idle. The URL path is where you want a webhook posted to to report on file changes.
|    |`--allow-live`| Will allow Theme Kit to make changes to the live theme

**Special Note**
Supplying the `--notify` flag with a URL will send a payload like the following. For a file change to the file `assets/app.js` the URL with receive a `POST` with a payload of:

```json
{
  "files": ["assets/app.js"]
}
```
