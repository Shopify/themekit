# Bug Reporting

Themekit uses github issue tracking to manage bugs, please open an issue there.

# Feature Request

You can open a new issue on the github issues and label it with `enhancment`

# Developing Themekit

Themekit requires go 1.8. You can setup your development environment by running the following:

```
go get -u github.com/Shopify/themekit
cd $GOPATH/src/github.com/Shopify/themekit
make [mac_tools|linux_tools] # install platform specific development tools
glide install # see https://github.com/Masterminds/glide for glide usage
make # build themekit
theme version #should output the current themekit version
```

This will install theme kit into $GOPATH/bin and you will now have access to the theme command.

Requirements:

- Go 1.8 or higher

Helpful commands

- `make` will compile themekit into your GOPATH/bin
- `make check` will run linting/vetting/testing to make sure your code is of high standard

# Deploying Themekit

- You will need to have a valid `.env` file with credentials for the Amazon account. Please contact an admin with this info
- run `make check` to test and lint the code.
- Update ThemeKitVersion in `kit/version.go` and commit.
- run `git tag <version> && git push origin --tags`
- Create a release using `make dist`
- On Github create a new release for the tag and take note of any relevant changes.
  - Include a brief summary of all the changes
  - Include links to the Pull Requests that introduced these changes
- Update the [documentation website](https://shopify.github.io/themekit/)
  - run `make serve_docs`
  - update any changes to the API
  - commit any changes
- Update `themekitversion` in docs config `docs/_config.yml` to update the download links.
- Update `themekit.rb` formula for homebrew on https://github.com/Shopify/homebrew-shopify
  - run `make gen_sha` to generate the SHA256 for the darwin build
  - update the link and sha in the homebrew formula
