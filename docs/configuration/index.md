---
layout: default
---
# Theme Kit Configuration

There are 3 ways to do configurations and they have precedence of order. The config
file values can be overridden by environment variable and environment variables
can be overridden by command line flags. Please keep this in mind while trying
to debug your config.

There are general values that you will be able to config for all of Theme Kit actions.

| Attribute    | Description
|:-------------|:---------------------
| config       | A custom path to your config file. This defaults to the current path that the command is run in.
| store        | Your store's Shopify domain with the `.myshopify.com` postfix. Please see the [setup docs]({{ '/#get-api-access' | prepend: site.baseurl }}) on how to get this value.
| password     | Your API password. Please see the [setup docs]({{ '/#get-api-access' | prepend: site.baseurl }}) on how to get this value.
| theme_id     | The theme that you want the command to take effect on. Please see the [setup docs]({{ '/#get-api-access' | prepend: site.baseurl }}) on how to get this value.
| directory    | The project root directory. This allows you to run the command from another directory.
| ignore_files | A list of patterns to ignore when executing commands. Please see the [Ignore Patterns]({{ '/ignores' | prepend: site.baseurl }})  documentation.
| ignores      | A list of file paths to files that contain ignore patterns. Please see the [Ignore Patterns]({{ '/ignores' | prepend: site.baseurl }})  documentation.
| proxy        | A full URL to proxy your requests through. The URL only supports the `http` protocol.
| timeout      | Request timeout. If you have larger files in your project that may take longer than the default 30s to upload, you may want to increase this value. You can set this value to 60s for seconds or 1m for one minute.
| readonly     | All actions are readonly. This means you can download from this environment but you cannot do any modifications to the theme on shopify.

## Config File

Your configuration will be setup per environment. Here is an example config file
for you to understand the usage of environments and config.

```yaml
development:
  password: 16ef663594568325d64408ebcdeef528
  theme_id: "123"
  store: can-i-buy-a-feeling.myshopify.com
  proxy: http://localhost:3000
  ignore_files:
    - "*.gif"
    - "*.jpg"
    - config/settings_data.json
production:
  password: 16ef663594568325d64408ebcdeef528
  theme_id: "456"
  store: can-i-buy-a-feeling.myshopify.com
  timeout: 60s
  readonly: true
test:
  password: 16ef663594568325d64408ebcdeef528
  theme_id: "789"
  store: can-i-buy-a-feeling.myshopify.com
  ignores: ignore.txt
```

## Environment Variables

It is prudent to not store your private secrets in your repository so you can set
these values in your environment and there are a couple of ways to do this.

### Variable interpolation into config.yml
You can also interpolate variables into your config.yml file using `${}` notation.
An example of this looks like this:

```yaml
development:
  password: ${DEV_PASSWD}
  theme_id: ${DEV_THEMEID}
  store: ${DEV_SHOP}
```

To help facilitate this as well there are special files that can be used to automatically
load environment variables for themekit. These filepaths are:

| Platform   | Path |
| :--------- | :--- |
| Windows    | `%APPDATA%\Shopify\Themekit\variables`
| Linux/BSDs | `${XDG_CONFIG_HOME}/Shopify/Themekit/variables`
| MacOSX     | `${HOME}/Library/Application Support/Shopify/Themekit/variables`
| Any        | `--vars` flag provides a path to a file for loading variables

The variables file has the same format as most .env type files. For our example
config above, our variables file would look like this:

```
DEV_PASSWD=0bwef09hn23048sdkl2345n2k3
DEV_THEMEID=123
DEV_SHOP=can-i-buy-a-feeling.myshopify.com
```

This allows your to commit your config.yml to your repo but keep your secrets
out of the repo.


### Flag type environment variables
Most of the global flags also have a corresponding environment variable. All of
the variables are prefixed with `THEMEKIT_`

| Attribute    | Environment Variable |
|:-------------|:---------------------|:------------------|
| password     | THEMEKIT_PASSWORD    |                   |
| theme_id     | THEMEKIT_THEME_ID    |                   |
| store        | THEMEKIT_STORE       |                   |
| directory    | THEMEKIT_DIRECTORY   |                   |
| ignore_files | THEMEKIT_IGNORE_FILES| Use a ':' as a pattern separator.  |
| ignores      | THEMEKIT_IGNORES     | Use a ':' as a file path separator. |
| proxy        | THEMEKIT_PROXY       |                   |
| timeout      | THEMEKIT_TIMEOUT     |                   |

**Note** Any environment variable will take precedence over your `config.yml` values
so please keep that in mind while debugging your config.
Environment variables are platform dependant. For more information on environment
variables please see [this link for macOS and linux](https://www.cyberciti.biz/faq/set-environment-variable-linux/)
and [for windows please see this link](http://www.computerhope.com/issues/ch000549.htm)

## Flags

You can enforce any setting manually using a command line flag. This is useful for
debugging settings or scripting calls to Theme Kit from something like `cron`.

| Attribute    | Flag            | Shortcut
|:-------------|:----------------|:--------|
| config       | `--config`      | -c      |
| password     | `--password`    |         |
| theme_id     | `--themeid`     |         |
| store        | `--store`       | -s      |
| directory    | `--dir`         | -d      |
| ignore_files | `--ignored-file`|         |
| ignores      | `--ignores`     |         |
| proxy        | `--proxy`       |         |
| timeout      | `--timeout`     |         |

**Note** Any flag will take precedence over your `config.yml` and environment values
so please keep that in mind while debugging your config.
