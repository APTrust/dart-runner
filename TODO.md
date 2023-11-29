# TO DO

## Priority

- [x] In job files list, add option to show/hide hidden files.
- [x] In job files list, sort directories and files in case-insensitive alpha order, showing directories first, then files.
- [x] Set up Code Climate
- [x] Test Workflow Batch endpoints
- [ ] Fix _errors_ format in AbortWithErrorJSON. Most JS handlers are expecting an object, not an array or scalar. Need to check all handlers.
- [ ] Code refactor (de-dupe: Code Climate). We currently have lots of duplicate code, especially in tests.
- [x] Add logging to all critical sections
- [ ] Automated UI testing (Selenium)
- [x] Save artifacts from jobs (manifests, tag files)
- [ ] Expose artifacts in UI
- [x] Settings export
- [x] Settings export questions
- [x] Settings import
- [x] Settings import questions
- [ ] Replace calls to window.confirm() with a custom dialog because confirm() is deprecated.
- [ ] Centralize test code for loading JSON fixtures. (Too much duplication right now.)
- [ ] Centralize factory code for generating test objects. (Too much duplication right now.)
- [x] Build Dashboard & APTrust client
- [ ] Maybe - Add app setting for log level, with options Debug and Info.


## Later

- [x] Fix auto generation of output file path on job packaging screen
- [x] Fix web font load error
- [ ] Autofill other properties on job package page (e.g. if profile only allows tar, autoselect tar as serialization format)
- [x] Fix invalid JSON "EOF" being returned to job run page
- [x] Move job run JS to shared location so jobs, workflows, and batches can use it
- [ ] Script to pre-load sample jobs and workflows for developers
- [ ] PID file (or db entry) for running process
- [ ] Ping script from front-end
- [ ] Auto-open browser on start
- [ ] Move Minio to Docker (requires substantial changes to post-build tests)
- [ ] Context-sensitive help
- [ ] Test DART tar files with 7-zip. See https://github.com/APTrust/dart/issues/229
- [ ] [Systray](https://github.com/getlantern/systray/) or [Wails](https://wails.io) or [Fyne Systray](https://developer.fyne.io/explore/systray.html)
- [ ] Fix Windows paths
- [ ] Windows code signing
- [ ] Mac code signing
- [ ] User acceptance tests
