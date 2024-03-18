# Contributing to Theme Kit

Theme Kit is open source and we welcome your contributions to the project. The following steps illustrate the workflow for making a contribution:

1. [Open an issue](https://github.com/Shopify/themekit/issues) to report a bug or request a feature.
2. If the feature or bug is deemed valid, then [fork the repo](https://docs.github.com/en/github/getting-started-with-github/fork-a-repo).
3. Implement a patch to resolve the issue.
4. Include tests to prevent regressions and validate the patch.
5. Update the documentation for any API changes.
6. Submit [a pull request](https://github.com/Shopify/themekit/pulls) and mention maintainers. The current maintainers are @shopify/app-management

## Developing Theme Kit

Although Theme Kit is a command line interface, it's also a library that can be used in Go application development. For more information, refer to the [development documentation for Theme Kit](https://pkg.go.dev/github.com/Shopify/themekit).

### Requirements

- Go 1.22.1 or higher
- [Golint](https://github.com/golang/lint)

### Set up your development environment

You can set up your development environment by running the following commands:

```
git clone git@github.com:Shopify/themekit.git # get the code
cd themekit                                   # change into the themekit directory
make                                          # build themekit, will install theme kit into $GOPATH/bin
theme version                                 # should output the current themekit version
```

### Helpful commands

- `make`: Compiles themekit into your `GOPATH/bin`.
- `make test`: Runs linting/vetting/testing to make sure your code meets high quality standards.
- `make help`: Shows all of the commands available to you.

### Go development

You can run the following command to get started and retrieve the `themekit` library:

```
go get -u github.com/Shopify/themekit
```

### Javascript, Gulp, and Node development

Shopify uses Theme Kit in many of our Gulp processes. To interact with Theme Kit using Javascript, check out our [node-themekit library](https://github.com/Shopify/node-themekit).

## Test mocks

If interfaces change, then you need to regenerate the mocks with [Mockery](https://github.com/vektra/mockery).

To install Mockery, run the following command:

`go get github.com/vektra/mockery/...`

After Mockery is installed, `cd` into the directory of the interface and run:

`mockery --name=InterfaceName --output=_mocks`

## Debug requests

A [mitmproxy (man in the middle proxy)](https://mitmproxy.org/) is the easiest way to introspect the requests that themekit makes.

To use the proxy, complete the following steps:

1. Run `brew install mitmproxy`.
2. Run `mitmproxy -p 5000 -w themekit_dump`. This will start it listening on port 5000 and write to the file `themekit_dump`.
3. In another console, in your project directory, run `theme deploy --proxy http://localhost:5000`.
4. After the `theme deploy` command executes, you can quit `mitmproxy` by entering `q` and then `y`.

> Note:
> If you are troubleshooting an issue for a partner, then they must provide the `themekit_dump` log. The log can then be loaded into `mitmproxy` for analysis.
