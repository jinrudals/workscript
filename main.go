package main

import (
	"fmt"
	"os"

	cli "github.com/jinrudals/workscript/cli"
	cmd "github.com/jinrudals/workscript/command"
)

func main() {
	cli.InitCLI() // Initialize CLI

	parsedCommand, err := cli.App.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing command: %v\n", err)
		os.Exit(1)
	}

	// Example: simple dispatch
	switch parsedCommand {
	case cli.MergeCmd.FullCommand():
		fmt.Println("Running Merge:", *cli.MergeStages, *cli.MergeInformation)
		merge := cmd.MergeCommand{
			StagesPath:      *cli.MergeStages,
			InformationPath: *cli.MergeInformation,
			OutputPath:      *cli.MergeOutput,
		}
		if err := merge.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running merge: %v\n", err)
			os.Exit(1)
		}
	case cli.RunCmd.FullCommand():
		err := cmd.Run(*cli.RunStages, *cli.RunOnly, *cli.RunWithDeps, *cli.RunWorkers, "all")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Run failed: %v\n", err)
			os.Exit(1)
		}
	case cli.PostCmd.FullCommand():
		fmt.Println("Running Post")
		err := cmd.Post(*cli.PostStages, *cli.PostOnly, *cli.PostWithDeps, *cli.PostWorkers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Post failed: %v\n", err)
			os.Exit(1)
		}

	case cli.ReportCmd.FullCommand():
		fmt.Println("Running Report")
		// and so on...
	}
}
