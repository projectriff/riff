## riff init java

Initialize a Java function

### Synopsis


Generate the function based on the function source code specified as the filename, using the artifact (jar file),
  the function handler(classname), the name and version specified for the function image repository and tag. 

For example, from a maven project directory named 'greeter', type:

riff init java -i greetings -l java -a target/greeter-1.0.0.jar --handler=Greeter

to generate the required Dockerfile and resource definitions using sensible defaults.

```
riff init java [flags]
```

### Options

```
      --handler string   the fully qualified class name of the function handler
  -h, --help             help for java
```

### Options inherited from parent commands

```
  -a, --artifact string       path to the function artifact, source code or jar file
      --config string         config file (default is $HOME/.riff.yaml)
      --dry-run               print generated function artifacts content to stdout only
  -f, --filepath string       path or directory used for the function resources (defaults to the current directory)
      --force                 overwrite existing functions artifacts
  -i, --input string          the name of the input topic (defaults to function name)
  -n, --name string           the name of the function (defaults to the name of the current directory)
  -o, --output string         the name of the output topic (optional)
  -p, --protocol string       the protocol to use for function invocations
      --riff-version string   the version of riff to use when building containers (default "latest")
  -u, --useraccount string    the Docker user account to be used for the image repository (default "trisberg")
  -v, --version string        the version of the function image (default "0.0.1")
```

### SEE ALSO
* [riff init](riff_init.md)	 - Initialize a function

