# Releasing Theme Kit

To release a new version of Theme Kit, make sure to complete all of the following steps.

## 1. Update the Theme Kit version

i) Update `ThemeKitVersion` in `src/release/release.go` and commit your changes.
ii) Make the release tool by running `go install ./cmd/tkrelease`.
iii) Run `git tag <version> && git push origin --tags && git push`.
iv) Update `changelog.txt` with the date of the version release.

## 2. Release using tool

i) Build all distributions by running `make all`.
ii) Release `tkrelease -k="AWS_ACCESS_KEY" -s="AWS_SECRET_KEY" vX.X.X`.

    > Note:
    > If you're releasing a different version than in `src/release/version.go`, then you can use `-f` to force. Sometimes this is necessary for specific issue tags like `v0.0.0-issue432` when trying to debug a issue.
    > Using beta/alpha tags on the version number will stop `themekit` from automatically updating to that version. It would have to be typed in specifically (for example, `theme update --version=v1.0.4-rc1`).

## 3. Create a new release on GitHub

i) In GitHub, create a new release for the tag.
ii) Include a brief summary of all the changes pertaining to the release.
iii) Include links to the pull requests that introduced the changes.

## 4. Update the Theme Kit documentation on shopify.dev

i) Open a [pull request in shopify-dev](https://github.com/Shopify/shopify-dev/pulls) to update the [Theme Kit documentation](https://shopify.dev/tools/theme-kit)
ii) Commit your changes and tag the partner-facing docs team for review. After the PR is approved, and the new Theme Kit version is released, merge the docs PR.

## 5. Update Theme Kit installation links on GitHub

> Note:
> Before proceeding with the remaining steps, consider waiting a day. If a defect is reported shortly after we make the release public, then we'll need to issue a patch release.

i) Update the manual installation links on the [GitHub releases page](https://github.com/Shopify/themekit/releases).
ii) Verify that the links work, as expected.

## 6. Update the Chocolately package

Update the Chocolatey package in the `choco` folder:

i) Update the version in `choco/themekit.nuspec`.
ii) Update the version and checksums in `choco/tools/chocolateyinstall.ps1`.
iii) Run `choco pack` in a window VM.
iv) Log into [https://chocolatey.org/](chocolatey.org) (use `themekit@shopify.com` for your credentials) and submit an update for approval.

## 7. Update ThemeKit for Homebrew

Update the `themekit.rb` formula for Homebrew on [homebrew-shopify](https://github.com/Shopify/homebrew-shopify):

i) Run `make sha` to generate the SHA256 for the darwin build.
ii) Update the link, SHA, and version in the Homebrew formula.
