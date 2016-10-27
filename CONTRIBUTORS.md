# Contribution Guidelines

- Open an issue to discuss the feature/bug
- If feature/bug is deemed valid then fork repo
- Implement patch to resolve issue, include tests to prevent regressions/validate patch/be super awesome
- Submit pull request and mention maintainers
  - Current Maintainers: @ilikeorangutans, @chrisbutcher, @tanema

# Pre-Requisites

- [Go 1.7 or higher](https://golang.org/dl)
- [Glide](https://github.com/Masterminds/glide)

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

- Using [SemVer](http://semver.org/) update the version number in [version.go](version.go)
- Before continuing, verify that all tests are passing and binaries build cleanly.
  To easily verify everything you can simply enter the following: `make check`
- Merge changes into master, update the version number in kit/version.go, then
  create a tag named after the version. For example: `git tag 0.0.1 && git push origin --tags`
- Create a release using `make dist`
  - This will create binaries for all supported platforms and upload them to S3
  - It will also update the [manifest file](https://shopify-themekit.s3.amazonaws.com/releases/all.json)
    as well as the [latest release file](https://shopify-themekit.s3.amazonaws.com/releases/latest.json)
- Verify that both the manifest file and latest release file have been correctly updated
- On Github create a new release for the tag and take note of any relevant changes
  - Include a brief summary of all the changes
  - Include links to the Pull Requests that introduced these changes
  - Also upload the zipped binaries manually to Github so people can easily download them
- Update the [themekit](http://themekit.cat) website
  - `gem install jekyll`
  - `cd docs`
  - edit `_config.yml` and update the `themekitversion`
  - `jekyll serve` and check the website.
  - update any changes to the API
  - `git add . && git commit -m "Updating website"`

# Authors

- Chris Saunders <[https://github.com/csaunders](https://github.com/csaunders)>
- Jakob Külzer <[Shopify](https://shopify.com)>
- Chris Butcher <[Shopify](https://shopify.com)>
- Tim Anema <[Shopify](https://shopify.com)>
