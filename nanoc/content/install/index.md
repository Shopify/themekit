---
title: Installation Guide
---

# Installation Guide

Download and Unzip the <%= link_to 'latest release from Github', "#{@config[:repository]}/releases" %>

If you are using <strong>Mac OS or Linux</strong> follow <%= link_to 'these instruction', '#unix-like' %>

If you are using <strong>Windows</strong> follow <%= link_to 'these instruction', '#windows' %>


<hr />


<a name="unix-like"></a>

# Installing on Mac OS & Linux

- Create a directory called `Applications/bin`. You can enter the command in a terminal window to create it:

  `mkdir -p ~/Applications`

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

<a name="windows"></a>

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

<a name="expected-command-output"></a>

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
