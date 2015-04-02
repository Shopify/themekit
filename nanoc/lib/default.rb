# All files in the 'lib' directory will be loaded
# before nanoc starts compiling.
require 'open-uri'
require 'rss'
include Nanoc::Helpers::LinkTo

VERSIONS = [
  {platform: 'Mac OS X',       hidden: false, cls: 'mac', zip_name: 'darwin-amd64'},
  {platform: '64-bit Windows', hidden: false, cls: 'win', zip_name: 'windows-amd64'},
  {platform: '64-bit Linux',   hidden: false, cls: 'lin', zip_name: 'linux-amd64'},
  {platform: '32-bit Windows', hidden: true,  cls: 'win', zip_name: 'windows-386'},
  {platform: '32-bit Linux',   hidden: true,  cls: 'lin', zip_name: 'linux-386'}
]

def download_url_for(version)
  "#{download_url}/#{version[:zip_name]}.zip"
end

def classes_for(version)
  classes = ["version", version[:cls]]
  if version[:hidden]
    classes << "col-1-2"
    classes << "hidden"
  else
    classes << "col-1-3"
  end
  classes.join(" ")
end

private

def download_url
  "#{@config[:repository]}/releases/download/#{latest_release_version}"
end

def latest_release_version
  latest_release_path.split("/").last
end

def latest_release_path
  feed.entries.first.link.href
end

def feed
  @feed ||= begin
    content = open("#{@config[:repository]}/releases.atom")
    RSS::Parser.parse(content)
  end
end
