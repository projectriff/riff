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

package riff

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/projectriff/riff/pkg/env"
	"github.com/projectriff/riff/pkg/fs"
	"github.com/projectriff/riff/pkg/k8s"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	ViperConfigFile string
	KubeConfigFile  string
	k8s.Client
	FileSystem fs.FileSystem
}

func Initialize() *Config {
	c := &Config{}

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
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Search config in home directory with name ".riff" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName("." + env.Cli.Name)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
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
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		c.KubeConfigFile = filepath.Join(home, ".kube", "config")
	}
}

func (c *Config) init() {
	if c.FileSystem == nil {
		c.FileSystem = &fs.Local{}
	}
	if c.Client == nil {
		c.Client = k8s.NewClient(c.KubeConfigFile)
	}
}
