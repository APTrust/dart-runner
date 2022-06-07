# Change Log

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
