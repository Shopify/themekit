---
title: Documentation
---

# Documentation

1. [Overview](#overview)
2. [Commands](#commands)
3. [The Configuration File](#config-file)

<%= @config[:project_name] %> comes with a number of utilities that you can use to interact with a theme on Shopify. To see the list
of commands that are available, enter the following command into your terminal:

`theme --help`

You will see output similar to the following:

<pre><code>
#!bash
Usage of theme:
  -command="download": An operation to be performed against the theme.
  Valid commands are:
    remove <file> [<file2> ...]:
        Remove file(s) from theme
    replace [<file> ...]:
        Overwrite theme file(s)
    watch:
        Watch directory for changes and update remote theme
    configure:
        Create a configuration file
    bootstrap:
        Bootstrap a new theme using Shopify Timber
    upload <file> [<file2> ...]:
        Add file(s) to theme
    download [<file> ...]:
        Download file(s) from theme [default]
</code></pre>

<a name="commands"></a>

## Commands

0. [Common Arguments](#common-args)
1. [configure](#configure)
2. [bootstrap](#bootstrap)
3. [watch](#watch)
4. [download](#download)
5. [remove](#remove)
6. [replace](#replace)
7. [upload](#upload)


-----------------------------------

<a name="common-args"></a>

### Common Arguments

There are some arguments that you can pass into every command:

1. `env`: The *environment* to register the configuration settings under. These are just free form text, so you can name it anything. Common names are `staging`, `production`, `test` and the default is `development`
2. `dir`: The directory that the configuration file will live. This allows you to update multiple themes without having to change into each themes directory.

<a name="configure"></a>

### Configure

Use this command to create or update configuration files. If you don't want to run the command in the directory that will contain the configuration file you will need to specify the directory manually.

The available options are follows:

1. [required] `access_token`: Your Shopify API Access Token. This is needed to make authenticated calls against the Shopify API. Create a Private Application and **use the value from the Password field**
2. [required] `domain`: Your `.myshopify.com` domain without any protocol or other information.
3. `bucketSize`: Shopify uses a leaky bucket strategy for rate limiting API access. If you have been granted additional API usage you can update that here. Internally it is used to prevent failures that can be caused by exhausted API limits.
4. `refillRate`: The rate at which tickets are restored in the leaky bucket (per second)

<a name="bootstrap"></a>

### Bootstrap

If you are starting a new theme and want to have some sane defaults, you can use the bootstrap command. It will create a new theme based on [Timber](https://github.com/shopify/timber), update your configuration file for that theme and download it to your computer.

**Configuration is required before bootstrapping**

The available options are as follows:

1. `setid`: When setid is true your configuration file will be updated to include the appropriate Theme ID. This will ensure that changes are uploaded to the correct place. You can opt out of this by passing setting the flag to false (i.e. `--setid=false`)
2. `prefix`: Prefix to add to your theme. If a prefix is provided, the theme created on Shopify will be called `PREFIX-Timber-VERSION`, otherwise it will simply be `Timber-VERSION`
3. `version`: The version of Timber that you want to use for developing a theme. By default it will use the [latest stable release](https://github.com/Shopify/Timber/releases). Latest master can also be used by setting the version flag (i.e. `--version=master`)


<a name="watch"></a>

### Watch

Theme watch will start a process that will watch the specified directory for changes and upload them to the specified Shopify theme. Any changes will be logged to the terminal and the status of the upload will be logged as well. The program can be stopped by simply typing `ctrl+C`.

By default watch starts up two workers who will perform the upload, allowing for faster uploads to Shopify. The number of workers can be changed in your [configuration file](#config-files) though be aware that you may run the risk of slowdowns due to the leaky bucket getting drained.

<a name="download"></a>

### Download

If called without any arguments, it will download the entire theme. Otherwise if you specify the files you want to download, then only those files will be retrieved. For example if you wanted to download the `404` and `article` liquid templates you could enter the following command:

<pre><code>theme download templates/404.liquid templates/article.liquid`</code></pre>

<a name="remove"></a>

### Remove

Deletes theme files both locally and on Shopify. This command must be called with at least one filename.

<a name="replace"></a>

### Replace

Removes remote files and replaces them with local files. Equivalent to upload, so this may go away or change in the future. This command requires at least one filename.

<a name="upload"></a>

### Upload

Upload specified files to Shopify theme. This command requires at least one filename.

-----------------------

<a name="config-file"></a>

## The Configuration File

### Environments

Environments allow you to manage where to upload your theme changes to. This helps reduce the errors that could happen when swapping between environments by directly modifying the config file. Environments are named whatever you want and will all contain roughly the same information, though perhaps with some minor changes (store, theme id, etc.)

Environments all live within `config.yml` and a basic configuration could look like this:

<pre><code>
#!yaml
production:
  theme_id:
  access_token: abracadabra
  store: pokeshop.myshopify.com
development:
  theme_id: 123
  access_token: abracadabra
  store: pokeshop.myshopify.com
</code></pre>

### Configuration Variables

A configuration can contain a number of things on top of what can be added by invoking `theme configure`.

- `theme_id`: The ID of the theme to upload changes to. **Beware** that if this is blank all changes will be uploaded to your user visible theme.
- `access_token`: API credentials to update and manipulate themes on Shopify
-  `store`: Your `.myshopify.com` domain (i.e. pokeshop.myshopify.com). Note that you **do not** need to include `http://` or `https://`
- `ignore_files`: ???
- `ignores`: ???
- `bucket_size`: The size of your Access Tokens bucket. By default this is 40 and cannot be configured by the shop.
- `refill_rate`: The rate at which tickets are replenished by the Shopify API (per second). By default this is 2 and cannot be configured by the shop.
- `concurrency`: The number of workers to spawn when running `theme watch`. By default it is 2 to ensure that API limits don't get hit. This can be changed, though keep in mind that it can cause slowdowns if API limits are indeed reached. For most regular usage this might not be the case, but could happen if using automated that create a lot of File System events.
- `proxy`: Proxy server to route HTTP requests through. If you've run into a bug this can be paired up with [mitmproxy](http://mitmproxy.org/) to provide further insight into what is going on. You may be asked to use this feature when making bug reports.
