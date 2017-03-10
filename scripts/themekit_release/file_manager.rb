# frozen_string_literal: true

# FileManager handles a single file upload for a release
class FileManager
  ALL_RELEASES = 'releases/all.json'
  LATEST_RELEASE = 'releases/latest.json'

  attr_reader :bucket

  def initialize(bucket)
    @bucket = bucket
  end

  def upload!(filename, content)
    file = bucket.files.new(
      key: filename,
      body: content,
      public: true
    )
    file.save
  end

  def url(filename)
    get_file(filename).public_url
  end

  def fetch(filename, default: '')
    exists?(filename) ? get_file(filename).body : default
  end

  private

  def get_file(name)
    bucket.files.get(name)
  end

  def exists?(name)
    bucket.files.head(name)
  rescue
    false
  end
end
