/*
 * Copyright 2018 the original author or authors.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package docker

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os/exec"
	"os"
)

var _ = Describe("DockerUtils", func() {
	Describe("createDockerCommand", func() {

		var (
			cmd  *exec.Cmd
			args []string
		)

		JustBeforeEach(func() {
			cmd = createDockerCommand("subcommand", args...)
		})

		Context("when args is empty", func() {
			BeforeEach(func() {
				args = []string{}
			})

			It("should invoke the subcommand with no arguments", func() {
				Expect(cmd.Args).To(ConsistOf("docker", "subcommand"))
			})
		})

		Context("when args are provided", func() {
			BeforeEach(func() {
				args = []string{"arg1", "arg2"}
			})

			It("should invoke the subcommand with the given arguments", func() {
				Expect(cmd.Args).To(ConsistOf("docker", "subcommand", "arg1", "arg2"))
			})
		})

		Describe("Environment variable propagation", func() {
			var (
				homeWasSet    bool
				oldHome       string
				pathWasSet    bool
				oldPath       string
				dockerWasSet  bool
				oldDockerHost string
				oldDockerCert string
				oldDockerTLS  string
				oldDockerAPI  string
			)

			BeforeEach(func() {
				oldHome, homeWasSet = os.LookupEnv("HOME")
				Expect(os.Unsetenv("HOME")).To(Succeed())

				oldPath, pathWasSet = os.LookupEnv("PATH")
				Expect(os.Unsetenv("PATH")).To(Succeed())

				oldDockerHost, dockerWasSet = os.LookupEnv("DOCKER_HOST")
				oldDockerCert, _ = os.LookupEnv("DOCKER_CERT_PATH")
				oldDockerTLS, _ = os.LookupEnv("DOCKER_TLS_VERIFY")
				oldDockerAPI, _ = os.LookupEnv("DOCKER_API_VERSION")
				Expect(os.Unsetenv("DOCKER_HOST")).To(Succeed())
				Expect(os.Unsetenv("DOCKER_CERT_PATH")).To(Succeed())
				Expect(os.Unsetenv("DOCKER_TLS_VERIFY")).To(Succeed())
				Expect(os.Unsetenv("DOCKER_API_VERSION")).To(Succeed())
			})

			AfterEach(func() {
				if homeWasSet {
					Expect(os.Setenv("HOME", oldHome)).To(Succeed())
				} else {
					Expect(os.Unsetenv("HOME")).To(Succeed())
				}

				if pathWasSet {
					Expect(os.Setenv("PATH", oldPath)).To(Succeed())
				} else {
					Expect(os.Unsetenv("PATH")).To(Succeed())
				}

				if dockerWasSet {
					Expect(os.Setenv("DOCKER_HOST", oldDockerHost)).To(Succeed())
					Expect(os.Setenv("DOCKER_CERT_PATH", oldDockerCert)).To(Succeed())
					Expect(os.Setenv("DOCKER_TLS_VERIFY", oldDockerTLS)).To(Succeed())
					Expect(os.Setenv("DOCKER_API_VERSION", oldDockerAPI)).To(Succeed())
				} else {
					Expect(os.Unsetenv("DOCKER_HOST")).To(Succeed())
					Expect(os.Unsetenv("DOCKER_CERT_PATH")).To(Succeed())
					Expect(os.Unsetenv("DOCKER_TLS_VERIFY")).To(Succeed())
					Expect(os.Unsetenv("DOCKER_API_VERSION")).To(Succeed())
				}
			})

			Context("when $HOME is set", func() {
				BeforeEach(func() {
					Expect(os.Setenv("HOME", "/some/home")).To(Succeed())
				})

				It("should propagate $HOME", func() {
					Expect(cmd.Env).To(ConsistOf("HOME=/some/home"))
				})
			})

			Context("when $PATH is set", func() {
				BeforeEach(func() {
					Expect(os.Setenv("PATH", "/a:/b")).To(Succeed())
				})

				It("should propagate $PATH", func() {
					Expect(cmd.Env).To(ConsistOf("PATH=/a:/b"))
				})
			})

			Context("when $DOCKER_HOST is set", func() {
				BeforeEach(func() {
					Expect(os.Setenv("DOCKER_HOST", "tcp://192.168.99.100:1234")).To(Succeed())
				})

				It("should propagate $DOCKER_HOST", func() {
					Expect(cmd.Env).To(ConsistOf("DOCKER_HOST=tcp://192.168.99.100:1234"))
				})
			})

			Context("when $DOCKER_CERT_PATH is set", func() {
				BeforeEach(func() {
					Expect(os.Setenv("DOCKER_CERT_PATH", "/Users/riff/.minikube/certs")).To(Succeed())
				})

				It("should propagate $DOCKER_CERT_PATH", func() {
					Expect(cmd.Env).To(ConsistOf("DOCKER_CERT_PATH=/Users/riff/.minikube/certs"))
				})
			})

			Context("when $DOCKER_TLS_VERIFY is set", func() {
				BeforeEach(func() {
					Expect(os.Setenv("DOCKER_TLS_VERIFY", "1")).To(Succeed())
				})

				It("should propagate $DOCKER_TLS_VERIFY", func() {
					Expect(cmd.Env).To(ConsistOf("DOCKER_TLS_VERIFY=1"))
				})
			})

			Context("when $DOCKER_API_VERSION is set", func() {
				BeforeEach(func() {
					Expect(os.Setenv("DOCKER_API_VERSION", "1.23")).To(Succeed())
				})

				It("should propagate $DOCKER_API_VERSION", func() {
					Expect(cmd.Env).To(ConsistOf("DOCKER_API_VERSION=1.23"))
				})
			})
		})
	})
})
