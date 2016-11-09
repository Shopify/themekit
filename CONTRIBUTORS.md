# Contribution Guidelines

- Open an issue to discuss the feature/bug
- If feature/bug is deemed valid then fork repo.
- Implement patch to resolve issue.
- Include tests to prevent regressions and validate the patch.
- Update the docs for any API changes.
- Submit pull request and mention maintainers. Current maintainers are @ilikeorangutans, @chrisbutcher, @tanema

# Pre-Requisites

Run `make tools` to install all the tools except go.

- [Go 1.7 or higher](https://golang.org/dl)
- [Glide](https://github.com/Masterminds/glide)
- [golint](https://github.com/golang/lint)
- [jekyll](https://jekyllrb.com/docs/installation/)

# Getting the Source Code

You can easily obtain the source code and all dependencies by typing the following
into a console:

```
go get -u github.com/Shopify/themekit
```

Switch to that directory and run the `make`. This will install theme kit into your
go bin path and if you have your go/bin folder on your path, you should now have
access to the theme command. Run `theme` to make sure it is installed.

# Creating Releases

- run `make check`
- update version using [SemVer](http://semver.org/) prefixed with a v. (e.g. 'v0.5.0')
  - update ThemeKitVersion in `kit/version.go`
  - update themekitversion `docs/_config.yml`
  - run `git tag <version> && git push origin --tags`
    Any tags that are postfixed with `-beta` will not prompt users for update so if
    you want to release to a small group please use this method.
- Create a release using `make dist` this will:
  - Create binaries for all supported platforms
  - Upload to S3
  - Update the [manifest file](https://shopify-themekit.s3.amazonaws.com/releases/all.json)
  - Update the [latest release file](https://shopify-themekit.s3.amazonaws.com/releases/latest.json)
- On Github create a new release for the tag and take note of any relevant changes
  - Include a brief summary of all the changes
  - Include links to the Pull Requests that introduced these changes
  - Also upload the zipped binaries from `build/dist/` manually to Github so people can easily download them
- Update the [documentation website](https://shopify.github.io/themekit/)
  - run `gem install jekyll`
  - run `make serve_docs`
  - update any changes to the API
  - commit any changes
- Update `themekit.rb` formula for homebrew on https://github.com/Shopify/homebrew-shopify

# Authors

- Chris Saunders <[https://github.com/csaunders](https://github.com/csaunders)>
- Jakob Külzer <[Shopify](https://shopify.com)>
- Chris Butcher <[Shopify](https://shopify.com)>
- Tim Anema <[Shopify](https://shopify.com)>
