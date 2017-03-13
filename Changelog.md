# Change Log
All released changes to this project will be documented in this file.


## [v0.6.10](https://github.com/Shopify/themekit/releases/tag/v0.6.10) Mar 10, 2017

Released By: [@tanema](https://github.com/tanema)

- minimal vcs integration for `watch` command so that watch is reloaded when branch is changed https://github.com/Shopify/themekit/pull/364
- `watch` CPU usage optimizations: https://github.com/Shopify/themekit/pull/363
- `theme open -E` command to open the editor. https://github.com/Shopify/themekit/pull/358
- Development Profiling functionality https://github.com/Shopify/themekit/pull/352
- Go 1.8 sorting optimization https://github.com/Shopify/themekit/pull/348
- Windows Color Support https://github.com/Shopify/themekit/pull/347
- Better Config File Loading https://github.com/Shopify/themekit/issues/345

## [v0.6.9](https://github.com/Shopify/themekit/releases/tag/v0.6.9) Feb 21, 2017

Released By: [@tanema](https://github.com/tanema)

Bump number for bad update feed.


## [v0.6.8](https://github.com/Shopify/themekit/releases/tag/v0.6.8) Feb 21, 2017

Released By: [@tanema](https://github.com/tanema)

Stub release to release the windows installer properly.


## [v0.6.7](https://github.com/Shopify/themekit/releases/tag/v0.6.7) Feb 16, 2017

Released By: [@tanema](https://github.com/tanema)

Fixing watching symlinked directories. https://github.com/Shopify/themekit/issues/336


## [v0.6.6](https://github.com/Shopify/themekit/releases/tag/v0.6.6) Feb 08, 2017

Released By: [@tanema](https://github.com/tanema)

Fix for singlle file operations https://github.com/Shopify/themekit/pull/334


## [v0.6.5](https://github.com/Shopify/themekit/releases/tag/v0.6.5) Feb 07, 2017

Released By: [@tanema](https://github.com/tanema)

Watch now will reload automatically if the config.yml file has changed. https://github.com/Shopify/themekit/pull/329
Fixed install script for python 3 https://github.com/Shopify/themekit/pull/328
Upload directories and download wildcards https://github.com/Shopify/themekit/pull/327
Progress bar and verbose flag https://github.com/Shopify/themekit/pull/325


## [v0.6.4](https://github.com/Shopify/themekit/releases/tag/v0.6.4) Jan 20, 2017

Released By: [@tanema](https://github.com/tanema)

- Fixes the watch command delete operations https://github.com/Shopify/themekit/pull/313
- Removes deprecated Shopify api usage. https://github.com/Shopify/themekit/pull/320

The last point is notable because it will make replace work faster, however it will make download seem slower because instead of being able to download the whole theme in one request, the files are downloaded individually. This is done concurrently however so it should not be that much of a slow down.


## [v0.6.3](https://github.com/Shopify/themekit/releases/tag/v0.6.3) Jan 04, 2017

Released By: [@tanema](https://github.com/tanema)

Fixes for windows path handling with a more isolated approach so that windows paths are not so fragile.


## [v0.6.2](https://github.com/Shopify/themekit/releases/tag/v0.6.2) Jan 04, 2017

Released By: [@tanema](https://github.com/tanema)

Added readonly flag t config https://github.com/Shopify/themekit/pull/304
Added url and name flag to bootstrap https://github.com/Shopify/themekit/pull/302
Restricted all operations to project directories only for less requests https://github.com/Shopify/themekit/pull/301
Fixes bad windows watch paths https://github.com/Shopify/themekit/pull/297
Better Remote Asset Filtering https://github.com/Shopify/themekit/pull/293/files


## [v0.6.1](https://github.com/Shopify/themekit/releases/tag/v0.6.1) Dec 08, 2016

Released By: [@tanema](https://github.com/tanema)

Better Logging: https://github.com/Shopify/themekit/pull/280
Catching list connection errors: https://github.com/Shopify/themekit/pull/292


## [v0.6.0](https://github.com/Shopify/themekit/releases/tag/v0.6.0) Nov 28, 2016

Released By: [@tanema](https://github.com/tanema)

In in this version.

Fixing sections data upload https://github.com/Shopify/themekit/pull/271
Allowing directory config from config file https://github.com/Shopify/themekit/pull/270
Fixed notify file in watch command https://github.com/Shopify/themekit/pull/269
JSON config file support (defaults to yml) https://github.com/Shopify/themekit/pull/265
Expanding default ignores and watching less files https://github.com/Shopify/themekit/pull/261
Added Open command https://github.com/Shopify/themekit/pull/260
Added environment to logging https://github.com/Shopify/themekit/pull/259

misc:
https://github.com/Shopify/themekit/pull/266
https://github.com/Shopify/themekit/pull/256


## [v0.5.3](https://github.com/Shopify/themekit/releases/tag/v0.5.3) Nov 16, 2016

Released By: [@tanema](https://github.com/tanema)

Fixing constant errors on theme watch and giving better error messages for theme errors.

https://github.com/Shopify/themekit/pull/255


## [v0.5.2](https://github.com/Shopify/themekit/releases/tag/v0.5.2) Nov 09, 2016

Released By: [@tanema](https://github.com/tanema)

Fixing debouncing of file events https://github.com/Shopify/themekit/pull/245
Fixing bad error handling https://github.com/Shopify/themekit/pull/247


## [v0.5.1](https://github.com/Shopify/themekit/releases/tag/v0.5.1) Nov 07, 2016

Released By: [@tanema](https://github.com/tanema)

Just updating the command line documentation to scrub any mention of themekit.cat.


## [v0.5.0](https://github.com/Shopify/themekit/releases/tag/v0.5.0) Nov 07, 2016

Released By: [@tanema](https://github.com/tanema)

This is partial rewrite of theme kit that will allow better maintenance in the future. The docs have been updated as well as the command line enabling more interaction. Please check out the feature release here https://github.com/Shopify/themekit/pull/235


## [0.4.7](https://github.com/Shopify/themekit/releases/tag/0.4.7) Oct 12, 2016

Released By: [@tanema](https://github.com/tanema)

fixes https://github.com/Shopify/themekit/pull/225
fixes https://github.com/Shopify/themekit/pull/224

Fixing downloads and multiple events on watch.


## [0.4.6](https://github.com/Shopify/themekit/releases/tag/0.4.6) Oct 11, 2016

Released By: [@tanema](https://github.com/tanema)

Release version updated so people can get the update.


## [0.4.5](https://github.com/Shopify/themekit/releases/tag/0.4.5) Oct 11, 2016

Released By: [@tanema](https://github.com/tanema)

Patch for bad watching pattern.

Fixes https://github.com/Shopify/themekit/pull/222


## [0.4.4](https://github.com/Shopify/themekit/releases/tag/0.4.4) Oct 11, 2016

Released By: [@tanema](https://github.com/tanema)

Large update to take care of common problems.
- Large file uploads and odd file saving behaviour using watch have been fixed by debouncing file watching operation. https://github.com/Shopify/themekit/pull/208
- fixing `-allenvs` flag on watch. https://github.com/Shopify/themekit/pull/201
- We have compiled themekit using compiler flags to minimize build size.
- ignore applies to download https://github.com/Shopify/themekit/pull/212
- better developers api https://github.com/Shopify/themekit/pull/214
- less strict config validation https://github.com/Shopify/themekit/pull/210
- json files won't be downloaded in a single line https://github.com/Shopify/themekit/pull/204
- added continuous integration https://github.com/Shopify/themekit/pull/198


## [0.4.3](https://github.com/Shopify/themekit/releases/tag/0.4.3) Sep 23, 2016

Released By: [@chrisbutcher](https://github.com/chrisbutcher)

Fix PRs: https://github.com/Shopify/themekit/pull/191, https://github.com/Shopify/themekit/pull/192

Running previous builds of themekit with macOS Sierra causes crashes (https://github.com/Shopify/themekit/issues/195). New builds of themekit (> v. 0.4.3), made with with Go version 1.7.1, should address this.


## [0.4.2](https://github.com/Shopify/themekit/releases/tag/0.4.2) Aug 17, 2016

Released By: [@chrisbutcher](https://github.com/chrisbutcher)

Fixes https://github.com/Shopify/themekit/issues/176


## [0.4.1](https://github.com/Shopify/themekit/releases/tag/0.4.1) Aug 17, 2016

Released By: [@chrisbutcher](https://github.com/chrisbutcher)

Rolls back breaking changes introduced in 0.4.0

https://github.com/Shopify/themekit/issues/175
https://github.com/Shopify/themekit/issues/176


## [0.4.0](https://github.com/Shopify/themekit/releases/tag/0.4.0) Aug 17, 2016

Released By: [@chrisbutcher](https://github.com/chrisbutcher)

Fixes https://github.com/Shopify/themekit/issues/157

theme_id can now be written as a string (number with quotes) in config.yml, and when set to the value of "live", will allow users to opt into affecting the live, production theme.


## [0.3.6](https://github.com/Shopify/themekit/releases/tag/0.3.6) Mar 09, 2016

Released By: [@chrisbutcher](https://github.com/chrisbutcher)

Bugfix: https://github.com/Shopify/themekit/pull/145


## [0.3.5](https://github.com/Shopify/themekit/releases/tag/0.3.5) Jan 15, 2016

Released By: [@ilikeorangutans](https://github.com/ilikeorangutans)

Fixed issues in this release:

#119 better error reporting for invalid configuration values
#102 fixed incorrect glob that broke `replace` in sub directories


## [0.3.4](https://github.com/Shopify/themekit/releases/tag/0.3.4) Dec 16, 2015

Released By: [@ilikeorangutans](https://github.com/ilikeorangutans)

- https://github.com/Shopify/themekit/pull/122 On MacOS X we now automatically set ulimit to sane values
- https://github.com/Shopify/themekit/pull/124 Updated CONTRIBUTORS to describe release process
- https://github.com/Shopify/themekit/pull/123 New folders to watch by default


## [0.3.3](https://github.com/Shopify/themekit/releases/tag/0.3.3) Dec 16, 2015

Released By: [@ilikeorangutans](https://github.com/ilikeorangutans)

Addressed issues:
- https://github.com/Shopify/themekit/issues/101
- https://github.com/Shopify/themekit/pull/120
- https://github.com/Shopify/themekit/pull/115


## [0.3.2](https://github.com/Shopify/themekit/releases/tag/0.3.2) Nov 17, 2015

Released By: [@csaunders](https://github.com/csaunders)

Fixes upload bugs that could be experienced when working with extremely large text files. Also fixes a bug that was introduced for Windows users relating to the incorrect path separator being used when uploading to Shopify.
- https://github.com/Shopify/themekit/pull/112
- https://github.com/Shopify/themekit/pull/110


## [0.3.1](https://github.com/Shopify/themekit/releases/tag/0.3.1) Sep 21, 2015

Released By: [@csaunders](https://github.com/csaunders)

Whenever no command is provided themekit will output the `--help` information to show the user what operations they can perform.

This does mean that the default operation (`download`) has been removed, so a command must always be provided for themekit to operate.

---

If you are using themekit version 0.3.0 you can apply this update by entering the following in your console:

`theme update`

You can verify what version of themekit you are on by entering the following:

`theme version`


## [0.3.0](https://github.com/Shopify/themekit/releases/tag/0.3.0) Sep 18, 2015

Released By: [@csaunders](https://github.com/csaunders)

There are a number of improvements and features that have been included in this release. This release consisted of a [milestone whose goal was to make Themekit a true alternative to the theme gem](https://github.com/Shopify/themekit/issues?q=milestone%3A%22Features+needed+in+order+to+Sunset+Shopify+Theme%22+is%3Aclosed)

The major changes include:
- Truly Cross-Platform terminal colours
- An automatic installer for Mac OS and Linux
- Improved handling of HTTP errors from Shopify
- Fixes to the "key parser" which resulted in attempts to upload assets under an invalid name
- Usability Improvements
- Ignore "compiled assets" when downloading from shopify (`application.scss.liquid` automatically becomes `application.css`)
- Fixed an issue where projects with large amounts of files would cause the application to crash with a cryptic error

Major Features:
- Themekit will check for updates and if they are available will inform the user about them. For now, the user will have to opt-in to applying the update by typing the command `theme update`

Props to @chrisbutcher for all the codereviews!


