# How to build OttoMap from scratch

This project is written in Go, a statically typed, compiled programming language designed for simplicity and efficiency.
Building the project is easy because Go is a cross-platform language, and the build process is straightforward.

The basic steps are:

1. Install the build tools (Git and Go)
2. Use Git to clone the source code
3. Use Go to build the executable

This document will walk you through the process for Windows, macOS, and Linux.
Please use the Discord server if you have questions or run into issues.

Here are some helpful definitions if you're new to building Go programs:

- **Go** is a programming language.
- **Repository** is a TODO.
- **Clone** is TODO.
- **Build tool** is a program that automates the process of building a program.

## Prerequisites

Before you begin, ensure that you have the following tools installed:

- Git (to download the source and keep it up to date)

- Go (to build the project)

You can install these from their official websites, but I recommend using a package manager to install them.
It will make it easier to keep them up to date.

Popular package managers are `winget` (Windows), `snap` (Linux), and `brew` (macOS).

* For more information about `winget`, visit the [Windows Package Manager](https://learn.microsoft.com/en-us/windows/package-manager/) site.
* For more information about `snap`, visit the [Snapcraft](https://snapcraft.io/docs/installing-snapd) site.
* For more information about `brew`, visit the [Homebrew](https://brew.sh/) site.

## Git

Git is used to download the source code and keep it up to date with the repository on GitHub.
You can download a ZIP file from the GitHub repository, but that's not the easiest way to get the code and stay up to date with the latest changes.

I started writing the steps, but the instructions on GitHub are quite good.
Please visit their [Install Git](https://github.com/git-guides/install-git) and follow the instructions for your operating system to install Git.

## Go

Go is used to build the OttoMap executable.
The installation steps depend on your operating system and if you're using a Package Manager.

### Winget (Windows)

Using `winget` (Windows Package Manager):

1. Open a command prompt.
2. Run the following command:

```
winget install Go
```

### Snap (Linux)

Using `snap` (Linux Package Manager):

1. Open a terminal.
2. Run the following command:

```
sudo snap install go --classic
```

### Brew (macOS)

Using `brew` (macOS Package Manager):

1. Open a terminal.
2. Run the following command:

```
brew install go
```

### The Official Installer

Visit the [Download and install](https://go.dev/doc/install) page and follow the instructions to download and install Go for your operating system.

## Clone the Repository

You need a copy of the source code before you can build the project.
You can use Git to clone it (recommended) or download a ZIP file containing the source code.

If you have Git installed, you can clone the repository using the following steps:

1. Open your terminal or command prompt.
2. Navigate to the directory where you want to clone the repository.
3. Run the following command to clone the repository:

```bash
git clone https://github.com/mdhender/ottomap.git
```

If you prefer to work from a ZIP file, the official OttoMap repository is hosted on GitHub at https://github.com/mdhender/ottomap.

## Building

The build process is the same on all the operating systems:

1. Open your terminal or command prompt.
2. Navigate to the directory with the source code.
3. Run the following command to build:

```bash
go build
```

This creates an executable named `ottomap` (macOS and Linux) or `ottomap.exe` (Windows) in the current directory.

Please use the Discord server if you have questions or issues building OttoMap.

The Go team has great documentation; check out https://go.dev/doc/tutorial/compile-install.
