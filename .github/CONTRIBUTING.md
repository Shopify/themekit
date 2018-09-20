# Contributing

We welcome your contributions to the project. There are a few steps to take when looking to make a contribution.

- Open an issue to discuss the feature/bug
- If feature/bug is deemed valid then fork repo.
- Implement patch to resolve issue.
- Include tests to prevent regressions and validate the patch.
- Update the docs for any API changes.
- Submit pull request and mention maintainers. Current maintainers are @tanema, @chrisbutcher

# Bug Reporting

Themekit uses github issue tracking to manage bugs, please open an issue there.

# Feature Request

You can open a new issue on the github issues and describe the feature you would like to see.

# Developing Themekit

Requirements:

- Go 1.11 or higher

You can setup your development environment by running the following:

```
git clone git@github.com:Shopify/themekit.git # get the code
cd themekit                                   # change into the themekit directory
make                                          # build themekit, will install theme kit into $GOPATH/bin
theme version                                 # should output the current themekit version
```

Helpful commands

- `make` will compile themekit into your GOPATH/bin
- `make test` will run linting/vetting/testing to make sure your code is of high standard
- `make help` will tell you all the commands available to you.

# Deploying Themekit

- Update ThemeKitVersion in `kit/version.go` and commit.
- run `git tag <version> && git push origin --tags && git push`
- create a deploy on Buildkite and set the DEPLOY_VERSION environment variable in the build
  settings to the tag you want to deploy. If the themekit version does not equal the deploy
  version (like a prerelease version), use the FORCE_DEPLOY environment var.
- On Github create a new release for the tag and take note of any relevant changes.
  - Include a brief summary of all the changes
  - Include links to the Pull Requests that introduced these changes
- Update the [documentation website](https://shopify.github.io/themekit/)
  - run `make serve_docs`
  - update any changes to the API
  - commit changes
- Update `themekitversion` in docs config `docs/_config.yml` to update the download links,
  then run `make md5s` to generate the checksums for the new files. Add these to the `docs/_config.yml`
  file as well.
- Update the changelog.txt
- Update `themekit.rb` formula for homebrew on https://github.com/Shopify/homebrew-shopify
  - run `make gen_sha` to generate the SHA256 for the darwin build
  - update the link and sha in the homebrew formula
- Notify the maintainer of the AUR themekit package https://aur.archlinux.org/packages/shopify-themekit-bin
  of an update so they can release a new version.
