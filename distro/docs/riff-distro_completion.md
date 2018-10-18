## riff-distro completion

Generate shell completion scripts

### Synopsis

Generate shell completion scripts

```
riff-distro completion [bash|zsh] [flags]
```

### Examples

```
To install completion for bash, assuming you have `bash-completion` installed:

    riff completion bash > /etc/bash_completion.d/riff

or wherever your `bash_completion.d` is, for example `$(brew --prefix)/etc/bash_completion.d` if using homebrew.

Completion for zsh is a work in progress
```

### Options

```
  -h, --help   help for completion
```

### SEE ALSO

* [riff-distro](riff-distro.md)	 - Commands for creating a riff distribution

