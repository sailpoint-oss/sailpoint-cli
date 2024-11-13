// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package jsonpath

import (
	"fmt"
	"os"

	"github.com/bhmj/jsonslice"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

func newEvalCommand() *cobra.Command {
	var filepath string
	var path string

	cmd := &cobra.Command{
		Use:     "eval",
		Short:   "Evaluate a jsonpath against a json file",
		Long:    "\nEvaluate a jsonpath against a json file\n\n",
		Example: "sail jsonpath eval | sail jsonpath e",
		Aliases: []string{"e"},
		Args:    cobra.OnlyValidArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			var data []byte
			var err error

			if filepath != "" {
				data, err = os.ReadFile(filepath)
				if err != nil {
					return err
				}
			} else {
				log.Error("You must provide a file to preview")
			}

			result, err := jsonslice.Get(data, path)

			if err != nil {
				return err
			}

			fmt.Print(string(result))

			return nil

		},
	}

	cmd.Flags().StringVarP(&filepath, "file", "f", "", "The path to the json you wish to evaluate")
	cmd.Flags().StringVarP(&path, "path", "p", "", "The json path to evaluate against the file provided")

	return cmd
}
