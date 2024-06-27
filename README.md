[![Build Status](https://travis-ci.com/APTrust/dart-runner.svg?branch=master)](https://travis-ci.org/APTrust/dart-runner)
[![Maintainability](https://api.codeclimate.com/v1/badges/afced50b57b1e02432f6/maintainability)](https://codeclimate.com/github/APTrust/dart-runner/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/afced50b57b1e02432f6/test_coverage)](https://codeclimate.com/github/APTrust/dart-runner/test_coverage)

# What is dart-runner?

dart-runner will run [DART](https://github.com/APTrust/dart) jobs and workflows without requiring a UI. This means you can run DART workflows on a server.

## Downloads

For Mac and Linux beta versions, see  https://aptrust.github.io/dart-docs/users/dart_runner/#downloads

## Usage

Run `dart-runner --help`, or view the [detailed help page](https://aptrust.github.io/dart-docs/users/dart_runner/).

## Definitions

A [job](https://aptrust.github.io/dart-docs/users/jobs/) is the creation and shipping of a single bag. It typically involves bagging a list of files according to a BagIt profile and then sending that bag to an SFTP or S3-compliant server.

A [workflow](https://aptrust.github.io/dart-docs/users/workflows/) is a set of jobs that all follow the same pattern. For example, bag 300 folders according to the same BagIt profile and send them all to an S3 bucket in Wasabi.

[This video](https://aptrust.github.io/dart-docs/videos/) shows examples of jobs and workflows.

# Why build it?

Many DART users want to define a workflow to bag and ship hundreds or thousands of items and then let that workflow run unattended, either as a one-off process or on a daily/weekly/monthly schedule. This is by far our most requested feature.

A server-side job runner can work with existing systems such as Fedora, Hyrax, LOCKSS, Archivematica, etc. to periodically bag and push materials into a remote preservation system.

DART is an Electron app that can run in command-line mode, but the underlying Electron framework requires a UI and windowing system to be present on the OS before it starts up, even when it’s not going to use a graphical interface. This limitation is inherent in Electron, so DART cannot fix it.

Requiring a graphics system is fine on a desktop or laptop, but not on servers. It means installing more than 1 GB of dependencies, including the entire X window system and a desktop manager like Gnome or KDE. Electron requires the graphics system to be running, consuming considerable resources, even though no other program on the server uses it. (And even though Electron itself doesn't use it when running in command-line mode.)

This is too much to ask just to run DART on a server.

dart-runner will run any DART job or workflow on virtually any Windows, Mac, Linux or Unix computer, with no need for graphics capabilities. Installing the app would require only a single binary (no external libraries or dependencies) and a text-based configuration file.

Running the app would require:

* the application binary
* the config file
* access to the files you want to bag (through local disk or network attached storage)
* access to a network (to send files to remote S3 / SFTP servers)

To run a workflow on your server:

1. Define and test the workflow using the DART GUI on a desktop computer.
2. Choose your BagIt profile.
3. Define default tag values.
4. Define the locations to which bags should be shipped.
5. Run the workflow to test it.
6. Create a CSV file (easily done with Excel) to list all of the files/folders to be run through the workflow.
7. Copy the CSV file and the workflow description (a JSON file exported by DART) to your server and run.

Once you’ve defined a workflow and a CSV list, you can set the workflow to run as a daily/weekly/monthly cron job.

For more dynamic workflows where you want to bag and ship only new and updated items, you can run a script that generates a CSV file of new/updated items and feed that file to the DART Runner application.

dart-runner can also help replicate bags between digital repositories using the [Beyond the Repository BagIt format](https://github.com/dpscollaborative/btr_bagit_profile).

## DART User Group

APTrust hosts a [DART User Group](https://aptrust.org/resources/user-groups/dart-user-group/) for the entire digital preservation community. This group will primarily be a [mailing list](https://groups.google.com/a/aptrust.org/g/dart-users), where users can share experiences, ask questions, and support one another. Depending on the level of interest and engagement, we may expand this initiative to include regular virtual meetings and more structured activities in the future.

## Developer Setup

* Download the code using `git clone https://github.com/APTrust/dart-runner`.
* You will need to install [Docker](https://docs.docker.com/get-docker/) to run SFTP tests.
* You'll need GCC for testing. See the Testing section below.

## VS Code Setup

If you're using Visual Studio Code, you should install the following:

* VS Code Go language support. VS Code will prompt you to install the recommended components. Simply follow the recommendations to install Go, the Go language server and a few other standard components.
* [vscode-gotemplate](https://github.com/casualjim/vscode-gotemplate). This plugin properly handles Go HTML templates. It's not strictly necessary, but without it, VS Code will flag many Go template tags as invalid HTML and your project will be littered with unnecessary "problem" messages.

## Building

`./scripts/build.sh` or `bash ./scripts/build.sh`

## Running

To run DART interactively:

`ruby ./scripts/run.rb dart`

This will start DART on http://localhost:8080. It will also start a local Minio server to handle S3 uploads, and a local SFTP server for SFTP uploads. You can upload to these services using the following settings:

### Local Minio Service

```json
{
	"id": "d9ba0629-6870-48a3-9dd7-89e21410453b",
	"allowsDownload": true,
	"allowsUpload": true,
	"bucket": "test",
	"description": "Local Minio s3 service",
	"host": "127.0.0.1",
	"login": "minioadmin",
	"loginExtra": "",
	"name": "Local Minio",
	"password": "minioadmin",
	"port": 9899,
	"protocol": "s3"
}
```

### Local SFTP Service

You can connect to this with a password or an SSH key. You should keep both these entries in your local dev/test environment so you can do interactive testing with them. This one uses a key:

```json
{
	"id": "2b0439bc-66d2-4d01-a73d-19d3eb9edf73",
	"allowsDownload": false,
	"allowsUpload": true,
	"bucket": "uploads",
	"description": "Local SFTP server using SSH key for authentication",
	"host": "127.0.0.1",
	"login": "key_user",
	"loginExtra": "/home/diamond/aptrust/dart-runner/testdata/sftp/sftp_user_key",
	"name": "Local SFTP (key)",
	"password": "",
	"port": 2222,
	"protocol": "sftp"
}
```

And this uses a password:

```json
{
	"id": "d250eda9-d761-4c03-ab5b-266bacc40f3f",
	"allowsDownload": false,
	"allowsUpload": true,
	"bucket": "uploads",
	"description": "Local SFTP service using password authentication",
	"host": "127.0.0.1",
	"login": "pw_user",
	"loginExtra": "",
	"name": "Local SFTP (password)",
	"password": "password",
	"port": 2222,
	"protocol": "sftp"
}
```

## Testing

`ruby ./scripts/run.rb tests`

Note that in addition to having a recent version of Go (1.20+), running tests requires the following dependencies:

* A recent version of Ruby (3.0+)
* A recent version of Docker (24+)
* GCC, the GNU Compiler Collection, to enable race detection during tests. We do test with the `-race` flag. 
  On Windows, follow [these instructions](https://code.visualstudio.com/docs/cpp/config-mingw) to install GCC,
  and be sure to add the MINGW bin to your path, so GCC is always accessible.

### Post-Build Test

```
./scripts/build.sh
./dist/dart-runner --workflow=./testdata/files/postbuild_test_workflow.json --batch=./testdata/files/postbuild_test_batch.csv --output-dir=<DIR>
```
