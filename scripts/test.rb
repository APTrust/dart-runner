#!/usr/bin/env ruby
# coding: utf-8

require 'fileutils'

class TestRunner

  def initialize
    @start_time = Time.now
    @minio_pid = 0
  end

  def run
    make_test_dirs
    start_minio
    `go clean -testcache`
    cmd = "go test -p 1 ./... -coverprofile c.out"
    pid = Process.spawn(ENV, cmd, chdir: project_root)
    Process.wait pid
    exit_code = $?.exitstatus
    if $?.success?
      puts "😊 PASSED 😊"
      puts "To generate HTML report: > go tool cover -html=c.out"
    else
      puts "😡 FAILED 😡"
    end
    stop_minio
    exit(exit_code)
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

  def stop_minio
	if !@minio_pid
      puts "Pid for minio is zero. Can't kill that..."
	  return
	end
	puts "Stopping minio service (pid #{@minio_pid})"
	begin
	  Process.kill('TERM', @minio_pid)
	rescue
	  puts "Hmm... Couldn't kill #{@minio_pid}."
      puts "Check system processes to see if a version "
      puts "of that process is lingering from a previous test run."
	end
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
    ]
    buckets.each do |bucket|
      full_bucket = File.join(base, "minio", bucket)
      puts "Creating local minio bucket #{bucket}"
      FileUtils.mkdir_p full_bucket
    end
  end
end


if __FILE__ == $0
  test_runner = TestRunner.new
  test_runner.run
end
