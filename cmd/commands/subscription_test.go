package commands_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/projectriff/riff/cmd/commands"
	"github.com/projectriff/riff/pkg/core"
	"github.com/projectriff/riff/pkg/core/mocks"
	"github.com/spf13/cobra"
	"strings"
)

var _ = Describe("The riff subscription create command", func() {

	var (
		client        core.Client
		clientMock    *mocks.Client
		createCommand *cobra.Command
	)

	BeforeEach(func() {
		client = new(mocks.Client)
		clientMock = client.(*mocks.Client)
		createCommand = commands.SubscriptionCreate(&client)
	})

	AfterEach(func() {
		clientMock.AssertExpectations(GinkgoT())
	})

	It("should be documented", func() {
		Expect(createCommand.Name()).To(Equal("create"))
		Expect(createCommand.Short).NotTo(BeEmpty(), "missing short description")
		Expect(createCommand.Long).NotTo(BeEmpty(), "missing long description")
		Expect(createCommand.Example).NotTo(BeEmpty(), "missing example")
	})

	It("should define flags", func() {
		Expect(createCommand.Flag("processor")).NotTo(BeNil())
		Expect(createCommand.Flag("from")).NotTo(BeNil())
		Expect(createCommand.Flag("to")).NotTo(BeNil())
		Expect(createCommand.Flag("namespace")).NotTo(BeNil())
	})

	Context("when given wrong args or flags", func() {

		It("should fail with missing required flags", func() {
			createCommand.SetArgs([]string{})

			err := createCommand.Execute()

			Expect(err).To(MatchError(`required flag(s) "from", "processor" not set`))
		})

		It("should fail with too many args", func() {
			createCommand.SetArgs([]string{
				"too", "much", "--processor", "service", "--from", "input"})

			err := createCommand.Execute()

			Expect(err).To(MatchError(`accepts at most 1 arg(s), received 2`))
		})

		It("should fail with an invalid subscription name", func() {
			createCommand.SetArgs([]string{
				"@@invalid@@", "--processor", "service", "--from", "input"})

			err := createCommand.Execute()

			Expect(err.Error()).To(HavePrefix("a DNS-1123 subdomain must consist"))
		})
	})

	Context("when given valid args and flags", func() {
		It("should create the subscription with the provided name", func() {
			stdout := &strings.Builder{}
			createCommand.SetOutput(stdout)
			createCommand.SetArgs([]string{
				"subscription-name", "--from", "coco-chanel", "--processor", "my-service"})
			clientMock.On("CreateSubscription", core.CreateSubscriptionOptions{
				Name:       "subscription-name",
				Subscriber: "my-service",
				Channel:    "coco-chanel",
			}).Return(nil, nil)

			err := createCommand.Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal("create completed successfully\n"))
		})

		It("should create the subscription with the service name by default", func() {
			stdout := &strings.Builder{}
			createCommand.SetOutput(stdout)
			createCommand.SetArgs([]string{
				"--from", "coco-chanel", "--processor", "my-service"})
			clientMock.On("CreateSubscription", core.CreateSubscriptionOptions{
				Name:       "my-service",
				Subscriber: "my-service",
				Channel:    "coco-chanel",
			}).Return(nil, nil)

			err := createCommand.Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal("create completed successfully\n"))
		})

		It("should create the subscription with the output channel binding", func() {
			stdout := &strings.Builder{}
			createCommand.SetOutput(stdout)
			createCommand.SetArgs([]string{
				"subscription-name", "--from", "coco-chanel", "--processor", "my-service",
				"--to", "chanel-number-five"})
			clientMock.On("CreateSubscription", core.CreateSubscriptionOptions{
				Name:       "subscription-name",
				Subscriber: "my-service",
				Channel:    "coco-chanel",
				ReplyTo:    "chanel-number-five",
			}).Return(nil, nil)

			err := createCommand.Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal("create completed successfully\n"))
		})

		It("should create the subscription in the output provided namespace", func() {
			stdout := &strings.Builder{}
			createCommand.SetOutput(stdout)
			createCommand.SetArgs([]string{
				"subscription-name",
				"--from", "coco-chanel",
				"--processor", "my-service",
				"--to", "chanel-number-five",
				"--namespace", "myspace"})
			expectedOptions := core.CreateSubscriptionOptions{
				Name:       "subscription-name",
				Subscriber: "my-service",
				Channel:    "coco-chanel",
				ReplyTo:    "chanel-number-five",
			}
			expectedOptions.Namespace = "myspace"
			clientMock.On("CreateSubscription", expectedOptions).Return(nil, nil)

			err := createCommand.Execute()

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout.String()).To(Equal("create completed successfully\n"))
		})

		It("should propagate the client error", func() {
			stdout := &strings.Builder{}
			createCommand.SetOutput(stdout)
			createCommand.SetArgs([]string{
				"--from", "coco-chanel", "--processor", "my-service"})
			expectedError := errors.New("client failure")
			clientMock.On("CreateSubscription", core.CreateSubscriptionOptions{
				Name:       "my-service",
				Subscriber: "my-service",
				Channel:    "coco-chanel",
			}).Return(nil, expectedError)

			err := createCommand.Execute()

			Expect(err).To(MatchError("client failure"))
		})
	})

})
