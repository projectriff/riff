package function

import (
	"path/filepath"
	"fmt"
	"errors"

	"github.com/dturanski/riff-cli/pkg/osutils"
)

type InitOptions struct {
	UserAccount  string
	FunctionName string
	Version      string
	FunctionPath string
	Language     string
	Protocol     string
	Input        string
	Output       string
	Artifact     string
	Classname    string
	RiffVersion  string
	Push         bool
}


type Initializer struct {
	options      InitOptions
	functionFile string
}

func NewInitializer() *Initializer {
	return &Initializer{}
}


func (this Initializer) Initialize(opts InitOptions) error {



	err := this.deriveOptionsFromFunctionPath(opts)
	if err != nil {
		return err
	}

	err = this.resolveArtifact(opts.Artifact)
	if err != nil {
		return err
	}

	err = this.resolveProtocol(opts.Protocol)
	if err != nil {
		return err
	}

	if this.options.Language == "java" {
		if opts.Classname == "" {
			return errors.New("'classname is required for java")
		}
	}


	if opts.Input == "" {
		this.options.Input = this.options.FunctionName
	}

	this.options.Output = opts.Output
	this.options.UserAccount = opts.UserAccount
	this.options.Push = opts.Push
	this.options.RiffVersion = opts.RiffVersion
	this.options.Version = opts.Version

	fmt.Printf("function file: %s\noptions: %+v\n",this.functionFile, this.options)

	return nil
}

func (this *Initializer) deriveOptionsFromFunctionPath(opts InitOptions) error {
	var fileExtenstions = map[string]string {
		"shell"		:  	"sh",
		"java"		:   "java",
		"node"		:   "js",
		"js"		:   "js",
		"python"	: 	"py",
	}



	if !osutils.FileExists(opts.FunctionPath) {
		return errors.New(fmt.Sprintf("File does not exist %s", opts.FunctionPath))
	}

	this.options.FunctionPath, _ = filepath.Abs(opts.FunctionPath)

	if osutils.IsDirectory(this.options.FunctionPath) {
		if opts.FunctionName == "" {
			this.options.FunctionName = filepath.Base(this.options.FunctionPath)
		} else {
			this.options.FunctionName = opts.FunctionName
		}

		if opts.Language == "" {
			for lang, ext := range fileExtenstions {
				fileName := fmt.Sprintf("%s.%s",this.options.FunctionName, ext)
				functionFile := filepath.Join(this.options.FunctionPath, fileName)
				if osutils.FileExists(functionFile) {
					this.options.Language = lang
					this.functionFile = functionFile
					break
				}
			}
			if this.options.Language == "" {
				return errors.New(fmt.Sprintf("cannot find function source for function %s in directory %s", this.options.FunctionName, this.options.FunctionPath))
			}
		} else {
			ext := fileExtenstions[opts.Language]
			if ext == "" {
				return errors.New(fmt.Sprintf("language %s is unsupported", opts.Language))
			}
			this.options.Language = opts.Language

			fileName := fmt.Sprintf("%s.%s",this.options.FunctionName, ext)
			this.functionFile = filepath.Join(this.options.FunctionPath, fileName)
			if !osutils.FileExists(this.functionFile) {
				return errors.New(fmt.Sprintf("cannot find function source for function %s", this.functionFile))
			}
		}
	} else {
		//regular file given
		ext := filepath.Ext(this.options.FunctionPath)
		if opts.Language == "" {
			for lang, e := range fileExtenstions {
				if e == ext {
					this.options.Language = lang
					break
				}
			}
			if this.options.Language == "" {
				return errors.New(fmt.Sprintf("cannot find function source for function %s in directory %s", this.options.FunctionName, this.options.FunctionPath))
			}
		} else {
			this.options.Language = opts.Language
			if fileExtenstions[this.options.Language] != ext {
				fmt.Printf("WARNING non standard extension %s given for language %s. We'll see what we can do", ext, this.options.Language)
			}
		}
	}


	return nil
}

func (this *Initializer) resolveProtocol(protocol string) error {
	var defaultProtocols = map[string]string {
		"shell"		:  	"stdio",
		"java"		:   "http",
		"node"		:   "http",
		"js"		:   "http",
		"python"	: 	"stdio",
	}

	var supportedProtocols = []string{"stdio","http","grpc"}

	if protocol == "" {
		this.options.Protocol = defaultProtocols[this.options.Language]

	} else {
		supported := false
		for _, p := range supportedProtocols {
			if protocol == p {
				supported =  true
			}
		}
		if (!supported) {
			return errors.New(fmt.Sprintf("protocol %s is unsupported \n", protocol))
		}
		this.options.Protocol = protocol
	}
	return nil
}

func (this *Initializer) resolveArtifact(artifact string) error {
	if artifact == "" {
		////TODO: Needs work...
		this.options.Artifact = filepath.Base(this.functionFile)
		return nil
	}

	//TODO: What if the artifact ext doesn't match the language?
	if !osutils.FileExists(artifact) {
		return errors.New(fmt.Sprintf("Artifact does not exist %s", artifact))
	}
	return nil
}

