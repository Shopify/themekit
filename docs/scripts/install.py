#!/usr/bin/env python
'''
    File name: install
    Author: Tim Anema
    Date created: Sep 29, 2016
    Date last modified: Sep 14 2018
    Python Version: 2.7
    Description: Install script for themekit. It will download a release and make it executable
'''
import os, urllib, json, sys, hashlib

class Installer(object):
    LATEST_RELEASE_URL = "https://shopify-themekit.s3.amazonaws.com/releases/latest.json"
    ARCH_MAPPING = {
        "darwin x86_64": "darwin-amd64",
        "darwin i386": "darwin-386",
        "linux x86_64": "linux-amd64",
        "linux i386": "linux-386",
        "freebsd x86_64": "freebsd-amd64",
        "freebsd i386": "freebsd-386"
    }

    def __init__(self, path="/usr/local/bin"):
        self.install_path = os.path.expanduser(path)
        self.bin_path = "%s/theme" % self.install_path
        self.arch = self.__getArch()
        print("Fetching release data")
        self.release = json.loads(urllib.urlopen(Installer.LATEST_RELEASE_URL).read().decode("utf-8"))
        print("Downloading version %s of Shopify Themekit" % self.release['version'])
        self.__download()
        print("Theme Kit has been installed at %s" % self.bin_path)
        print('To verify themekit is working simply type "theme"')

    def __getArch(self):
        pipe = os.popen("echo \"$(uname) $(uname -m)\"")
        arch_name = pipe.readline().strip().lower()
        pipe.close()
        if arch_name not in Installer.ARCH_MAPPING:
            print("Cannot find binary to match your architecture [%s]" % arch_name)
            sys.exit("Please open an issue at https://github.com/Shopify/themekit/issues")
        return Installer.ARCH_MAPPING[arch_name]

    def __findReleasePlatform(self):
        for index, platform in enumerate(self.release['platforms']):
            if platform['name'] == self.arch:
                return platform

    def __download(self):
        platform = self.__findReleasePlatform()
        data = urllib.urlopen(platform['url']).read()
        if hashlib.md5(data).hexdigest() != platform['digest']:
            sys.exit("Downloaded binary did not match checksum.")
        else:
            print("Validated binary checksum")
        if not os.path.exists(self.install_path):
            os.makedirs(self.install_path)
        with open(self.bin_path, "wb") as themefile:
            themefile.write(data)
        os.chmod(self.bin_path, 0o755)

if sys.version_info[0] < 3:
    Installer()
else:
    sys.exit("Python 2 is required for this script.")
