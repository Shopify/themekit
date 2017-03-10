# frozen_string_literal: true
require 'git'
require 'semverly'

# Repository is used for checking the current tag is the most recent
class Repository
  GIT_DIR = File.expand_path(__FILE__ + '/../../..')

  def initialize
    @git = Git.open(GIT_DIR)
    ensure_head_is_at_latest_version
  end

  def latest_version
    latest_tag.name
  end

  private

  def latest_tag
    @git.tags.sort { |a, b| SemVer.parse(a.name) <=> SemVer.parse(b.name) }.last
  end

  def ensure_head_is_at_latest_version
    return if @git.object('HEAD').sha == latest_tag.sha
    puts "Your current HEAD does not match tag #{latest_version}."
    puts 'Verify you are at the right commit or create a new tag.'
    exit(1)
  end
end
