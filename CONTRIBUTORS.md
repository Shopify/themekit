# Contribution Guidelines

- Open an issue to discuss the feature/bug
- If feature/bug is deemed valid then fork repo.
- Implement patch to resolve issue.
- Include tests to prevent regressions and validate the patch.
- Update the docs for any API changes.
- Submit pull request and mention maintainers. Current maintainers are @ilikeorangutans, @chrisbutcher, @tanema

# Pre-Requisites

Run `make mac_tools` or `make linux_tools` to install all the tools except go.

- [Go 1.7 or higher](https://golang.org/dl)
- [Glide](https://github.com/Masterminds/glide)
- [golint](https://github.com/golang/lint)
- [jekyll](https://jekyllrb.com/docs/installation/)
- makensis

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

- You will need to have a valid `.env` file with credentials for the Amazon account.
- run `make check`
- Update ThemeKitVersion in `kit/version.go`
- run `git tag <version> && git push origin --tags`
  Any tags that are postfixed with `-beta` will not prompt users for update so if
  you want to release to a small group please use this method.
- Create a release using `make dist`
- On Github create a new release for the tag and take note of any relevant changes
  - Include a brief summary of all the changes
  - Include links to the Pull Requests that introduced these changes
  - Upload the zipped binaries from `build/dist/` manually to the Github release.
  - Upload both the 32bit and 64bit windows installers from `build/dist` to the Github release as well.
- Update the [documentation website](https://shopify.github.io/themekit/)
  - run `make serve_docs`
  - update any changes to the API
  - commit any changes
- Update themekitversion in docs config `docs/_config.yml` to update the download links.
- Update `themekit.rb` formula for homebrew on https://github.com/Shopify/homebrew-shopify
  - run `make gen_sha` to generate the SHA256 for the darwin build
  - update the link and sha in the homebrew formula

# Authors

- Chris Saunders <[https://github.com/csaunders](https://github.com/csaunders)>
- Jakob Külzer <[Shopify](https://shopify.com)>
- Chris Butcher <[Shopify](https://shopify.com)>
- Tim Anema <[https://github.com/tanema](https://github.com/tanema)>
