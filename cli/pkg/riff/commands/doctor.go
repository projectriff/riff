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

package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/projectriff/riff/cli/pkg/cli"
	"github.com/projectriff/riff/cli/pkg/cli/printers"
	"github.com/spf13/cobra"
	authv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const riffSystemNamespace = "riff-system"

type DoctorOptions struct {
	Namespace string
}

var (
	_ cli.Validatable = (*DoctorOptions)(nil)
	_ cli.Executable  = (*DoctorOptions)(nil)
)

func (opts *DoctorOptions) Validate(ctx context.Context) cli.FieldErrors {
	errs := cli.FieldErrors{}

	if opts.Namespace == "" {
		errs = errs.Also(cli.ErrMissingField(cli.NamespaceFlagName))
	}

	return errs
}

func (opts *DoctorOptions) Exec(ctx context.Context, c *cli.Config) error {
	riffNamespaces := []string{
		opts.Namespace,
		riffSystemNamespace,
	}
	err := opts.checkNamespaces(c, riffNamespaces)
	if err != nil {
		return err
	}

	verbs := []string{"get", "list", "create", "update", "delete", "patch", "watch"}
	readVerbs := []string{"get", "list", "watch"}
	accessChecks := doctorAccessChecks{
		{Attributes: &authv1.ResourceAttributes{Namespace: riffSystemNamespace, Group: "core", Resource: "configmaps", Name: "builders"}, Verbs: readVerbs},
		{Attributes: &authv1.ResourceAttributes{Namespace: opts.Namespace, Group: "core", Resource: "configmaps"}, Verbs: verbs},
		{Attributes: &authv1.ResourceAttributes{Namespace: opts.Namespace, Group: "core", Resource: "secrets"}, Verbs: verbs},
		{Attributes: &authv1.ResourceAttributes{Namespace: opts.Namespace, Group: "core", Resource: "pods"}, Verbs: readVerbs},
		{Attributes: &authv1.ResourceAttributes{Namespace: opts.Namespace, Group: "core", Resource: "pods", Subresource: "log"}, Verbs: readVerbs},
		{Attributes: &authv1.ResourceAttributes{Namespace: opts.Namespace, Group: "build.projectriff.io", Resource: "applications"}, Verbs: verbs},
		{Attributes: &authv1.ResourceAttributes{Namespace: opts.Namespace, Group: "build.projectriff.io", Resource: "containers"}, Verbs: verbs},
		{Attributes: &authv1.ResourceAttributes{Namespace: opts.Namespace, Group: "build.projectriff.io", Resource: "functions"}, Verbs: verbs},
	}
	if c.Runtimes[cli.CoreRuntime] {
		accessChecks = append(accessChecks,
			&doctorAccessCheck{Attributes: &authv1.ResourceAttributes{Namespace: opts.Namespace, Group: "core.projectriff.io", Resource: "deployers"}, Verbs: verbs},
		)
	}
	if c.Runtimes[cli.StreamingRuntime] {
		accessChecks = append(accessChecks,
			&doctorAccessCheck{Attributes: &authv1.ResourceAttributes{Namespace: opts.Namespace, Group: "streaming.projectriff.io", Resource: "processors"}, Verbs: verbs},
			&doctorAccessCheck{Attributes: &authv1.ResourceAttributes{Namespace: opts.Namespace, Group: "streaming.projectriff.io", Resource: "streams"}, Verbs: verbs},
			&doctorAccessCheck{Attributes: &authv1.ResourceAttributes{Namespace: opts.Namespace, Group: "streaming.projectriff.io", Resource: "inmemorygateways"}, Verbs: verbs},
			&doctorAccessCheck{Attributes: &authv1.ResourceAttributes{Namespace: opts.Namespace, Group: "streaming.projectriff.io", Resource: "kafkagateways"}, Verbs: verbs},
			&doctorAccessCheck{Attributes: &authv1.ResourceAttributes{Namespace: opts.Namespace, Group: "streaming.projectriff.io", Resource: "pulsargateways"}, Verbs: verbs},
		)
	}
	if c.Runtimes[cli.KnativeRuntime] {
		accessChecks = append(accessChecks,
			&doctorAccessCheck{Attributes: &authv1.ResourceAttributes{Namespace: opts.Namespace, Group: "knative.projectriff.io", Resource: "adapters"}, Verbs: verbs},
			&doctorAccessCheck{Attributes: &authv1.ResourceAttributes{Namespace: opts.Namespace, Group: "knative.projectriff.io", Resource: "deployers"}, Verbs: verbs},
		)
	}

	err = opts.checkAccess(c, accessChecks)
	if err != nil {
		return err
	}

	return nil
}

func NewDoctorCommand(ctx context.Context, c *cli.Config) *cobra.Command {
	opts := &DoctorOptions{}

	cmd := &cobra.Command{
		Use:     "doctor",
		Aliases: []string{"doc"},
		Short:   "check " + c.Name + "'s permissions",
		Long: strings.TrimSpace(`
The doctor checks that the current user has permission to access ` + c.Name + `, and ` + c.Name + `
related, resources in a namespace.

The doctor is not a tool for monitoring the health of the cluster or the ` + c.Name + `
install.
`),
		Example: "riff doctor",
		PreRunE: cli.ValidateOptions(ctx, opts),
		RunE:    cli.ExecOptions(ctx, c, opts),
	}

	cli.NamespaceFlag(cmd, c, &opts.Namespace)

	return cmd
}

func (*DoctorOptions) checkNamespaces(c *cli.Config, requiredNamespaces []string) error {
	printer := printers.GetNewTabWriter(c.Stdout)
	defer printer.Flush()
	fmt.Fprintf(printer, "NAMESPACE\tSTATUS\n")
	for _, namespace := range requiredNamespaces {
		status := cli.Ssuccessf("ok")
		_, err := c.Core().Namespaces().Get(namespace, metav1.GetOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				return err
			}
			status = cli.Serrorf("missing")
		}
		fmt.Fprintf(printer, "%s\t%s\n", namespace, status)
	}
	return nil
}

func (*DoctorOptions) checkAccess(c *cli.Config, accessChecks doctorAccessChecks) error {
	err := accessChecks.ResolveStatus(c)
	if err != nil {
		return err
	}
	c.Printf("\n")
	printer := printers.GetNewTabWriter(c.Stdout)
	defer printer.Flush()
	fmt.Fprintf(printer, "RESOURCE\tNAMESPACE\tNAME\tREAD\tWRITE\n")
	for _, check := range accessChecks {
		resource := check.Attributes.Resource
		if check.Attributes.Group != "core" {
			resource = fmt.Sprintf("%s.%s", resource, check.Attributes.Group)
		}
		if check.Attributes.Subresource != "" {
			resource = fmt.Sprintf("%s/%s", resource, check.Attributes.Subresource)
		}
		name := check.Attributes.Name
		if name == "" {
			name = "*"
		}
		fmt.Fprintf(printer, "%s\t%s\t%s\t%s\t%s\n", resource, check.Attributes.Namespace, name, check.ReadStatus.String(), check.WriteStatus.String())
	}
	return nil
}

type doctorAccessCheck struct {
	Attributes  *authv1.ResourceAttributes
	Verbs       []string
	ReadStatus  doctorAccessStatus
	WriteStatus doctorAccessStatus
}

func (check *doctorAccessCheck) ResolveStatus(c *cli.Config) error {
	if strings.Contains(check.Attributes.Group, ".") {
		missing, err := check.isCustomResourceMissing(c, fmt.Sprintf("%s.%s", check.Attributes.Resource, check.Attributes.Group))
		if err != nil {
			return err
		}
		if missing {
			check.ReadStatus = doctorAccessMissing
			check.WriteStatus = doctorAccessMissing
			return nil
		}
	}
	for _, verb := range check.Verbs {
		attributes := check.Attributes.DeepCopy()
		attributes.Verb = verb
		review, err := c.Auth().SelfSubjectAccessReviews().Create(&authv1.SelfSubjectAccessReview{
			Spec: authv1.SelfSubjectAccessReviewSpec{
				ResourceAttributes: attributes,
			},
		})
		if err != nil {
			return err
		}
		if review.Status.EvaluationError != "" {
			return fmt.Errorf(review.Status.EvaluationError)
		}
		status := doctorAccessUndefined
		if review.Status.Allowed {
			status = doctorAccessAllowed
		} else if review.Status.Denied {
			status = doctorAccessDenied
		} else {
			status = doctorAccessUnknown
		}
		if verb == "get" || verb == "list" || verb == "watch" {
			check.ReadStatus = check.ReadStatus.Combine(status)
		} else {
			check.WriteStatus = check.WriteStatus.Combine(status)
		}
	}
	return nil
}

func (check *doctorAccessCheck) isCustomResourceMissing(c *cli.Config, name string) (bool, error) {
	_, err := c.APIExtension().CustomResourceDefinitions().Get(name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return true, nil
	}
	return false, err
}

type doctorAccessChecks []*doctorAccessCheck

func (checks doctorAccessChecks) ResolveStatus(c *cli.Config) error {
	for _, check := range checks {
		if err := check.ResolveStatus(c); err != nil {
			return err
		}
	}
	return nil
}

func (checks doctorAccessChecks) IsHealthy() bool {
	for _, check := range checks {
		if check.ReadStatus != doctorAccessAllowed && check.ReadStatus != doctorAccessUndefined {
			return false
		}
		if check.WriteStatus != doctorAccessAllowed && check.WriteStatus != doctorAccessUndefined {
			return false
		}
	}
	return true
}

type doctorAccessStatus int

const (
	doctorAccessUndefined doctorAccessStatus = iota
	doctorAccessAllowed                      /* right is granted */
	doctorAccessDenied                       /* right is denied */
	doctorAccessMixed                        /* for the same resource, some rights are granted, some are denied */
	doctorAccessMissing                      /* resource not deployed */
	doctorAccessUnknown                      /* ambiguous review */
)

func (das doctorAccessStatus) Combine(new doctorAccessStatus) doctorAccessStatus {
	if das == doctorAccessUndefined {
		return new
	}
	if das == doctorAccessUnknown || new == doctorAccessUnknown {
		return doctorAccessUnknown
	}
	if das != new {
		return doctorAccessMixed
	}
	if das == doctorAccessAllowed {
		return doctorAccessAllowed
	}
	return doctorAccessDenied
}

func (das doctorAccessStatus) String() string {
	switch das {
	case doctorAccessAllowed:
		return cli.Ssuccessf("allowed")
	case doctorAccessMixed:
		return cli.Swarnf("mixed")
	case doctorAccessDenied:
		return cli.Swarnf("denied")
	case doctorAccessMissing:
		return cli.Serrorf("missing")
	case doctorAccessUnknown:
		return cli.Serrorf("unknown")
	}
	return "n/a"
}
