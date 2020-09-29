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
- [Golint](https://github.com/golang/lint)

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
