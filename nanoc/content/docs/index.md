---
title: Documentation
---

# Documentation

- [Overview](#overview)
- [Commands](#commands)
- [Configuration](#config-file)
  - [Environments](#environments)
  - [Configuration Variables](#config-variables)
  - [Example Configuration](#config-example)
  - [Ignore File](#ignore-file)

## <a id="overview" href="#overview">Overview</a><i class="fa fa-bookmark"></i>

<%= @config[:project_name] %> comes with a number of utilities that you can use to interact one or more themes on Shopify. To see the list
of commands that are available, enter the following command into your terminal:

<pre><code>
#!bash
theme --help
</code></pre>

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

## <a id="commands" href="#commands">Commands</a><i class="fa fa-bookmark"></i>

0. [Common Arguments](#common-args)
1. [configure](#configure)
2. [bootstrap](#bootstrap)
3. [watch](#watch)
4. [download](#download)
5. [remove](#remove)
6. [replace](#replace)
7. [upload](#upload)


-----------------------------------

### <a id="common-args" href="#common-args">Common Arguments</a><i class="fa fa-bookmark"></i>

There are some arguments that you can pass into every command:

1. `env`: The *environment* to register the configuration settings under. These are just free form text, so you can name it anything. Common names are `staging`, `production`, `test` and the default is `development`
2. `dir`: The directory that the configuration file (called `conf.yml`) will live. This allows you to update multiple themes without having to change into each themes directory.

### <a id="configure" href="#configure">Configure</a><i class="fa fa-bookmark"></i>

Use this command to create or update configuration files. If you don't want to run the command in the directory that will contain the configuration file you will need to specify the directory manually.

The following options **must be provided**:

1. `password`: Your Shopify Private App password. This is needed to make authenticated calls against the Shopify API. Create a Private Application and use the value from the **Password** field, obtained at `https://<your subdomain>.myshopify.com/admin/apps/private/<id>`
2. `domain`: Your `<your subdomain>.myshopify.com` domain without any protocol or other information.

Additional arguments you can provide:

1. `bucketSize`: Shopify uses a leaky bucket strategy for rate limiting API access. If you have been granted additional API usage you can update that here. Internally it is used to prevent failures that can be caused by exhausted API limits.
2. `refillRate`: The rate at which tickets are restored in the leaky bucket (per second).

### <a id="bootstrap" href="#bootstrap">Bootstrap</a><i class="fa fa-bookmark"></i>

If you are starting a new theme and want to have some sane defaults, you can use the bootstrap command. It will create a new theme based on [Timber](https://github.com/shopify/timber), update your configuration file for that theme and download it to your computer.

**Configuration is required before bootstrapping**

The available options are as follows:

1. `setid`: When setid is true your configuration file will be updated to include the appropriate Theme ID. This will ensure that changes are uploaded to the correct place. You can opt out of this by passing setting the flag to false (i.e. `--setid=false`)
2. `prefix`: Prefix to add to your theme. If a prefix is provided, the theme created on Shopify will be called `PREFIX-Timber-VERSION`, otherwise it will simply be `Timber-VERSION`
3. `version`: The version of Timber that you want to use for developing a theme. By default it will use the [latest stable release](https://github.com/Shopify/Timber/releases). Latest master can also be used by setting the version flag (i.e. `--version=master`)

### <a id="watch" href="#watch">Watch</a><i class="fa fa-bookmark"></i>

Theme watch will start a process that will watch the specified directory for changes and upload them to the specified Shopify theme. Any changes will be logged to the terminal and the status of the upload will be logged as well. The program can be stopped by simply typing `ctrl+C`.

By default watch starts up two workers who will perform the upload, allowing for faster uploads to Shopify. The number of workers can be changed in your [configuration file](#config-example) though be aware that you may run the risk of slowdowns due to the leaky bucket getting drained.

To ease integrating the watcher with tools such as [LiveReload](http://livereload.com/), you can provide an the optional `--notify` argument to a file you want to have updated when the workers have gone idle. For example, if you had LiveReload watching for update made to a file at `/tmp/theme.update` you would enter the following command:

<pre><code>theme watch --notify=/tmp/theme.update</code></pre>

**Concurrent Uploads to multiple Environments**

Another command that can be useful for theme development is the `--allenvs` flag. By passing in this flag to the `theme watch` command, themekit will create watchers for every environment in your [configuration file](#config-files). **Note** currently there is no way to opt-out specific configurations, which means that changes made here will be sent to all your configured themes which may include production versions.

### <a id="download" href="#download">Download</a><i class="fa fa-bookmark"></i>

If called without any arguments, it will download the entire theme. Otherwise if you specify the files you want to download, then only those files will be retrieved. For example if you wanted to download the `404` and `article` liquid templates you could enter the following command:

<pre><code>theme download templates/404.liquid templates/article.liquid</code></pre>

### <a id="remove" href="#remove">Remove</a><i class="fa fa-bookmark"></i>

Deletes theme files both locally and on Shopify. This command must be called with at least one filename.

### <a id="replace" href="#replace">Replace</a><i class="fa fa-bookmark"></i>

Removes remote files and replaces them with local files.

### <a id="upload" href="#upload">Upload</a><i class="fa fa-bookmark"></i>

Upload specified files to Shopify theme. This command requires at least one filename.

-----------------------

## <a id="config-file" href="#config-file">Configuration</a><i class="fa fa-bookmark"></i>

### <a id="environments" href="#environments">Environments</a><i class="fa fa-bookmark"></i>

Environments allow you to manage where to upload your theme changes to. This helps reduce the errors that could happen when swapping between environments by directly modifying the config file. Environments are named whatever you want and will all contain roughly the same information, though perhaps with some minor changes (store, theme id, etc.)

### <a id="config-variables" href="#config-variables">Configuration Variables</a><i class="fa fa-bookmark"></i>

A configuration can contain a number of things on top of what can be added by invoking `theme configure`.

- `theme_id`: The ID of the theme to upload changes to. **Beware** that if this is blank all changes will be uploaded to your user visible theme.
- `password`: API credentials to update and manipulate themes on Shopify
-  `store`: Your `.myshopify.com` domain (i.e. pokeshop.myshopify.com). Note that you **do not** need to include `http://` or `https://`
- `ignore_files`: A list of specific files to ignore (i.e. `config/settings.html`). You can also ignore based on patterns (such as all png files), but those patterns will need to be wrapped in double quotes (i.e. `"*.png"`).
- `ignores`: A list of files containing patterns of files that should be ignored. These files are somewhat similar to the [.gitignore file](https://www.kernel.org/pub/software/scm/git/docs/gitignore.html)
- `bucket_size`: The size of your Access Tokens bucket. By default this is 40 and cannot be configured by the shop.
- `refill_rate`: The rate at which tickets are replenished by the Shopify API (per second). By default this is 2 and cannot be configured by the shop.
- `concurrency`: The number of workers to spawn when running `theme watch`. By default it is 2 to ensure that API limits don't get hit. This can be changed, though keep in mind that it can cause slowdowns if API limits are indeed reached. For most regular usage this might not be the case, but could happen if using automated that create a lot of File System events.
- `proxy`: Proxy server to route HTTP requests through. If you've run into a bug this can be paired up with [mitmproxy](http://mitmproxy.org/) to provide further insight into what is going on. You may be asked to use this feature when making bug reports.

### <a id="ignore-file" href="#ignore-file">The Ignore File</a><i class="fa fa-bookmark"></i>

Ignore files are simply lists of patterns or files you'd like to skip when uploading or downloading files. This can be useful for
situations where your build system generates a lot of temporary files, or files that Shopify won't allow. If you have a build system that generates temporary files in a directory called `build`, your ignorefile could look like this:

<pre><code>
#!text
build/*
settings.html
</code></pre>

### <a id="config-example" href="#config-example">Example Configuration</a><i class="fa fa-bookmark"></i>

The following is a comprehensive example of what the contents of a configuration might look like.

<pre><code>
#!yaml
production:
  theme_id:
  password: abracadabra
  store: pokeshop.myshopify.com
staging:
  theme_id:
  password: alakazam
  store: pokeshop-staging.myshopify.com
development:
  theme_id: 123
  password: abracadabra
  store: pokeshop.myshopify.com
  ignores:
    - myignores
  ignore_files:
    - config/settings.html
    - "*.png"
    - "*.jpg"
  refill_rate: 2
  bucket_size: 40
  concurrency: 4
  proxy: http://localhost:8080
</code></pre>
