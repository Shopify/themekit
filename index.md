# Phoenix

# Theme asset interaction library and management tools written in Go

----

## Command Line Usage

The Phoenix library is used for the creation of several tools that make it easy to
work with the Shopify Assets API. Before you can get started, download the appropriate
collection of binaries for your system:

* **[Current Release - 0.0.3 alpha](https://github.com/csaunders/phoenix/releases/tag/0.0.3)**

Older Releases

* [0.0.2](https://github.com/csaunders/phoenix/releases/tag/0.0.2)
* [0.0.1](https://github.com/csaunders/phoenix/releases/tag/0.0.1)

Nightly Releases

When new commits are pushed and a release hasn't been made yet, there will be a nightly branch. If nothing has diverged from the latest release, a nightly build will not be available.

-------

Place the binary files somewhere and add that location to your classpath. On a Unix-like operating
system using Bash you'd do something like this:

    mkdir -p ~/utils/phoenix
    echo "export PATH=$PATH:/~/utils/phoenix" > ~/.bashrc
    source ~/.bashrc


### Theme Tools

#### theme-configure

If you are using Phoenix for the first time, this is what you will want to use to get your
`config.yml` setup correctly. This will build out the configuration file so you can interact
with your theme on Shopify.

Basic usage includes:

    cd theme-dir
    theme-configure --domain=somedomain.myshopify.com --access_token=TOKEN

**Access Tokens**

You will need to go to the *Apps area* of your shops Admin and create a new private app.
Copy the key from the **Password** field and use that for value of the `access_token` flag.

If all goes well you should have a `config.yml` file in your `theme-dir`

#### theme-watch

By running `theme-watch` you will start up a long running process that will watch your current directory
(and all subdirectories) for changes. If you update or remove files, those changes will be passed off to
Shopify and updated on your theme.

**Concurrency**

You can tune how many uploads you'd like to do at once by adjusting the `concurrency` field in your `config.yml`. By default you'll have a concurrency of **2** which means you should never have to worry about running out of API requests when working on a theme. If you aren't making a super large amount of updates at the same time you can tweak this value into something that is a bit more acceptable for you.

*Caveats*

Most API clients have an upper limit of 40 requests with a 2 request refill. So if you make a request that does more than 40 updates, after those first 40 requests you'll end up with an effective concurrency of 2 until all your files have been processed.

#### theme-manipulate

This tool provides a bunch of various actions you can perform. Many of these are what you'd do to
a single asset file, such as uploading or downloading.

##### theme-manipulate upload file [...]

Provide one or more files that you'd like to upload to Shopify.

#### theme-manipulate download [file ...]

Download one or more files from Shopify.

If no files are provided then all the themes assets will be downloaded to your local filesystem.
