# Contribution Guidelines

- Open an issue to discuss the feature/bug
- If feature/bug is deemed valid then fork repo
- Implement patch to resolve issue, include tests to prevent regressions/validate patch/be super awesome
- Submit pull request and mention maintainers
  - Current Maintainers: @ilikeorangutans, @chrisbutcher

# Pre-Requisites

In order to work on this project **you'll need to have godep installed** on your system. This can easily
be done by running the following command:

  go get github.com/tools/godep

Please see the [godep documentation](https://github.com/tools/godep) for more details.

# Getting the Source Code

You can easily obtain the source code and all dependencies by typing the following into a console:

  go get -u github.com/Shopify/themekit

Switch to that directory and run the Makefile. This will create a `build/development` directory which will
contain the program. The Makefile will use `godep` to ensure that the builds are reliable for your system.

# Creating Releases

- Using [SemVer](http://semver.org/) update the version number in [version.go](version.go)
- Before continuing, verify that all tests are passing and binaries build cleanly
  - To easily verify everything you can simply enter the following: `make test && make`
- Merge changes into master and create a tag named after the version
  - For example: `git tag 0.0.1 && git push origin --tags`
- Create a release using `make dist`
  - This will create binaries for all supported platforms and upload them to S3
  - It will also update the [manifest file](https://shopify-themekit.s3.amazonaws.com/releases/all.json) as well as the [latest release file](https://shopify-themekit.s3.amazonaws.com/releases/latest.json)
- Verify that both the manifest file and latest release file have been correctly updated
- On Github create a new release for the tag and take note of any relevant changes
  - Include a brief summary of all the changes
  - Include links off to the Pull Requests that introduced these changes
  - Also upload the zipped binaries manually to Github so people can easily download them
- Update the [themekit](http://themekit.cat) website
  - `git checkout gh-pages && pushd nanoc`
  - `nanoc compile && popd`
  - `git add . && git commit -m "Updating website"`
  - `git push origin gh-pages`

# Authors

- Chris Saunders <[Shopify](https://shopify.com)>
- Jakob Külzer <[Shopify](https://shopify.com)>
- Chris Butcher <[Shopify](https://shopify.com)>
