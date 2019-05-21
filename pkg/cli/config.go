/*
 * Copyright 2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/projectriff/riff/pkg/k8s"
	"github.com/projectriff/riff/pkg/pack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	CompiledEnv
	ViperConfigFile string
	KubeConfigFile  string
	k8s.Client
	Exec   func(ctx context.Context, command string, args ...string) *exec.Cmd
	Pack   pack.Client
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func NewDefaultConfig() *Config {
	return &Config{
		CompiledEnv: env,
		Exec:        exec.CommandContext,
		Stdin:       os.Stdin,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
	}
}

func (c *Config) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(c.Stdout, format, a...)
}

func (c *Config) Eprintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(c.Stderr, format, a...)
}

func (c *Config) Infof(format string, a ...interface{}) (n int, err error) {
	return InfoColor.Fprintf(c.Stdout, format, a...)
}

func (c *Config) Einfof(format string, a ...interface{}) (n int, err error) {
	return InfoColor.Fprintf(c.Stderr, format, a...)
}

func (c *Config) Successf(format string, a ...interface{}) (n int, err error) {
	return SuccessColor.Fprintf(c.Stdout, format, a...)
}

func (c *Config) Esuccessf(format string, a ...interface{}) (n int, err error) {
	return SuccessColor.Fprintf(c.Stderr, format, a...)
}

func (c *Config) Errorf(format string, a ...interface{}) (n int, err error) {
	return ErrorColor.Fprintf(c.Stdout, format, a...)
}

func (c *Config) Eerrorf(format string, a ...interface{}) (n int, err error) {
	return ErrorColor.Fprintf(c.Stderr, format, a...)
}

func Initialize() *Config {
	c := NewDefaultConfig()

	cobra.OnInitialize(c.initViperConfig)
	cobra.OnInitialize(c.initKubeConfig)
	cobra.OnInitialize(c.init)

	return c
}

// initViperConfig reads in config file and ENV variables if set.
func (c *Config) initViperConfig() {
	if c.ViperConfigFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(c.ViperConfigFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			// avoid color since we don't know if it should be enabled yet
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Search config in home directory with name ".riff" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName("." + c.Name)
	}

	viper.SetEnvPrefix(c.Name)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	// hack for no-color since we urgently need to know if color should be disabled
	if viper.GetBool(StripDash(NoColorFlagName)) {
		color.NoColor = true
	}
	if err == nil {
		c.Einfof("Using config file: %s\n", viper.ConfigFileUsed())
	}
}

// initKubeConfig defines the default location for the kubectl config file
func (c *Config) initKubeConfig() {
	if c.KubeConfigFile != "" {
		return
	}
	if kubeEnvConf, ok := os.LookupEnv("KUBECONFIG"); ok {
		c.KubeConfigFile = kubeEnvConf
	} else {
		home, err := homedir.Dir()
		if err != nil {
			c.Errorf("%s\n", err)
			os.Exit(1)
		}
		c.KubeConfigFile = filepath.Join(home, ".kube", "config")
	}
}

func (c *Config) init() {
	if c.Client == nil {
		c.Client = k8s.NewClient(c.KubeConfigFile)
	}
	if c.Pack == nil {
		packClient, err := pack.NewClient(c.Stdout, c.Stderr)
		if err != nil {
			c.Eerrorf("%s\n", err)
			os.Exit(1)
		}
		c.Pack = packClient
	}
}
