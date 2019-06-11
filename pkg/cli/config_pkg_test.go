/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

func TestInitViperConfig(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = noColor }()

	defer viper.Reset()

	c := NewDefaultConfig()
	output := &bytes.Buffer{}
	c.Stdout = output
	c.Stderr = output

	c.ViperConfigFile = "testdata/.riff.yaml"
	c.initViperConfig()

	expectedViperSettings := map[string]interface{}{
		"no-color": true,
	}
	if diff := cmp.Diff(expectedViperSettings, viper.AllSettings()); diff != "" {
		t.Errorf("Unexpected viper settings (-expected, +actual): %s", diff)
	}
	if diff := cmp.Diff("Using config file: testdata/.riff.yaml", strings.TrimSpace(output.String())); diff != "" {
		t.Errorf("Unexpected output (-expected, +actual): %s", diff)
	}
}

func TestInitViperConfig_HomeDir(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = noColor }()

	defer viper.Reset()

	home, homeisset := os.LookupEnv("HOME")
	defer func() {
		homedir.Reset()
		if homeisset {
			os.Setenv("HOME", home)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	c := NewDefaultConfig()
	output := &bytes.Buffer{}
	c.Stdout = output
	c.Stderr = output

	os.Setenv("HOME", "testdata")
	c.initViperConfig()

	expectedViperSettings := map[string]interface{}{
		"no-color": true,
	}
	if diff := cmp.Diff(expectedViperSettings, viper.AllSettings()); diff != "" {
		t.Errorf("Unexpected viper settings (-expected, +actual): %s", diff)
	}
}

func TestInitKubeConfig_Flag(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = noColor }()

	c := NewDefaultConfig()
	output := &bytes.Buffer{}
	c.Stdout = output
	c.Stderr = output

	c.KubeConfigFile = "testdata/.kube/config"
	c.initKubeConfig()

	if expected, actual := "testdata/.kube/config", c.KubeConfigFile; expected != actual {
		t.Errorf("Expected kubeconfig path %q, actually %q", expected, actual)
	}
	if diff := cmp.Diff("", strings.TrimSpace(output.String())); diff != "" {
		t.Errorf("Unexpected output (-expected, +actual): %s", diff)
	}
}

func TestInitKubeConfig_EnvVar(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = noColor }()

	kubeconfig, kubeconfigisset := os.LookupEnv("KUBECONFIG")
	defer func() {
		if kubeconfigisset {
			os.Setenv("KUBECONFIG", kubeconfig)
		} else {
			os.Unsetenv("KUBECONFIG")
		}
	}()

	c := NewDefaultConfig()
	output := &bytes.Buffer{}
	c.Stdout = output
	c.Stderr = output

	os.Setenv("KUBECONFIG", "testdata/.kube/config")
	c.initKubeConfig()

	if expected, actual := "testdata/.kube/config", c.KubeConfigFile; expected != actual {
		t.Errorf("Expected kubeconfig path %q, actually %q", expected, actual)
	}
	if diff := cmp.Diff("", strings.TrimSpace(output.String())); diff != "" {
		t.Errorf("Unexpected output (-expected, +actual): %s", diff)
	}
}

func TestInitKubeConfig_HomeDir(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = noColor }()

	home, homeisset := os.LookupEnv("HOME")
	defer func() {
		homedir.Reset()
		if homeisset {
			os.Setenv("HOME", home)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	c := NewDefaultConfig()
	output := &bytes.Buffer{}
	c.Stdout = output
	c.Stderr = output

	os.Setenv("HOME", "testdata")
	c.initKubeConfig()

	if expected, actual := filepath.FromSlash("testdata/.kube/config"), c.KubeConfigFile; expected != actual {
		t.Errorf("Expected kubeconfig path %q, actually %q", expected, actual)
	}
	if diff := cmp.Diff("", strings.TrimSpace(output.String())); diff != "" {
		t.Errorf("Unexpected output (-expected, +actual): %s", diff)
	}
}

func TestInit(t *testing.T) {
	noColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = noColor }()

	c := NewDefaultConfig()
	output := &bytes.Buffer{}
	c.Stdout = output
	c.Stderr = output

	c.KubeConfigFile = "testdata/.kube/config"
	c.init()

	if expected, actual := "default", c.DefaultNamespace(); expected != actual {
		t.Errorf("Expected default namespace %q, actually %q", expected, actual)
	}
	if diff := cmp.Diff("", strings.TrimSpace(output.String())); diff != "" {
		t.Errorf("Unexpected output (-expected, +actual): %s", diff)
	}
	if c.Client == nil {
		t.Errorf("Expected c.Client tp be set, actually %v", c.Client)
	}
	if c.Pack == nil {
		t.Errorf("Expected c.Pack tp be set, actually %v", c.Pack)
	}
	if c.Kail == nil {
		t.Errorf("Expected c.Kail tp be set, actually %v", c.Kail)
	}
}
