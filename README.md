# rcfix

deprecates broken rc.d files after they fall from grace by converting systemd unit definitions to system 5 services (sysVinit init scripts)

fix your broken systems today, _uninstall systemd._

## Usage

you have two ways to save these sad, sad, broken services.

#### method 1: stdio

`cat poor.service | ./rcfix > perfectangle.sh`

#### method 2: files

`./rcfix poor.service perfectangle.sh`

---

# [Automatic Builds](https://github.com/yunginnanet/rcfix/releases)
