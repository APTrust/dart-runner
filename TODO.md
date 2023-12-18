# TO DO

## Priority

- [ ] Fix _errors_ format in AbortWithErrorJSON. Most JS handlers are expecting an object, not an array or scalar. Need to check all handlers.
- [ ] Fix question bug in export settings. Question select lists won't populate until settings are saved.
- [ ] Implement validation-only jobs.
- [ ] Ensure we can validate loose bags (directories)
- [ ] Implement upload-only jobs.
- [ ] Ensure we can upload directories (recursive file list?)
- [ ] Context-sensitive help
- [ ] BUG - Job Metadata page won't load if packaging format is not BagIt
- [ ] Allow import of DART-native profiles
- [ ] Replace calls to window.confirm() with a custom dialog because confirm() is deprecated.
- [ ] Code refactor (de-dupe: Code Climate). We currently have lots of duplicate code, especially in tests.
- [ ] Automated UI testing (Selenium)
- [ ] Centralize test code for loading JSON fixtures. (Too much duplication right now.)
- [ ] Centralize factory code for generating test objects. (Too much duplication right now.)
- [ ] Maybe - Add app setting for log level, with options Debug and Info.

## Later

- [ ] Ensure backwards compatibility with DART Runner v0.96-beta
- [ ] Script to pre-load sample jobs and workflows for developers
- [ ] PID file (or db entry) for running process
- [ ] Ping script from front-end
- [ ] Auto-open browser on start
- [ ] Test DART tar files with 7-zip. See https://github.com/APTrust/dart/issues/229
- [ ] [Systray](https://github.com/getlantern/systray/) or [Wails](https://wails.io) or [Fyne Systray](https://developer.fyne.io/explore/systray.html)
- [ ] Windows code signing
- [ ] Mac code signing
- [ ] User acceptance tests

## Done

- [x] Fix Windows paths in bagging
- [x] Fix Windows integration tests 
- [x] Fix auto generation of output file path on job packaging screen
- [x] Fix web font load error
- [x] Fix invalid JSON "EOF" being returned to job run page
- [x] Move job run JS to shared location so jobs, workflows, and batches can use it
- [x] In job files list, add option to show/hide hidden files.
- [x] In job files list, sort directories and files in case-insensitive alpha order, showing directories first, then files.
- [x] Set up Code Climate
- [x] Test Workflow Batch endpoints
- [x] Add logging to all critical sections
- [x] Save artifacts from jobs (manifests, tag files)
- [x] Expose artifacts in UI
- [x] Settings export
- [x] Settings export questions
- [x] Settings import
- [x] Settings import questions
- [x] Build Dashboard & APTrust client
- [x] Add paging to list pages. Show only 25-50 items at a time.
- [x] Improve job results display: bagging profile and all uploads (succeeded & failed)
- [x] Delete artifacts when deleting job 
- [x] Fix output path autofill on job packaging page
- [x] Move Minio to Docker (requires substantial changes to post-build tests)
- [x] Autofill other properties on job package page (e.g. if profile only allows tar, autoselect tar as serialization format)
- [x] Show flash confirmation message on successful save and delete
- [x] Init DB with profiles and bagging dir for fresh installation
