---
layout: default
---
# Theme Kit Ignore Patterns

Theme kit has a couple of ways to ignore files from your commands. You can provide
a list of ignore patterns to your `ignore_files` value or provide a list of file
paths to your `ignores` value.

## Ignore Files

Ignore files should have a certain format to be valid. One pattern per line. However
it is valid to have blank lines for comment lines prefixed with a `#` value.

## Patterns

There are a few rules to for the specifications of ignore patters.

- Patterns will be trimmed of whitespace at the beginning and end of the pattern
- Any plain file name (without a `*` character in it) will be matched within the
  project directory. So if you specify a pattern `no.txt` this will be matched with
  `$PROJECT_DIR/no.txt` but also `$PROJECT_DIR/templates/no.txt`
- Any file pattern has a `/` at the end like `config/`, this will be matched with
  a glob `config/*`
- Any pattern containing a glob `*` will be scoped to the project directory. So
  with the pattern `*.gif` this will be matched as `$PROJECT_DIR/*.gif`
- Any glob pattern that does not start with a glob, will be matched with a prefixed
  glob. So a pattern like `build/*` will be matched as `$PROJECT_DIR/*build/*`
- Any patter that starts with a **/** and ends with a **/** will be considered a
  regular expression and will match the whole path. An example pattern would be
  `/\.(txt|gif|bat)$/` that would match any file with the txt, gif or bat extentions.

## Ignores in config.yml example

```yaml
development:
  ... #other content
  ignore_files:
  - config/settings_data.json
  - "*.png" #patterns that start with * need to be quoted to have vaild yaml
  - /\.(txt|gif|bat)$/
  ignores:
  - themekit_ignores #file to load ignore patterns, check out the ignore file example
```

## Ignore File example

```
# $PROJECT_DIR/themekit_ignores
#plain file names
config/settings_data.json

# globs
*.png

# regex
/\.(txt|gif|bat)$/
```


