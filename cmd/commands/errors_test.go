package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/projectriff/riff/cmd/commands"
)

var _ = Describe("The result handling utility", func() {
	It("returns a single error message as is", func() {
		input := []error{errors.New("nope")}

		result := commands.MergeResults(input)

		Expect(result).To(MatchError("nope"))
	})

	It("returns a single non-nil error message as is", func() {
		input := []error{nil, errors.New("nope"), nil}

		result := commands.MergeResults(input)

		Expect(result).To(MatchError("nope"))
	})

	It("merges error messages", func() {
		input := []error{errors.New("nope"),
			errors.New("still nope")}

		result := commands.MergeResults(input)

		Expect(result).To(MatchError(`# Call 1 - error
	nope
# Call 2 - error
	still nope
`))
	})

	It("merges error messages, including nil ones", func() {
		input := []error{errors.New("nope"),
			nil,
			errors.New("still nope")}

		result := commands.MergeResults(input)

		Expect(result).To(MatchError(`# Call 1 - error
	nope
# Call 2 - no error
# Call 3 - error
	still nope
`))
	})

	It("indents multi-line messages", func() {
		input := []error{errors.New("nope\nnope\nnope"), errors.New("ah")}

		result := commands.MergeResults(input)

		Expect(result).To(MatchError(`# Call 1 - error
	nope
	nope
	nope
# Call 2 - error
	ah
`))
	})


	It("returns nil when there is no error", func() {
		var input []error

		result := commands.MergeResults(input)

		Expect(result).To(BeNil())
	})

	It("returns nil when there are only nil errors", func() {
		input := []error{nil, nil}

		result := commands.MergeResults(input)

		Expect(result).To(BeNil())
	})
})
