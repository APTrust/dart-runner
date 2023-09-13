#!/usr/bin/env ruby
# coding: utf-8

require 'fileutils'

class Runner

  def initialize
    @start_time = Time.now
    @minio_pid = 0
    @dart_pid = 0
    @docker_sftp_id = ''
    @sftp_started = false
  end

  def run_tests
    make_test_dirs
    start_minio
    start_sftp
    `go clean -testcache`
    cmd = "go test -race -p 1 ./... -coverprofile c.out"
    pid = Process.spawn(ENV, cmd, chdir: project_root)
    Process.wait pid
    exit_code = $?.exitstatus
    if $?.success?
      puts "😊 PASSED 😊"
      puts "To generate HTML report: > go tool cover -html=c.out"
    else
      puts "😡 FAILED 😡"
    end
    exit(exit_code)
  end

  def run_dart
    begin
      @dart_pid = Process.spawn(ENV, "go run -race dart/main.go", chdir: project_root)
      sleep(1)
      puts "\n\n\n"
      puts "DART is running at http://localhost:8080"
      start_minio
      start_sftp
      puts "\n"
      puts "Control-C will stop DART, SFTP and Minio\n\n"
      Process.wait @dart_pid
    rescue SystemExit, Interrupt
      puts "\nEt tu, Brute! Then fall, Caesar."
    end
  end

  def start_minio
    bin = self.bin_dir
    minio_cmd = "#{bin}/minio server --quiet --address=localhost:9899 ~/tmp/minio"
    log_file = log_file_path("minio")
    puts "Minio is running on localhost:9899. User/Pwd: minioadmin/minioadmin"
    @minio_pid = Process.spawn(ENV, minio_cmd, out: log_file, err: log_file)
    Process.detach @minio_pid
    puts "Minio PID is #{@minio_pid} logging to #{log_file}"
  end

  def stop_dart
    if !@dart_pid
        puts "Pid for DART is zero. Can't kill that..."
        return
    end
    puts "Stopping DART service (pid #{@dart_pid})"
    begin
      Process.kill('TERM', @dart_pid)
    rescue
      puts "Could not kill DART. :("
    end
  end

  def stop_minio
    if !@minio_pid
        puts "Pid for Minio is zero. Can't kill that..."
        return
    end

    puts "Stopping minio service (pid #{@minio_pid})"

    begin
      Process.kill('TERM', @minio_pid)
    rescue
      # We'll handle this below
    end

    ps_pid = `ps -ef | grep minio`.split(/\s+/)[1].to_i
    if (ps_pid > 0)
      begin
        Process.kill('TERM', ps_pid)
        puts "Also stopped minio child process #{ps_pid}"
      rescue
        puts "Couldn't kill minio."
        puts "Check system processes to see if a version "
        puts "of that process is lingering from a previous test run."
        end
    end
  end

  # This command starts a docker container that runs an SFTP service.
  # We use this to test SFTP uploads.
  #
  # The first -v option sets #{sftp_dir}/sftp_user_key.pub as the public
  # key for user "key_user" inside the docker container. We set this so
  # we can test connections the use an SSH key.
  #
  # The second -v option tells the container to create accounts for the
  # users listed in #{sftp_dir}/users.conf. There are two. key_user has
  # no password and will connect with an SSH key. pw_user will connect
  # with the password "password".
  #
  # We forward local port 2222 to the container's port 22, which means we
  # can get to the SFTP server via locahost:2222 or 127.0.0.1:2222.
  def start_sftp
    sftp_dir = File.join(project_root, "testdata", "sftp")
    puts "Using SFTP config options from #{sftp_dir}"
    @docker_sftp_id = `docker run \
    -v #{sftp_dir}/sftp_user_key.pub:/home/key_user/.ssh/keys/sftp_user_key.pub:ro \
    -v #{sftp_dir}/users.conf:/etc/sftp/users.conf:ro \
    -p 2222:22 -d atmoz/sftp`
    if $?.exitstatus == 0
      puts "Started SFTP server with id #{@docker_sftp_id}"
      @sftp_started = true
    else
      puts "Error starting SFTP docker container. Is one already running?"
      puts @docker_sftp_id
    end
  end

  def stop_sftp
    if @sftp_started
      result = `docker stop #{@docker_sftp_id}`
      if $?.exitstatus == 0
        puts "Stopped docker SFTP service"
      else
        puts "Failed to stop docker SFTP service with id #{@docker_sftp_id}"
        puts "See if you can kill it."
        puts "Hint: run `docker ps` and look for the image named 'atmoz/sftp'"
      end
    else
      puts "Not killing SFTP service because it failed to start"
    end
  end

  def stop_all_services
    if @dart_pid > 0
      stop_dart
    end
    stop_minio
    stop_sftp
  end

  def project_root
    File.expand_path(File.join(File.dirname(__FILE__), ".."))
  end

  def bin_dir
    os = ""
    if RUBY_PLATFORM =~ /darwin/
      os = "osx"
    elsif RUBY_PLATFORM =~ /linux/
      os = "linux"
    else
      abort("Unsupported platform: #{RUBY_PLATFORM}")
    end
    File.join(project_root, "bin", os)
  end

  def log_file_path(service_name)
    log_dir = File.join(ENV['HOME'], "tmp", "logs")
    FileUtils.mkdir_p(log_dir)
    return File.join(log_dir, service_name + ".log")
  end

  def make_test_dirs
    base = File.join(ENV['HOME'], "tmp")
    if base.end_with?("tmp") # So we don't delete anyone's home dir
      puts "Deleting #{base}"
    end
    FileUtils.remove_dir(base ,true)
    dirs = ["bags", "bin", "logs", "minio"]
    dirs.each do |dir|
      full_dir = File.join(base, dir)
      puts "Creating #{full_dir}"
      FileUtils.mkdir_p full_dir
    end
    # S3 buckets for minio. We should ideally read these from the
    # .env.test file.
    buckets = [
      "dart-runner.test",
      "test",
    ]
    buckets.each do |bucket|
      full_bucket = File.join(base, "minio", bucket)
      puts "Creating local minio bucket #{bucket}"
      FileUtils.mkdir_p full_bucket
    end
  end

  def show_help
    puts "To run unit and integration tests:"
    puts "    run.rb tests\n"
    puts "To run DART, SFTP and Minio for interactive testing:"
    puts "    run.rb dart\n"
  end
end


if __FILE__ == $0
  runner = Runner.new
  at_exit { runner.stop_all_services }
  action = ARGV[0]
  if action == "tests"
    runner.run_tests
  elsif action == "dart"
    runner.run_dart
  else
    runner.show_help
  end
end
