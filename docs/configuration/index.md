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
| password     | Your API password
| theme_id     | The theme that you want the command to take effect on. If you want to make changes to the current live theme you may set this value to `'live'`
| store        | Your store's Shopify domain with the `.myshopify.com` postfix.
| directory    | The project root directory. This allows you to run the command from another directory.
| ignore_files | A list of patterns to ignore when executing commands. Please see the [Ignore Patterns]({{ '/ignores' | prepend: site.baseurl }})  documentation.
| ignores      | A list of file paths to files that contain ignore patterns. Please see the [Ignore Patterns]({{ '/ignores' | prepend: site.baseurl }})  documentation.
| proxy        | A full URL to proxy your requests through.
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
these values in your environment.

All of the Theme Kit environment variables are prefixed with `THEMEKIT_`

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

## Global Flags

You can enforce any setting manually using a command line flag. This is useful for
debugging settings or scripting calls to Theme Kit from something like `cron`.

| Attribute    | Flag            | Shortcut
|:-------------|:----------------|:--------|
| password     | `--password`    |         |
| theme_id     | `--themeid`     |         |
| store        | `--store`       | -s      |
| directory    | `--directory`   | -d      |
| ignore_files | `--ignored-file`|         |
| ignores      | `--ignores`     |         |
| proxy        | `--proxy`       |         |
| timeout      | `--timeout`     |         |

**Note** Any flag will take precedence over your `config.yml` and environment values
so please keep that in mind while debugging your config.
