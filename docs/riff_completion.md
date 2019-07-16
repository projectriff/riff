---
id: riff-completion
title: "riff completion"
---
## riff completion

Generate shell completion scripts

### Synopsis

Generate shell completion scripts

```
riff completion [bash|zsh] [flags]
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

* [riff](riff.md)	 - Commands for creating and managing function resources

