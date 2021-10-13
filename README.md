[![Build Status](https://travis-ci.com/APTrust/dart-runner.svg?branch=master)](https://travis-ci.org/APTrust/dart-runner)
[![Maintainability](img src="https://api.codeclimate.com/v1/badges/afced50b57b1e02432f6/maintainability)](https://codeclimate.com/github/APTrust/dart-runner/maintainability)


# What is dart-runner?

dart-runner will run [DART](https://github.com/APTrust/dart) jobs and workflows without requiring a UI. This means you can run DART workflows on a server.

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
