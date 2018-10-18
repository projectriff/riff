## riff-distro system download

Download a riff system.

### Synopsis

Download the kubernetes configuration files for a given riff manifest.

Use the `--output` flag to specify the path of a directory to contain the resultant kubernetes configuration files and rewritten riff manifest.The riff manifest is rewritten to refer to the downloaded configuration files.


```
riff-distro system download [flags]
```

### Examples

```
  riff system download --manifest=path/to/manifest.yaml --output=path/to/output/dir
```

### Options

```
  -h, --help              help for download
  -m, --manifest string   manifest for the download; can be a named manifest (stable or latest) or a path of a manifest file (default "stable")
  -o, --output string     path to contain the output file(s)
```

### SEE ALSO

* [riff-distro system](riff-distro_system.md)	 - Interact with riff systems

