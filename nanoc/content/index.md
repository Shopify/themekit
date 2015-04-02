---
title: Shopify Themes for Everyone
---
# <%= @config[:project_name] %> is a cross-platform tool for building Shopify Themes

<%= @config[:project_name] %> is a single binary that has *no dependencies*. Download the application and with a tiny bit of setup you're off to the theme creation races.

## Features

- Upload Themes to Multiple Environments
- Fast Uploads and Downloads
- Watch for local changes and upload automatically to Shopify
- Works on Windows, Linux and OS X

## Downloads

With <%= @config[:project_name] %> setup is easy, just download the app for your platform, unzip and
run it from the command line.

<div class="versions col-1-1">
  <% VERSIONS.each do |version| %>
    <div class="<%= classes_for(version) %>">
      <div class="platform"><%= version[:platform] %></div>
      <%= link_to 'Download', download_url_for(version), class: 'btn btn-secondary' %>
    </div>
  <% end %>
</div>
