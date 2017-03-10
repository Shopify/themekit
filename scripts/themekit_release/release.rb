# frozen_string_literal: true
require 'digest'

# Release represents a single platform release
class Release
  attr_reader :platform, :file, :version, :full_name, :data
  attr_accessor :location

  def initialize(version: nil, platform: nil, filename: nil)
    @platform = platform
    @version = version
    load_file(filename)
  end

  def hexdigest
    @hexdigest ||= Digest::MD5.hexdigest(data)
  end

  def as_hash
    {
      name: platform,
      url: location,
      digest: hexdigest
    }
  end

  private

  def load_file(filename)
    file = File.open(filename, 'rb')
    @full_name = [version, platform, File.basename(file.path)].join('/')
    @data = file.read
    file.close
  end
end
