---
layout: default
---
# FAQ

## How can I reload the page when I make changes

You will need to use another tool to do this, however there are many options available
to you to suit all use cases.

- [Prepros](https://prepros.io/)
- [LivePage Chrome Plugin](https://chrome.google.com/webstore/detail/livepage/pilnojpmdoofaelbinaeodfpjheijkbh?hl=en-US)
- [BrowserSync](https://www.browsersync.io/)
- [LiveReload](http://livereload.com/)

## How can I use LiveReload/Prepros/Browsersync/LivePage

You can find a guide to setup [Prepros on the shopify blog](https://www.shopify.com/partners/blog/live-reload-shopify-sass)
that can apply to many other tools as well.

In general though while running the `theme watch` command you can provide it with a
file to touch whenever a change has been completed.

```
theme watch --notify=/var/tmp/theme_ready
```

If this file does not exist, themekit will create it when the first change
happens. Then on subsequent events the file will be touched. This file is touched
only after the request to shopify has been completed and not after saving the file.
You can then provide this file path to your reloading program to trigger your browser
refresh.

## How do I remove Theme Kit?

You can easily remove Theme Kit from your command line by running the following
command:

```bash
rm $(which theme)
```

## Theme Kit does not upload my changes.

This usually means that your file is being ignored. Please check your ignore
patterns and see the [documentation on ignore patterns]({{ '/ignores' | prepend: site.baseurl }}).

## I am getting the error: 'TLS handshake timeout'

This usually has to do with a file descriptor limit over using Theme Kit for a
while. You can usually fix this by restarting your terminal, however to fix it
in the long term please see the next question.

## I am getting the error: 'Could not watch directory, too many open files'

You are probably using macOS and macOS has an unusually low limit on file descriptors.
You need to raise this limit manually. You can do this by running the following
commands.

```bash
echo kern.maxfiles=65536 | sudo tee -a /etc/sysctl.conf
echo kern.maxfilesperproc=65536 | sudo tee -a /etc/sysctl.conf
sudo sysctl -w kern.maxfiles=65536
sudo sysctl -w kern.maxfilesperproc=65536
ulimit -n 65536 65536
```

## I am using Cloud9 and it is replacing my files with other content

If you are using the Cloud9 Editor, you can make themekit work by placing all of
your theme files in a folder in your workspace, then run themekit under that new
folder. [Please refer to this corresponding issue](https://github.com/Shopify/themekit/issues/416)
