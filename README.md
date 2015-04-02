# Theme Kit
## Shopify Theme Manipulation Utilities and Libraries

[Read more about it on the website](http://themekit.cat)

### Utilities

Theme Kit provides a number of utilities when they are all built. These are to replicate the features in the [shopify_theme gem](https://github.com/shopify/shopify_theme) but aims to solve a few core problems:

- The architecture of the theme gem leaves much to be desired. It's a working tool, but adding new features is becoming very difficult
- The user experience for installing the theme gem leave much to be desired. A Windows user needs to jump through several hoops to get Ruby installed and not all OS X users are on a standard rubyists environment (rbenv, chruby, bundler, etc.)
- The gem isn't really extensible. It's a command line utility and that is it. Users aren't able to easily grab parts of the gem and use it for their own purposes.

### Library

The library is very much a work in progress but includes a few core things:

- YAML based Configuration Management
- Rudimentary Shopify API interaction via a ThemeClient
- API Limit control via a client side leaky bucket that is controlled via an arbiter (currently called Foreman)

## Downloads

You can get your hands on a build for your system by checking out the [releases section](https://github.com/csaunders/themekit/releases).


## Contributing

Please see CONTRIBUTORS file for details
