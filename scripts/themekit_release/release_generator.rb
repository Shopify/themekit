# frozen_string_literal: true
require 'json'

# ReleaseGenerator makes sure a release was nessecary and uploading it if it is.
class ReleaseGenerator
  attr_reader :version, :releases, :storage_manager

  DIST_DIR = File.expand_path(__FILE__ + '/../../../build/dist')
  BUILDS = %w(
    darwin-amd64
    linux-386
    linux-amd64
    windows-386
    windows-amd64
  ).freeze

  def initialize(version, storage_manager)
    ensure_builds_have_been_created
    @version = version
    @releases = prepare_releases
    @storage_manager = storage_manager
  end

  def upload!
    ensure_release_is_necessary
    releases.each do |release|
      puts "  - Uploading #{release.full_name}"
      storage_manager.upload!(release.full_name, release.data)
      release.location = storage_manager.url(release.full_name)
    end
  end

  private

  def ensure_builds_have_been_created
    return if File.exist?(DIST_DIR)
    puts "Distribution build at #{DIST_DIR} has not been created. \
          Run 'make dist' before attempting to create a new release"
    exit(1)
  end

  def ensure_release_is_necessary
    feed = JSON.parse(
      storage_manager.fetch(FileManager::ALL_RELEASES, default: [].to_json)
    )
    return unless feed.find { |r| r['version'] == version }

    puts "v#{version} has already been deployed. If this was intended to be \
    a new release, ensure version.go has been updated and add an appropriate \
    tag to git"
    exit(1)
  end

  def prepare_releases
    BUILDS.map do |platform|
      build_location = File.expand_path(DIST_DIR + "/#{platform}")
      binary = Dir.entries(build_location).find { |name| name =~ /theme/ }

      Release.new(
        version: version,
        platform: platform,
        filename: [build_location, binary].join('/')
      )
    end
  end
end
