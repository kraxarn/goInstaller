# goInstaller: Generic installer written in Go
(not just for applications written in Go)

![](https://i.vgy.me/5qjQO6.png)

## How it installs
* Downloads set files
* Extracts (if zip) or move downloded files to a specific location
* Creates a shortcut to the executable

## Missing/planned features
* Executing downloaded files (for runtime dependencies for example)
* Automatically downloading the latest version from GitHub
* Does not come with an uninstaller
* On Windows, nothing is added to the control panel

## OS Support
The installer is mostly targeted towards Windows, but also works fine on Linux.
There is some support for macOS, but due to how different the platform is when it comes to 
application structure, it may not work very well. Instead, it is recommended to just distribute 
a single zip file with the app file inside of it. macOS has also stopped supporting OpenGL, 
which this application relies on, which may cause issues on newer versions of macOS.

## Why are there no binary releases?
Currently, you need to edit the first few lines in the `main.go` file to customize for your
application. It does not make any sense to distrubute binary releases when you first need
to edit the source code to make it work. This may change in the future (like v2.0).

## Where is the documentation?
Proper documentation coming soon, but just check the top of the `main.go` file for now.

## Why?
I basically just wanted a basic installer for my own projects and an installer that lets me use the same executable, even if I update the software.

## Why Go?
Just makes it easier to port to other platforms and somewhat gets rid of the pain of compiling anything on Windows (where an installer is most needed). I also think it is important to have an application with no runtime dependencies, so you do not need an installer for the installer.