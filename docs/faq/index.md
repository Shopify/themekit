---
layout: default
---
# FAQ

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
