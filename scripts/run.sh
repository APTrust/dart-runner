#!/bin/bash
#
# Run tests or services for dart-runner.

MINIO_STARTED=false
SFTP_STARTED=false
DOCKER_MINIO_ID=""
DOCKER_SFTP_ID=""

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

MINIO_USER="minioadmin"
MINIO_PASSWORD="minioadmin"

sftp_image_name() {
    if [ "$(uname -m)" = "arm64" ]; then
        echo "jmcombs/sftp"
    else
        echo "atmoz/sftp"
    fi
}

make_test_dirs() {
    local base="$HOME/tmp"
    if [[ "$base" == */tmp ]]; then
        echo "Deleting $base"
        rm -rf "$base"
    fi
    for dir in bags bin logs minio; do
        local full_dir="$base/$dir"
        echo "Creating $full_dir"
        mkdir -p "$full_dir"
    done
}

start_minio() {
    echo "Starting Minio container"
    DOCKER_MINIO_ID=$(docker run -p 9899:9000 -p 9001:9001 -v ~/tmp/minio:/data -e MINIO_ROOT_USER="$MINIO_USER" -e MINIO_ROOT_PASSWORD="$MINIO_PASSWORD" -d quay.io/minio/minio server /data --console-address ":9001")
    local exit_code=$?
    DOCKER_MINIO_ID=$(echo "$DOCKER_MINIO_ID" | tr -d '[:space:]')
    if [ $exit_code -eq 0 ]; then
        echo "Started Minio server with id $DOCKER_MINIO_ID"
        echo "Minio is running on localhost:9899. User/Pwd: ${MINIO_USER}/${MINIO_PASSWORD}"
        echo "Minio console available at http://127.0.0.1:9001"
        MINIO_STARTED=true
        echo "Waiting for Minio to be ready..."
        local attempts=0
        until curl -sf http://localhost:9899/minio/health/ready > /dev/null 2>&1; do
            attempts=$((attempts + 1))
            if [ $attempts -ge 30 ]; then
                echo "Minio did not become ready in time"
                break
            fi
            sleep 1
        done
        docker exec "$DOCKER_MINIO_ID" mc alias set local http://localhost:9000 "$MINIO_USER" "$MINIO_PASSWORD" > /dev/null 2>&1
        echo "Creating Minio buckets..."
      # Make our two test buckets, plus receiving and
      # restoration buckets for test.edu.
      docker exec $DOCKER_MINIO_ID mc mb local/test
      docker exec $DOCKER_MINIO_ID mc mb local/dart-runner.test
      docker exec $DOCKER_MINIO_ID mc mb local/preservation-or
      docker exec $DOCKER_MINIO_ID mc mb local/preservation-va
      docker exec $DOCKER_MINIO_ID mc mb local/glacier-oh
      docker exec $DOCKER_MINIO_ID mc mb local/glacier-or
      docker exec $DOCKER_MINIO_ID mc mb local/glacier-va
      docker exec $DOCKER_MINIO_ID mc mb local/glacier-deep-oh
      docker exec $DOCKER_MINIO_ID mc mb local/glacier-deep-or
      docker exec $DOCKER_MINIO_ID mc mb local/glacier-deep-va
      docker exec $DOCKER_MINIO_ID mc mb local/wasabi-or
      docker exec $DOCKER_MINIO_ID mc mb local/wasabi-tx
      docker exec $DOCKER_MINIO_ID mc mb local/wasabi-va
      docker exec $DOCKER_MINIO_ID mc mb local/receiving
      docker exec $DOCKER_MINIO_ID mc mb local/staging
      docker exec $DOCKER_MINIO_ID mc mb local/aptrust.receiving.test.test.edu
      docker exec $DOCKER_MINIO_ID mc mb local/aptrust.restore.test.test.edu
      docker exec $DOCKER_MINIO_ID mc mb local/aptrust.receiving.test.institution1.edu
      docker exec $DOCKER_MINIO_ID mc mb local/aptrust.restore.test.institution1.edu
      docker exec $DOCKER_MINIO_ID mc mb local/aptrust.receiving.test.institution2.edu
      docker exec $DOCKER_MINIO_ID mc mb local/aptrust.restore.test.institution2.edu
      docker exec $DOCKER_MINIO_ID mc mb local/aptrust.receiving.test.example.edu
      docker exec $DOCKER_MINIO_ID mc mb local/aptrust.restore.test.example.edu
    else
        echo "Error starting Minio docker container. Is one already running?"
        echo "$DOCKER_MINIO_ID"
    fi
}

stop_minio() {
    if [ "$MINIO_STARTED" = "true" ]; then
        docker stop "$DOCKER_MINIO_ID"
        if [ $? -eq 0 ]; then
            echo "Stopped docker Minio service"
        else
            echo "Failed to stop docker Minio service with id $DOCKER_MINIO_ID"
            echo "See if you can kill it."
            echo "Hint: run \`docker ps\` and look for the image named 'minio/minio'"
        fi
    else
        echo "Not killing Minio service because it failed to start"
    fi
}

start_sftp() {
    local sftp_dir="$PROJECT_ROOT/testdata/sftp"
    local image
    image=$(sftp_image_name)
    echo "Using SFTP config options from $sftp_dir"
    DOCKER_SFTP_ID=$(docker run \
        -v "$sftp_dir/sftp_user_key.pub:/home/key_user/.ssh/keys/sftp_user_key.pub:ro" \
        -v "$sftp_dir/users.conf:/etc/sftp/users.conf:ro" \
        -p 2222:22 -d "$image")
    local exit_code=$?
    DOCKER_SFTP_ID=$(echo "$DOCKER_SFTP_ID" | tr -d '[:space:]')
    if [ $exit_code -eq 0 ]; then
        echo "Started SFTP server with id $DOCKER_SFTP_ID"
        echo "To log in and view the contents, use"
        echo "sftp -P 2222 pw_user@localhost"
        echo "The password is 'password' without the quotes"
        SFTP_STARTED=true
    else
        echo "Error starting SFTP docker container. Is one already running?"
        echo "$DOCKER_SFTP_ID"
    fi
}

stop_sftp() {
    if [ "$SFTP_STARTED" = "true" ]; then
        docker stop "$DOCKER_SFTP_ID"
        if [ $? -eq 0 ]; then
            echo "Stopped docker SFTP service"
        else
            echo "Failed to stop docker SFTP service with id $DOCKER_SFTP_ID"
            echo "See if you can kill it."
            echo "Hint: run \`docker ps\` and look for the image named 'atmoz/sftp'"
        fi
    else
        echo "Not killing SFTP service because it failed to start"
    fi
}

stop_all_services() {
    stop_minio
    stop_sftp
}

run_tests() {
    make_test_dirs
    start_minio
    start_sftp
    go clean -testcache
    cd "$PROJECT_ROOT"
    go test -race -p 1 ./... -coverprofile c.out
    local exit_code=$?
    if [ $exit_code -eq 0 ]; then
        echo "PASSED"
        echo "To generate HTML report: > go tool cover -html=c.out"
    else
        echo "FAILED"
    fi
    exit $exit_code
}

show_help() {
    echo "To run unit and integration tests:"
    echo "    run.sh tests"
    echo ""
    echo "To start SFTP and Minio containers for interactive testing:"
    echo "    run.sh services"
    echo ""
}

trap stop_all_services EXIT

action="${1}"

case "$action" in
    tests)
        run_tests
        ;;
    services)
        make_test_dirs
        start_minio
        start_sftp
        echo "Control-C to quit"
        while true; do
            sleep 1
        done
        ;;
    *)
        show_help
        ;;
esac
