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
	"github.com/projectriff/riff/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Params struct {
	ConfigFile     string
	KubeConfigFile string
	Client         *client.Client
}

func (p *Params) Initialize() {
	if p.Client == nil {
		c := client.NewClient(p.KubeConfigFile)
		p.Client = c
	}
}

func Initialize() *Params {
	p := &Params{}

	cobra.OnInitialize(p.initConfig)
	cobra.OnInitialize(p.initKubeConfig)
	cobra.OnInitialize(p.Initialize)

	return p
}

// initConfig reads in config file and ENV variables if set.
func (p *Params) initConfig() {
	if p.ConfigFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(p.ConfigFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Search config in home directory with name ".riff" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".riff")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// initKubeConfig defines the default location for the kubectl config file
func (p *Params) initKubeConfig() {
	if p.KubeConfigFile != "" {
		return
	}
	if kubeEnvConf, ok := os.LookupEnv("KUBECONFIG"); ok {
		p.KubeConfigFile = kubeEnvConf
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		p.KubeConfigFile = filepath.Join(home, ".kube", "config")
	}
}
