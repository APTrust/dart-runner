# TO DO

## Priority

- [ ] Display Job artifacts on separate page, possibly with side bar and main pane.
- [ ] Ensure we can write loose bags (directories) - Writer should implement same interface as TarWriter.
- [ ] Ensure we can validate loose bags (directories)
- [ ] Ensure we can upload directories (recursive file list?)
- [ ] BUG - Job Metadata page won't load if packaging format is not BagIt
- [ ] Code refactor (de-dupe: Code Climate). We currently have lots of duplicate code, especially in tests.
- [ ] Automated UI testing (Selenium)
- [ ] Centralize test code for loading JSON fixtures. (Too much duplication right now.)
- [ ] Centralize factory code for generating test objects. (Too much duplication right now.)
- [ ] Maybe - Add app setting for log level, with options Debug and Info.
- [ ] Rotate log files at about 5 MB
- [ ] When opening log file, alert user that file was opened in system text editor, which may appear on another desktop

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
- [x] Fix _errors_ format in AbortWithErrorJSON. Most JS handlers are expecting an object, not an array or scalar. Need to check all handlers.
- [x] Implement validation-only jobs.
- [x] Implement upload-only jobs.
- [x] Replace calls to window.confirm() and window.alert() with a custom dialog because user may inadvertantly silence these dialogs.
- [x] Replace history.back() with proper back links! 
- [x] Ensure that jobs won't initiate if already running (prevent double-get request)
- [x] Display workflow batch error if user doesn't choose a CSV file
- [x] Fix question bug in export settings. Question select lists won't populate until settings are saved.
- [x] Clicking Settings > Export saves settings but does not export them. It should show the export JSON.
- [x] Allow import of DART-native profiles
- [x] Context-sensitive help
