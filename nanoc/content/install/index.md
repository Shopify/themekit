---
title: Installation Guide
---

# Automatic Installation

If you are on **Mac or Linux** you can use the following installation script to automatically download and install
<%= @config[:project_name] %> for you. You will need to make a small change to your shell initialization file
(i.e. `.bashrc`) before `theme` will show up in your console:

`<%= @config[:install_unix_script] %>`

An automated installer for Windows will be coming soon.

# Manual Installation Guide

Download and Unzip the <%= link_to 'latest release from Github', "#{@config[:repository]}/releases" %>

If you are using <strong>Mac OS or Linux</strong> follow <%= link_to 'these instructions', '#unix-like' %>

If you are using <strong>Windows</strong> follow <%= link_to 'these instructions', '#windows' %>

If you want to <strong>install from source</strong> follow <%= link_to 'these instructions', '#install-from-source' %>

<hr />


<a id="unix-like"></a>

# Installing on Mac OS & Linux

- Create a directory called `Applications/bin`. You can enter the command in a terminal window to create it:

  <pre><code>mkdir -p ~/Applications</code></pre>

- Open the `~Applications/bin` directory and move the downloaded `theme` program into it. On Mac OS you can enter the following command to open both teh the `bin` and `Download` directories.

  <pre><code>open ~/Applications/bin && open ~/Downloads</code></pre>

- Open your configuration file for your terminal interface:
  - If you are using Bash your file will either be called `.profile` or `.bashrc`
  - If you are using ZSH your file will be called `.zshrc`
  - If you are using another shell you probably know how to do this already

- At the bottom of your file add the following:

<li class="nostyle">
  <ul>
    <li class="nostyle">
      <pre><code>
#!bash
export PATH=$PATH:~/Applications/bin
      </code></pre>
    </li>
  </ul>
</li>

- Reload your shell by closing your terminal window and opening a new one

- Verify that <%= @config[:project_name] %> has been installed by typing `theme --help` and [you should see output similar to this.](#expected-command-output)

<a id="windows"></a>

# Installing on Windows

- Create a folder inside `C:\Program Files\` called `<%= @config[:project_name] %>`
- Copy the extracted program into `C:\Program Files\<%= @config[:project_name] %>`
- If the program is missing `.exe` at the end, rename the file and add the `.exe` file extension
- Navigate to <strong><code>Control Panel > System and Security > System</code></strong>
  - Another way to get there is to Right-Click on `My Computer` and choose the `properties` item
- Look for the button or link called `Environment Variables`
- In the second panel look for the item called `Path` and double-click on it. This should open a window with a text field that is overflowing with content.
- Move your cursor all the way to the end and add the following: `;C:\Program Files\<%= @config[:project_name] %>\`
- Click **OK** until all the windows are gone.
- To verify that <%= @config[:project_name] %> has been installed, open `cmd.exe` and type in `theme --help`. [You should see output similar to this.](#expected-command-output)

# Install from Source

Before you can get started you will need to <%= link_to 'install Go for your platform', 'https://golang.org/dl/' %>. With
go installed, also <%= link_to 'verify go has been installed correctly', 'https://golang.org/doc/install#testing' %>.

You will also need to have <%= link_to 'git installed', 'https://git-scm.com/downloads' %> in order to properly download themekit and it's dependencies.

You will need to add the `bin` directory within your go installation to your PATH. On a unix-like system using Bash,
you'd do something like this:

<pre><code>
export PATH=PATH-TO-YOUR-GO-FOLDER/bin:$PATH &#62;&#62; ~/.bashrc
source ~/.bashrc
</code></pre>

Download themekit:

<pre><code>go get github.com/Shopify/themekit</code></pre>

Install themekit:

<pre><code>go install github.com/Shopify/themekit/cmd/theme</code></pre>

To verify that <%= @config[:project_name] %> has been installed by typing `theme --help`. [You should see output similar to this.](#expected-command-output)

<a id="expected-command-output"></a>

# Expected output when running `theme --help`

Below is what you should see when running the program with the `--help` flag.

<pre><code>
#!text
Usage of theme:
  -command="download": An operation to be performed against the theme.
  Valid commands are:
    configure:
        Create a configuration file
    bootstrap:
        Bootstrap a new theme using Shopify Timber
    upload <file> [<file2> ...]:
        Add file(s) to theme
    download [<file> ...]:
        Download file(s) from theme [default]
    remove <file> [<file2> ...]:
        Remove file(s) from theme
    replace [<file> ...]:
        Overwrite theme file(s)
    watch:
        Watch directory for changes and update remote theme
</code></pre>

<a id="install-from-source"></a>
