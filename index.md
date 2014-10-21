# Phoenix

# Theme asset interaction library and management tools written in Go

----

## Command Line Usage

The Phoenix library is used for the creation of several tools that make it easy to
work with the Shopify Assets API. Before you can get started, download the appropriate
collection of binaries for your system:

* **[Current Release - 0.0.1 alpha](https://github.com/csaunders/phoenix/releases/tag/0.0.1)**


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


#### theme-manipulate

This tool provides a bunch of various actions you can perform. Many of these are what you'd do to
a single asset file, such as uploading or downloading.

##### theme-manipulate upload file [...]

Provide one or more files that you'd like to upload to Shopify.

#### theme-manipulate download [file ...]

Download one or more files from Shopify.

If no files are provided then all the themes assets will be downloaded to your local filesystem.
