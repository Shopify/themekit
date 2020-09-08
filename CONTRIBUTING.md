# Contributing

We welcome your contributions to the project. There are a few steps to take when looking to make a contribution.

- Open an issue to discuss the feature/bug
- If feature/bug is deemed valid then fork repo.
- Implement patch to resolve issue.
- Include tests to prevent regressions and validate the patch.
- Update the docs for any API changes.
- Submit pull request and mention maintainers. Current maintainers are @tanema, @chrisbutcher

# Bug Reporting

Theme Kit uses GitHub issue tracking to manage bugs, please open an issue there.

# Feature Request

You can open a new issue on the GitHub issues and describe the feature you would like to see.

# Developing Theme Kit

Requirements:

- Go 1.12 or higher

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

## Test Mocks

If interfaces change, we need to regenerate the mocks with Mockery.

To install Mockery:

`go get github.com/vektra/mockery/...`

Then cd into the directory of the interface and run:

`mockery --name=InterfaceName --output=_mocks`

# Debugging Requests

A man in the middle proxy is the easiest way to introspect the requests that themekit makes. To start using it please do the following.

- `brew install mitmproxy`
- `mitmproxy -p 5000 -w themekit_dump` This will start it listening on port 5000 and write to the file `themekit_dump`
- in another console, in your project directory run `theme deploy --proxy http://localhost:5000`
- After that command finished, you can quit mitmproxy by pressing q then y

If troubleshooting an issue for a partner, we can ask that they provide the `themekit_dump` log which can then be loaded into mitmproxy for analysis.

# Deploying Theme Kit

- Update ThemeKitVersion in `src/release/release.go` and commit.
- Make the release tool `go install ./cmd/tkrelease`
- run `git tag <version> && git push origin --tags && git push`
- Update the changelog.txt with the date of the version release
- Release using tool
  - build all distributions `make all`
  - release `tkrelease -k="AWS_ACCESS_KEY" -s="AWS_SECRET_KEY" vX.X.X`
    - If releasing a different version than in `src/release/version.go` you can use `-f` to force, sometimes this is necessary for specific issue tags like `v0.0.0-issue432` when trying to debug a issue.
    - Using beta/alpha tags on the version number will stop themekit from automatically updating to that version. It would have to be typed in specifically like `theme update --version=v1.0.4-rc1`
- On GitHub create a new release for the tag and take note of any relevant changes.
  - Include a brief summary of all the changes
  - Include links to the Pull Requests that introduced these changes
- Update the [documentation website](https://shopify.github.io/themekit/)
  - run `make serve_docs`
  - update any changes to the API
  - commit changes
- (Consider waiting a day before performing the next steps, in case a defect is reported and we need to issue a patch release).
- Update `themekitversion` in docs config `docs/_config.yml` to update the download links,
  then run `make md5s` to generate the checksums for the new files. Add these to the `docs/_config.yml`
  file as well.
- Update Chocolatey package in the `choco` folder
    - Update the version in `choco/themekit.nuspec`
    - Update the version and checksums in `choco/tools/chocolateyinstall.ps1`
    - Run `choco pack` in a window VM
    - Log into [https://chocolatey.org/](chocolatey.org) (use the themekit@shopify.com credentials) and submit an update for approval
- Update `themekit.rb` formula for homebrew on https://github.com/Shopify/homebrew-shopify
  - run `make sha` to generate the SHA256 for the darwin build
  - update the link, sha and version in the homebrew formula
