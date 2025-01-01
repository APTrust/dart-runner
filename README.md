[![Build Status](https://travis-ci.com/APTrust/dart-runner.svg?branch=master)](https://travis-ci.org/APTrust/dart-runner)
[![Maintainability](https://api.codeclimate.com/v1/badges/afced50b57b1e02432f6/maintainability)](https://codeclimate.com/github/APTrust/dart-runner/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/afced50b57b1e02432f6/test_coverage)](https://codeclimate.com/github/APTrust/dart-runner/test_coverage)

# DART Runner

**DART Runner** executes [DART](https://github.com/APTrust/dart) jobs and workflows without a UI, allowing you to run BagIt packaging and file uploads on a server or headless environment.

## Key Features

- **Headless Operation**: No desktop dependencies or GUI required—ideal for servers.
- **Easy Workflow Definition**: Create or export workflows in the DART UI, then run them anywhere.
- **Flexible & Scriptable**: Integrate with cron jobs or scripts to automate bag creation and transfers.

## Downloads

- **Mac & Linux Beta**: [Download Links](https://aptrust.github.io/dart-docs/users/dart_runner/#downloads)
- **Windows**: (Coming soon or in progress—adjust as needed)

## Quick Usage

For immediate help, run:

```bash
dart-runner --help
```

or see the [Detailed DART Runner Docs](https://aptrust.github.io/dart-docs/users/dart_runner/) for usage, command-line flags, exit codes, and advanced examples.

## What Are Jobs & Workflows?

- **Job**: Creates and ships a **single bag** (files + metadata, packaged according to a BagIt profile).
- **Workflow**: Processes **multiple jobs** that share the same BagIt profile, tag values, and upload settings.

For example, a workflow can bag 300 folders and upload them to an S3 bucket.  
See [Jobs](https://aptrust.github.io/dart-docs/users/jobs/) and [Workflows](https://aptrust.github.io/dart-docs/users/workflows/) in the DART docs for more details.

## Why DART Runner?

DART (the Electron app) can run in a CLI mode but still requires a graphics stack on the machine. That’s not practical for servers. **DART Runner** solves this by providing a minimal binary with no graphics dependencies. After defining and testing a workflow in DART, you can export it and run it via DART Runner on virtually any Windows/Mac/Linux/Unix server.

### Typical Workflow Steps

1. Define and test your workflow in DART (desktop).
2. Export the workflow JSON.
3. Create a CSV listing files/folders to be bagged.
4. Copy the workflow JSON and CSV to your server.
5. Run `dart-runner --workflow=my_workflow.json --batch=my_batch.csv --output-dir=/some/dir`.
6. (Optional) Schedule the runner via cron or another scheduler for ongoing updates.

DART Runner supports advanced usage: replicating bags, daily/weekly archiving, or partial/incremental packaging.

## DART User Group

Join the [DART User Group](https://aptrust.org/resources/user-groups/dart-user-group/) to discuss use cases, ask questions, or get help. This group primarily uses a [mailing list](https://groups.google.com/a/aptrust.org/g/dart-users), with the possibility of virtual meetings if there’s enough interest.

---

# Developer Setup

If you’d like to build from source or contribute, follow these steps:

1. **Clone the Repository**
   ```bash
   git clone https://github.com/APTrust/dart-runner
   ```
2. **Install Dependencies**

   - **Docker**: Required to run SFTP tests.
   - **Go (1.20+)**: Install from [Go’s website](https://go.dev/).
   - **GCC**: Needed for race-detection tests. On Windows, see [these instructions](https://code.visualstudio.com/docs/cpp/config-mingw).

3. **Optional: VS Code Setup**
   - Install VS Code Go language support (extensions will prompt you).
   - [vscode-gotemplate](https://github.com/casualjim/vscode-gotemplate) handles Go HTML templates, preventing HTML lint errors.

## Building

Run either of the following:

```bash
./scripts/build.sh
```

or

```bash
bash ./scripts/build.sh
```

## Running (Local Dev)

Use Ruby to start DART and local services (S3/SFTP) for interactive testing:

```bash
ruby ./scripts/run.rb dart
```

- DART UI is at **http://localhost:8080**
- Local **Minio** for S3 uploads runs at `127.0.0.1:9899`
- Local **SFTP** server runs at `127.0.0.1:2222`

### Example Minio Config

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

### Example SFTP Config

Two sample credentials:

- **SSH Key**:
  ```json
  {
    "id": "2b0439bc-66d2-4d01-a73d-19d3eb9edf73",
    "allowsDownload": false,
    "allowsUpload": true,
    "bucket": "uploads",
    "description": "Local SFTP server using SSH key",
    "host": "127.0.0.1",
    "login": "key_user",
    "loginExtra": "/home/diamond/aptrust/dart-runner/testdata/sftp/sftp_user_key",
    "name": "Local SFTP (key)",
    "password": "",
    "port": 2222,
    "protocol": "sftp"
  }
  ```
- **Password**:
  ```json
  {
    "id": "d250eda9-d761-4c03-ab5b-266bacc40f3f",
    "allowsDownload": false,
    "allowsUpload": true,
    "bucket": "uploads",
    "description": "Local SFTP service using password",
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

```bash
ruby ./scripts/run.rb tests
```

- Requires **Docker** (24+), **Ruby** (3.0+), **Go** (1.20+), and **GCC** (for `-race` tests).

### Post-Build Test

```bash
./scripts/build.sh
./dist/dart-runner --workflow=./testdata/files/postbuild_test_workflow.json \
                   --batch=./testdata/files/postbuild_test_batch.csv \
                   --output-dir=<DIR>
```

---

## Contributing

Issues and PRs are welcome! See our [contributing guidelines](#) (if you have a dedicated CONTRIBUTING.md) or open an issue in GitHub.

## License

(Include your project’s license info/link here.)

## Further Resources

- **DART** (the GUI app): [GitHub Repo](https://github.com/APTrust/dart) | [User Guide](https://aptrust.github.io/dart-docs/)
- **DART Runner Docs**: [Official Docs](https://aptrust.github.io/dart-docs/users/dart-runner/)
- **Batch File Format**: [Reference](https://aptrust.github.io/dart-docs/users/workflows/batch_jobs/)
