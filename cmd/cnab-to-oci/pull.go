package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/docker/cnab-to-oci/remotes"
	"github.com/docker/distribution/reference"
	"github.com/docker/go/canonical/json"
	"github.com/spf13/cobra"
)

type pullOptions struct {
	output             string
	targetRef          string
	insecureRegistries []string
}

func pullCmd() *cobra.Command {
	var opts pullOptions
	cmd := &cobra.Command{
		Use:   "pull <ref> [options]",
		Short: "Pulls an image reference",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.targetRef = args[0]
			return runPull(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.output, "output", "o", "pulled.json", "output file")
	cmd.Flags().StringSliceVar(&opts.insecureRegistries, "insecure-registries", nil, "Use plain HTTP for those registries")
	return cmd
}

func runPull(opts pullOptions) error {
	ref, err := reference.ParseNormalizedNamed(opts.targetRef)
	if err != nil {
		return err
	}

	b, relocationMap, err := remotes.Pull(context.Background(), ref, createResolver(opts.insecureRegistries))
	fmt.Println("Relocation map", relocationMap)
	if err != nil {
		return err
	}
	bytes, err := json.MarshalCanonical(b)
	if err != nil {
		return err
	}
	if opts.output == "-" {
		fmt.Fprintln(os.Stdout, string(bytes))
		return nil
	}
	return ioutil.WriteFile(opts.output, bytes, 0644)
}
