## riff-distro create

Create a riff-distro distribution.

### Synopsis

Create a riff-distro distribution archive file (.tgz) from a given manifest.

If the output path is that of an existing directory, the file "distro.tgz" will be written in that directory. Otherwise, the file will be written at the output path.

```
riff-distro create [flags]
```

### Examples

```
  riff-distro create --output=./my-distro.tgz
```

### Options

```
  -h, --help              help for create
  -m, --manifest string   manifest for the download; can be a named manifest (stable or latest) or a path of a manifest file (default "stable")
  -o, --output string     path for the distribution archive (.tgz)
```

### SEE ALSO

* [riff-distro](riff-distro.md)	 - Commands for creating a riff distribution

