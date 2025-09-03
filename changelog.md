# Change Log

## v1.0 - 2025-09-03

* Added --skip-artifacts command-line option to skip saving artifacts (tag files and manifests) to seperate directory when bagging. https://trello.com/c/r7FEBuAK

## v0.99-beta - 2025-06-06

* Fixes to APTrust BagIt profile version 2.3
* Started automatically saving bag artifacts (tag files and manifests) to separate folder for each bag created.

## v0.98-beta - 2025-02-28

* Added Wasabi-TX as valid storage option for APTrust bags.

## v0.97-beta - 2025-01-22

* Updated Go dependencies to fix security vulnerabilities.

## v0.96-beta - 2023-03-09

* Fixed [DART runner writes incorrect tag values in workflow batch mode](https://github.com/APTrust/dart-runner/issues/9). This was a serious bug caused by race condition that we anticipated but handled incorrectly in versions 0.95 and earlier.

## v0.95-beta - 2022-08-11

* Fixed a problem reading from STDIN. Runner would fail to read piped STDIN that contained no newlines.

## v0.94-beta - 2022-07-05

* Fixed [Repeated fields have only one field (with last value) in bag-info.txt](https://github.com/APTrust/dart-runner/issues/7)

## v0.93-beta - 2022-06-07

* Fixed [missing root directory in tarred bags](https://github.com/APTrust/dart-runner/issues/5)
* Fixed [DART Runner is changing file names](https://github.com/APTrust/dart-runner/issues/6)

## v0.92-beta - 2021-12-14

* Fixed bug [No output on certain failures](https://github.com/APTrust/dart-runner/issues/2) that caused DART Runner to exit without sending detailed error info to stdout/stderr.
* Improved built-in help doc with info about error codes and output JSON format.

## v0.91-beta - 2021-12-02

* Added support for [piping JobParams into STDIN](https://github.com/APTrust/dart-runner/issues/1)
* Dropped the --job command-line argument. This was never implemented in the first place, and DART does not currently export jobs.


## v0.9-beta - 2021-11-03

* Initial release
