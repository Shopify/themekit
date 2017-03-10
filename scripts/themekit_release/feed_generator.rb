# frozen_string_literal: true
require 'json'

# FeedGenerator updates the release feeds
class FeedGenerator
  attr_reader :version, :releases, :storage_manager

  def initialize(version, releases, storage_manager)
    @version = version
    @releases = releases
    @storage_manager = storage_manager
  end

  def upload!
    storage_manager.upload!(FileManager::LATEST_RELEASE, latest_feed)
    storage_manager.upload!(FileManager::ALL_RELEASES, entire_feed)
  end

  private

  def latest_feed
    as_hash.to_json
  end

  def entire_feed
    full_feed = JSON.parse(
      storage_manager.fetch(FileManager::ALL_RELEASES, default: [].to_json)
    )
    full_feed << as_hash
    full_feed.uniq.to_json
  end

  def as_hash
    {
      version: version,
      platforms: releases.map(&:as_hash)
    }
  end
end
