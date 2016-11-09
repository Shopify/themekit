---
layout: default
---
# Shopify Theme Kit

Theme Kit is a single binary that has no dependencies. Download the application
and with a tiny bit of setup you’re off to the theme creation races.

Using Theme Kit will enable you to

* Upload Themes to Multiple Environments
* Fast Uploads and Downloads
* Watch for local changes and upload automatically to Shopify
* Works on Windows, Linux and OS X

## Installation

### Automatic Installation

If you are on Mac or Linux you can use the following installation script to automatically
download and install Theme Kit for you. Please follow the directions outputted to your
console to change your bash profile so that you will have access to the `theme` command

```bash
curl https://raw.githubusercontent.com/Shopify/themekit/master/scripts/install | python
```

### Homebrew

If you have [homebrew](http://brew.sh/) installed you can install theme kit by running the following commands.

```bash
brew tap shopify/shopify
brew install themekit
```

### Manual Installation

Download and unzip the latest release.

| OS     | Architecture |          |
| :------| :------------| :------- |
| OS X   | 64-bit       |  [download](https://github.com/Shopify/themekit/releases/download/{{ site.themekitversion }}/darwin-amd64.zip)
| Windows| 64-bit       |  [download](https://github.com/Shopify/themekit/releases/download/{{ site.themekitversion }}/windows-amd64.zip)
| Windows| 32-bit       |  [download](https://github.com/Shopify/themekit/releases/download/{{ site.themekitversion }}/windows-386.zip)
| Linux  | 64-bit       |  [download](https://github.com/Shopify/themekit/releases/download/{{ site.themekitversion }}/linux-amd64.zip)
| Linux  | 32-bit       |  [download](https://github.com/Shopify/themekit/releases/download/{{ site.themekitversion }}/linux-386.zip)

**For OSX or Linux run the following commands**

```bash
cp ~/Downloads/theme /usr/local/bin #install the command onto your path
theme #test output of theme and make sure it is working
```

**For installing on Windows run the following commands**

- Create a folder inside `C:\Program Files\` called `Theme Kit`
- Copy the extracted program into `C:\Program Files\Theme Kit`
- Navigate to `Control Panel > System and Security > System`. Another way to get there is to Right-Click on `My Computer` and choose the `properties` item
- Look for the button or link called `Environment Variables`
- In the second panel look for the item called `Path` and double-click on it. This should open a window with a text field that is overflowing with content.
- Move your cursor all the way to the end and add the following: `;C:\Program Files\Theme Kit\`
- Click `OK` until all the windows are gone.
- To verify that Theme Kit has been installed, open `cmd.exe` and type in `theme`.

## Get API Access

To develop themes with theme kit, you will need to authorize theme kit to access your store.
Head to `https://[you-store-name].myshopify.com/admin/apps/private` it should look something
like this:

<img src="{{ "/assets/images/private-apps.png" | prepend: site.baseurl }}" />

Click on the `Create private apps` button. You will see this screen:

<img src="{{ "/assets/images/create-private-app.png" | prepend: site.baseurl }}" />

Fill out the information at the top and set the permissions of `Theme templates and theme assets` to
read and write access. Press save and you will be presented with the next screen. In it you will
see your access credentials. Please make note of the password. You will need it later.

<img src="{{ "/assets/images/private-app-password.png" | prepend: site.baseurl }}" />

## Use a new theme.

If you are starting form scratch and want to get a quick start run the following:

```bash
theme bootstrap --password=[your-password] --store=[you-store.myshopify.com]
```

This will create a new theme for your online store from the [Timber](https://shopify.github.io/Timber/) template. Then
it will download all those assets from shopify and automatically create a config.yml file for you.

## Configure an existing theme.

If you already have a theme on shopify and want to start using it you will need to
view it in your browser and grab the theme id from the url. It should look like the
following:

<img src="{{ "/assets/images/theme-id.png" | prepend: site.baseurl }}" />

Then once you have noted your theme id run the following commands

```bash
# create configuration
theme configure --password=[your-password] --store=[you-store.myshopify.com] --themeid=[your-theme-id]
# download and setup project in the current directory
theme download
```
