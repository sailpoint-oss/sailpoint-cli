// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package jsonpath

import (
	"fmt"
	"os"

	"github.com/bhmj/jsonslice"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
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

			// Format the JSON
			formattedJSON := pretty.Pretty([]byte(result))

			// Color the JSON
			coloredJSON := pretty.Color(formattedJSON, nil)

			// Print formatted and colored JSON
			fmt.Print(string(coloredJSON))

			return nil

		},
	}

	cmd.Flags().StringVarP(&filepath, "file", "f", "", "The path to the json you wish to evaluate")
	cmd.Flags().StringVarP(&path, "path", "p", "", "The json path to evaluate against the file provided")

	return cmd
}
