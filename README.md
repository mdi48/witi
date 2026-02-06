# WITI (Why Is This Installed?)
A simple cli tool to give you information regarding the installation status of a given package. It will
tell you if it was installed explicitly, or as a result of another package. It will also list what the package's dependencies are and which installed packages require it. More importantly, however, is that it will show the installation chain of how it **may** have come to be installed in the first place (more useful for finding what packages are redundant or measuring their degree of importance if anything. Don't bother trying to use it with any GNU tools).

This is just a simple tool I made as a project to help me learn Go, and is intended for my own personal use (at least for now), and only supports the pacman package manager as a result (again, this might change in future.) There will, of course, be future improvements to this regardless of what I choose to do with it.

Also, shoutout to [witr](https://github.com/pranshuparmar/witr) for the name inspiration. It's a fantastic tool that I have found incredibly helpful.