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

