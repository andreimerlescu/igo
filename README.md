# IGO

This package is called `IGO` which stands for **Install Go**!

This program is written in Go and designed to install Go runtime.

You won't install this program in your traditional go-way... 

```bash
go install github.com/andreimerlescu/igo@latest
```

If you have go already installed, and you attempt to use igo to manage
your installation of Go, you can break your system configurations. 

Download the binaries and use them to install go!
    
    # MacOS Apple Silicon
    curl -L https://github.com/andreimerlescu/igo/releases/download/v1.1.0/igo-darwin-arm64 ~/bin/igo
    # MacOS Apple Intel
    curl -L https://github.com/andreimerlescu/igo/releases/download/v1.1.0/igo-darwin-amd64 ~/bin/igo
    # Linux arm64
    curl -L https://github.com/andreimerlescu/igo/releases/download/v1.1.0/igo-linux-arm64 ~/bin/igo
    # Linux amd64
    curl -L https://github.com/andreimerlescu/igo/releases/download/v1.1.0/igo-linux-amd64 ~/bin/igo

    # Remove Apple Quarantine Blocker
    command -v xattr 2> /dev/null && xattr -d com.apple.quarantine ~/bin/igo

    # Set Permissions
    chmod +x ~/bin/igo && export PATH=~/bin:$PATH

    # Use IGO!
    igo -l

        igo [open source at github.com/ProjectApario/igo]
        ┌──────────┬──────────────────┬─────────────┐                                                                                     
        │ VERSION  │     CREATION     │   STATUS    │
        ├──────────┼──────────────────┼─────────────┤
        │ 1.24.3   │ 2025-05-22 11:15 │             │
        │ 1.24.0   │ 2025-05-23 09:37 │             │
        │ 1.23.4   │ 2025-05-22 11:30 │             │
        │ 1.23.2   │ 2025-05-22 11:24 │             │
        │ 1.23.0   │ 2025-05-22 11:23 │             │
        │ 1.22.7   │ 2025-06-02 22:49 │             │
        │ 1.22.6   │ 2025-06-02 22:47 │  * ACTIVE   │
        │ 1.22.5   │ 2025-05-24 09:59 │             │
        ├──────────┼──────────────────┼─────────────┤
        │ I ❤ YOU! │ Made In America  │ Be Inspired │
        └──────────┴──────────────────┴─────────────┘


## Usage

    igo -v
    v1.1.0 - igo open source at github.com/ProjectApario/igo

    igo -l # list (lowercase "L")
    igo -e # show environment
    igo -a <version> # activate <version> if its installed
    igo -s <version> # switch to <version> if its installed (alias to activate)
    igo -f <version> # fix <version> installation
    igo -u <version> # uninstall <version> from -godir <path>
    igo -i <version> # install <version> from -godir <path>

    # custom godir with debug
    igo -i 1.23.4 -godir /Shared/go -debug

Additional arguments include: 

| Argument       | Kind   | Usage                | Notes                                         | 
|----------------|--------|----------------------|-----------------------------------------------|
| `-i <version>` | String | `igo -i 1.23.4`      | Installs `go` version **1.23.4**.             |
| `-u <version>` | String | `igo -u 1.23.4`      | Uninstall `go` version **1.23.4**.            |
| `-s <version>` | String | `igo -s 1.24.2`      | Switch to version **1.24.2**                  |
| `-f <version>` | String | `igo -f 1.24.2`      | Fixes installation of **1.24.2**              |
| `-a <version>` | String | `igo -a 1.24.2`      | Activates go version **1.24.2**               |
| `-e`           | Bool   | `igo -e`             | Display's environment of active installations |
| `-l`           | Bool   | `igo -l`             | List all installed Go versions                | 
| `-v`           | Bool   | `igo -v`             | Display version                               | 
| `-version`     | Bool   | `igo -version`       | Display `igo` binary version.                 |
| `-godir`       | String | `igo -godir /opt/go` | Installs `igo` in `/opt/go`.                  |
| `-goos`        | String | `igo -goos linux`    | Sets the GOOS environment.                    |
| `-goarch`      | String | `igo -goarch amd64`  | Sets the GOARCH environment.                  |
| `-help`        | Bool   | `igo -help`          | Displays help.                                |
| `-debug`       | Bool   | `igo -debug`         | Debug output enabled.                         |
| `-verbose`     | Bool   | `igo -verbose`       | Shows Verbose Output.                         |

## Project Notes

This project is inspired from https://github.com/andreiwashere/install-go that is written in
Bash. This project is great and has been used for years, but I always held off on writing a
Go installer with Go. It felt weird, but then again... truth is stranger than fiction. 

The **install-go** package uses a convention that `igo` does not use. It places the commands
for the package in the `~/go/scripts` directory as `igo`, `sgo`, and `rgo`. This script doesn't
do that, instead the functionality from `igo` was placed inside the `install()` func, and the 
functionality from `sgo` will be moved into the `use()` func. And finally, `rgo` will be moved
into the `uninstall()` func. Currently, I've migrated over `install` and `list`. 

## TODO

- [X] Implement `-u` **Uninstall** command.
- [X] Implement `-s` **Switch** command.
- [X] Implement `-e` **Environment** Display command.
- [X] Implement `-f` **Fix Version** command.
- [X] Implement `-a` **Activate Version** command.
- [X] Add GitHub Actions Workflow
- [X] Upload compiled binaries to release
- [X] Update README with new installation instructions
- [ ] Add `igo` to `yum install igo` to `epel-release` yum repository.
- [ ] Add `igo` to `apt-get install igo` to Ubuntu repository.
- [ ] Add `igo` to `brew install igo` for macOS.

## Development Notes

Isn't it ironic that I'm using Go to write a Go installer?

This project actively runs from my original repository, 
[andreiwashere/install-go](https://github.com/andreiwashere/install-go). However, 
I use GoLand as my IDE with multiple versions of Go installed. 

Therefore, when testing this, I am using Docker. 

There is a [test.sh](test.sh) script that runs the tests from the 
[Dockerfile](Dockerfile) and runs the [tester.sh](tester.sh) file
from a fresh container. 

As development continues, I will add more tests to the [tester.sh](tester.sh) file.

Test driven development is a great way to develop software.

This script connected to the workflow [test-igo.yml](.github/workflows/test-igo.yml),
which runs automatically on the protected branches.

```bash
# === ARGUMENT PARSING ===
./params.sh # bash argument parsing helper functions

# === USAGE ===
▶ ./test.sh --help
Usage: ./test.sh [OPTIONS]
       --build      Build the Docker image (default = 'true')
       --clear      Clear console before starting (default = 'true')
       --debug      Enable debug mode (default = 'false')
       --rm         Remove the Docker image
       --verbose    Enable verbose mode (default = 'false')
```

| Usage                                   | Description                                                  |
|-----------------------------------------|--------------------------------------------------------------|
| `./test.sh`                             | Executes `docker build` and `docker run`                     |
| `./test.sh --build false`               | Executes `docker run`                                        |
| `./test.sh --rm true`                   | Executes `docker rmi` and `docker build` and `docker run`    |
| `./test.sh --verbose true`              | Executes `tester.sh` with verbose logging enabled.           |
| `./test.sh --debug true`                | Executes `tester.sh` with debug logging enabled.             |
| `./test.sh --debug true --verbose true` | Executes `tester.sh` with debug and verbose logging enabled. |

You can browse the [test_results](test_results) to see the comprehensive log outputs of the various
functionalities of the `igo` package that is tested via the `test.sh` script. The output of these 
tests are included automatically in the GitHub Actions and can be viewed there, but for archival 
purposes, the snapshot of the logs as captured with the `feature/install` branch from May 2025. 

### Running on Local

     ./run-me-local.sh --build true --rm true --clear false --debug true --verbose true

## Versioning

The workflow enforces the versioning of the `igo` binary. The `VERSION` file is updated
to the version of the binary and enforced by the workflow. 

## Branching

The `master` branch is protected and can only be merged into by a pull request
from the `release` branch. The `release` branch is protected and can only be merged into
by a pull request from the `develop` branch. The `develop` branch is protected and can
only be merged into by a pull request from the `feature/*` branch. The `feature/*` branch.
Additionally, you can use `hotfix/*` branches to fix bugs in the `master` branch.

## Why

Why build an installer of Go in Go? Because, why not? All jokes aside... 

I wanted to put myself in a development environment that I couldn't do test driven
development in natively, and come up with a way to do test driven development using
automation and DevOps. Afterall, I am a DevOps architect =D. 

If you need to use Go on a system, installed as system service, don't use `igo`; its 
made for your `$HOME` environment running as a non-privileged user. You don't require
`sudo` permissions to use `igo` or install multiple versions of Go on your system. 
