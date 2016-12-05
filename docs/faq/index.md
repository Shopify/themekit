---
layout: default
---
# FAQ

## How do I remove Theme Kit?

You can easily remove theme kit from your command line by running the following
command:

```bash
rm $(which theme)
```

## Theme Kit does not upload my changes.

This usually means that your file is being ignored. Please check your ignore
patterns and see the [documentation on ignore patterns]({{ '/ignores' | prepend: site.baseurl }}).

## I am getting the error: 'TLS handshake timeout'

This usually has to do with a file descriptor limit over using theme kit for a
while. You can usually fix this by restarting your terminal, however to fix it
in the long term please see the next question.

## I am getting the error: 'Could not watch directory, too many open files'

You are probably using OSX and OSX has an unusually low limit on file descriptors.
You need to raise this limit manually. You can do this by running the following
commands.

```bash
echo kern.maxfiles=65536 | sudo tee -a /etc/sysctl.conf
echo kern.maxfilesperproc=65536 | sudo tee -a /etc/sysctl.conf
sudo sysctl -w kern.maxfiles=65536
sudo sysctl -w kern.maxfilesperproc=65536
ulimit -n 65536 65536
```
