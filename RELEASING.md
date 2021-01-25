# Releasing Theme Kit

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
- Update the [documentation website](https://shopify.dev/tools/theme-kit)
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
