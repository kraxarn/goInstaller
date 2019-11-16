# goInstaller: Generic installer written in Go
(not just for applications written in Go)

## How it installs
* Downloads set files
* Extracts (if zip) or move downloded files to a specific location
* Creates a shortcut to the executable

## Missing/planned features
* Executing downloaded files (for runtime dependencies for example)
* Automatically downloading the latest version from GitHub

## Why?
I basically just wanted a basic installer for my own projects and an installer that lets me use the same executable, even if I update the software.

## Why Go?
Just makes it easier to port to other platforms and somewhat gets rid of the pain of compiling anything on Windows (where an installer is most needed). I also think it is important to have an application with no runtime dependencies, so you do not need an installer for the installer.