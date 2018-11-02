/*
 * Copyright 2018 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"io"
	"io/ioutil"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("pollService", func() {
	const sleepsBeforeTimeout = 50

	var (
		errChan       = make(chan error, 1)
		timeout       = sleepsBeforeTimeout * time.Microsecond
		sleepDuration = time.Microsecond
		log           io.Writer
		err           error

		hardErr                         = errors.New("hard error")
		transientErr                    = errors.New("transient error")
		transientErrCount, hardErrCount int
		transientErrors, hardErrors     int

		check = func() (error, error) {
			transientErrCount++
			if transientErrCount <= transientErrors {
				return transientErr, nil
			}
			hardErrCount++
			if hardErrCount <= hardErrors {
				return nil, hardErr
			}
			return nil, nil
		}
	)

	BeforeEach(func() {
		log = ioutil.Discard
		transientErrCount = 0
		hardErrCount = 0
		transientErrors = 0
		hardErrors = 0
	})

	JustBeforeEach(func() {
		err = pollService(check, errChan, timeout, sleepDuration, log)
	})

	Context("when an error is sent to the error channel", func() {
		BeforeEach(func() {
			errChan <- errors.New("channel error")
		})

		It("should stop polling and return the error", func() {
			Expect(err).To(MatchError("channel error"))
		})
	})

	Context("when a hard error occurs", func() {
		BeforeEach(func() {
			hardErrors = 1
		})

		It("should stop polling and return the error", func() {
			Expect(err).To(MatchError(hardErr))
		})
	})

	Context("when a transient error occurs temporarily", func() {
		BeforeEach(func() {
			transientErrors = 1
		})

		It("should stop polling and return successfully", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when a transient error persists beyond the timeout", func() {
		BeforeEach(func() {
			transientErrors = sleepsBeforeTimeout + 1
		})

		It("should time out and return the transient error", func() {
			Expect(err).To(MatchError(transientErr))
		})
	})

	Context("when a transient error is followed by a hard error", func() {
		BeforeEach(func() {
			transientErrors = sleepsBeforeTimeout
			hardErrors = 1
		})

		It("should stop polling and return successfully", func() {
			Expect(err).To(MatchError(hardErr))
		})
	})
})
