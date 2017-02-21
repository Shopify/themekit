# Deploying Themekit
- You will need to have a valid `.env` file with credentials for the Amazon account.
- run `make check` to test and lint the code.
- Update ThemeKitVersion in `kit/version.go` and commit.
- run `git tag <version> && git push origin --tags`
  Any tags that are postfixed with `-beta` will not prompt users for update so if
  you want to release to a small group please use this method.
- Create a release using `make dist`
- On Github create a new release for the tag and take note of any relevant changes.
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
