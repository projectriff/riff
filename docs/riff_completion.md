---
id: riff-completion
title: "riff completion"
---
## riff completion

generate shell completion script

### Synopsis

Generate the completion script for your shell. The script is printed to stdout
and needs to be placed in the appropriate directory on your system.

```
riff completion [flags]
```

### Examples

```
riff completion
riff completion --shell zsh
```

### Options

```
  -h, --help          help for completion
      --shell shell   shell to generate completion for: bash or zsh (default "bash")
```

### Options inherited from parent commands

```
      --config file        config file (default is $HOME/.riff.yaml)
      --kube-config file   kubectl config file (default is $HOME/.kube/config)
      --no-color           disable color output in terminals
```

### SEE ALSO

* [riff](riff.md)	 - riff is for functions

