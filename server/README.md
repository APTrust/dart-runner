# DART 3

DART 3 is currently in ALPHA mode. We encourage you to use it for testing and to report bugs and feature requests at https://github.com/APTrust/dart-runner/issues.

DO NOT USE THE ALPHA VERSION FOR ESSENTIAL PRODUCTION WORKFLOWS! Wait for a stable release build if you want to use this in production.

## Known Issues

Major features are generally known to work in the current alpha build. However, the build has some known issues, including:

* occasional freezing of some dynamic menus and other UI items
* limited testing on Windows

## Getting Started

1. Dowload the app

| Operating System       | Download Link |
| ---------------------- | ------------- |
| Windows (Intel 64-bit) | |
| Mac (M chips)          | |
| Mac (Intel chips)      | |
| Linux (Intel 64-bit)   | |

2. Open a terminal window and change into the directory containing the dart3 download. 

3. Make the app excetable with this command: `chmod +x dart3`

4. Run the app with this command `./dart3` (Note the leading dot and slash.)

5. Open a browser and go to __http://localhost:8080__

If you want to run DART on a port other than 8080, start it with this command: `./dart3 -port <number>` where is number is any port number you choose. Number should be above 1024 on most systems, because ports below that may be reserved or require root privileges.

## Platform Rationale

The server component of DART runner will be the successor to DART 2. DART 2 is an Electron app that whose maintenance has been time consuming and difficult. We chose to write DART 3 in Go as a locally-running web app for a number of reasons, including:

* The tasks DART has to perform can be written much more clearly in Go than in JavaScript. This substantially eases our maintenance burden and makes it easier to add new features. Node's default async model is particularly ill-suited to some of DART's core tasks. (E.g. writing tar files, which MUST be done synchronously.)
* The Go ecosystem is more stable than the Electron/Node ecosystem. We know this from years of maintaining code on both platforms.
* Node and Electron often introduce breaking features in new releases, forcing us to abandon and rewrite working code. The Go language and the major browsers rarely do this.
* The rest of the APTrust ecosystem is written in Go, which allows us to reuse proven code for bagging, validation and file transport. This substantially reduces the burden of having to maintain complex code in two different languages (Go and JavaScript) with identical functionality and behavior.
* Electron apps like DART use substantial resources. Running the DART test suite consumed about 1.5 GB of RAM. DART 3 uses about 14 MB of RAM and considerably less CPU.
* Electron builds did not always behave the same way as Electron in the development environment. Spending days of developer time to debug these issues was a poor use of developer time.

We evaluated a number of platforms similar to Electron that would allow us to use Go instead of JavaScript for the heavy work. The most promising of these was [Wails](https://wails.io/), but in our early tests in 2022, we experienced some crashes and blank screens, and we didn't feel the platform was mature enough.

We decided to go with the simplest and most reliable technologies available, where are a basic web server and whichever browser the user prefers.

## Notes for Developers

Testing: `./scripts/run.rb tests`

Building for realease: `./scripts/build_dart.rb`

Running in dev mode: `./scripts/run.rb dart`

Note that running in dev mode also starts a local SFTP server and a Minio server, both in docker containers. DART will print the URLs and credentials for these local services in the console so you can look into them if necessary.

## Prerequisites for Development

* Go > 1.20
* Ruby > 2.0 (to run build and test scripts)
* Docker (to run Minio and SFTP containers)

