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

## Usage

| Argument   | Kind   | Usage                | Notes                         | 
|------------|--------|----------------------|-------------------------------|
| `-cmd`     | String | `igo -cmd <command>` | Run an `igo` command.         | 
| `-system`  | Bool   | `igo -system=true`   | Install system wide.          |
| `-version` | Bool   | `igo -version`       | Display `igo` binary version. |
| `-gover`   | String | `igo -gover 1.23.4`  | Installs go `1.23.4`.         |
| `-godir`   | String | `igo -godir /opt/go` | Installs `igo` in `/opt/go`.  |
| `-goos`    | String | `igo -goos linux`    | Sets the GOOS environment.    |
| `-goarch`  | String | `igo -goarch amd64`  | Sets the GOARCH environment.  |
| `-help`    | Bool   | `igo -help`          | Displays help.                |
| `-debug`   | Bool   | `igo -debug`         | Debug output enabled.         |
| `-verbose` | Bool   | `igo -verbose`       | Shows Verbose Output.         |

### Commands

| Command              | Usage                                        |
|----------------------|----------------------------------------------|
| `install` or `ins`   | Install's the `-gover` to the `-godir`.      |
| `uninstall` or `uni` | Removes the `-gover` from the `-godir`.      |
| `list` or `l`        | Lists the installed go versions in `-godir`. |
| `use` or `u`         | Activate a go version in `-godir`.           |
| `fix` or `f`         | Fixes a go version in `-godir`.              |

## Real World Example

```bash
igo -cmd list # long form
igo -cmd l # short form

igo -cmd install -gover 1.24.2 # long form
igo -cmd ins -gover 1.24.2 # short form

igo -cmd uninstall -gover 1.24.2 # long form
igo -cmd uni -gover 1.24.2 # short form

igo -cmd use -gover 1.24.2 # long form
igo -cmd u -gover 1.24.2 # short form

igo -cmd fix -gover 1.24.2 # long form
igo -cmd f -gover 1.24.2 # short form
```

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

- [ ] Implement `-cmd homebrew` to migrate a `brew install go` to `igo` compatible syntax
- [ ] Implement a check that verifies if go is installed with `~/go/bin` and `~/go/pkg` to integrate support
- [ ] Implement `-cmd env` to debug environment
- [ ] Implement `-cmd uninstall`
- [ ] Implement `-cmd fix`
- [ ] Implement `-cmd use`
- [ ] Add Unit Testing
- [ ] Add GitHub Actions Workflow
- [ ] Upload compiled binaries to release
- [ ] Update README with new installation instructions
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
       --debug       (default = 'false')
       --rm         Remove the Docker image
       --verbose     (default = 'false')
```

| Usage                                   | Description                                                  |
|-----------------------------------------|--------------------------------------------------------------|
| `./test.sh`                             | Executes `docker build` and `docker run`                     |
| `./test.sh --build false`               | Executes `docker run`                                        |
| `./test.sh --rm true`                   | Executes `docker rmi` and `docker build` and `docker run`    |
| `./test.sh --verbose true`              | Executes `tester.sh` with verbose logging enabled.           |
| `./test.sh --debug true`                | Executes `tester.sh` with debug logging enabled.             |
| `./test.sh --debug true --verbose true` | Executes `tester.sh` with debug and verbose logging enabled. |


## Versioning

The workflow enforces the versioning of the `igo` binary. The `VERSION` file is updated
to the version of the binary and enforced by the workflow. 

## Branching

The `master` branch is protected and can only be merged into by a pull request
from the `release` branch. The `release` branch is protected and can only be merged into
by a pull request from the `develop` branch. The `develop` branch is protected and can
only be merged into by a pull request from the `feature/*` branch. The `feature/*` branch.
Additionally, you can use `hotfix/*` branches to fix bugs in the `master` branch.
