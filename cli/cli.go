// cmd/cli.go
package cli

import (
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin/v2"
)

func init() {

}

var (
	baseConfig = filepath.Join("configs", "stages.json")

	App *kingpin.Application

	// Global flags
	Verbosity *bool
	Quiet     *bool
	Debug     *bool

	// Merge command
	MergeCmd         *kingpin.CmdClause
	MergeStages      *string
	MergeInformation *string
	MergeOutput      *string

	// Run command
	RunCmd      *kingpin.CmdClause
	RunStages   *string
	RunOnly     *string
	RunWithDeps *bool
	RunWorkers  *int
	// Post command
	PostCmd      *kingpin.CmdClause
	PostStages   *string
	PostOnly     *string
	PostWithDeps *bool
	PostWorkers  *int

	// Collect command
	CollectCmd    *kingpin.CmdClause
	CollectStages *string
	CollectOutput *string

	// Report command
	ReportCmd      *kingpin.CmdClause
	ReportAnalyzed *string
	ReportBranch   *string

	// Initiate command
	InitiateCmd    *kingpin.CmdClause
	InitiateInput  *string
	InitiateOutput *string

	// Use command
	UseCmd    *kingpin.CmdClause
	UseStages *string
	UseNames  *[]string
	UseWit    *string

	// Env command
	EnvCmd      *kingpin.CmdClause
	EnvLockfile *string
)

func InitCLI() {
	App = kingpin.New(filepath.Base(os.Args[0]), "A command-line tool.")

	Verbosity = App.Flag("verbosity", "Enable verbose output").Short('v').Bool()
	Quiet = App.Flag("quiet", "Suppress all output").Short('q').Bool()
	Debug = App.Flag("debug", "Enable debug mode").Short('d').Bool()

	// Merge
	MergeCmd = App.Command("merge", "Merge the given files")
	MergeStages = MergeCmd.Flag("stages", "Path to the stage file").Short('s').Default(baseConfig).ExistingFile()
	MergeInformation = MergeCmd.Flag("information", "Path to the target file").Short('i').Required().ExistingFile()
	MergeOutput = MergeCmd.Flag("output", "Output file path").Short('o').Default("merged.json").String()

	// Run
	RunCmd = App.Command("run", "Run the merged file")
	addRunPostFlags(RunCmd, false)

	// Post
	PostCmd = App.Command("post", "Run only posts")
	addRunPostFlags(PostCmd, true)

	// Collect
	CollectCmd = App.Command("collect", "Collect Analyzed files")
	CollectStages = CollectCmd.Flag("stages", "Path to the merged file").Short('s').Default("merged.json").String()
	CollectOutput = CollectCmd.Flag("output", "Output file path").Short('o').Default("analyzed.json").String()

	// Report
	ReportCmd = App.Command("report", "Report the merged file")
	ReportAnalyzed = ReportCmd.Flag("analyzed", "Path to the analyzed file").Default("analyzed.json").String()
	ReportBranch = ReportCmd.Flag("branch", "Branch runned").Required().String()

	// Initiate
	InitiateCmd = App.Command("initiate", "Generate base initial targets.json")
	InitiateInput = InitiateCmd.Flag("input", "Input stages file").Default(baseConfig).String()
	InitiateOutput = InitiateCmd.Flag("output", "Output of the file").Default(filepath.Join("configs", "targets.json")).String()

	// Use
	UseCmd = App.Command("use", "Use the given file")
	UseStages = UseCmd.Flag("stages", "Path to the file").Default(baseConfig).ExistingFile()
	UseNames = UseCmd.Flag("name", "Name(s) to use").Short('n').Required().Strings()
	UseWit = UseCmd.Flag("wit", "Path to wit workspace").Default("wit-workspace.json").String()

	// Env
	EnvCmd = App.Command("env", "Environment variables")
	EnvLockfile = EnvCmd.Flag("lockfile", "Lockfile path").Default("workscript.lock.json").String()
}

func addRunPostFlags(cmd *kingpin.CmdClause, isPost bool) {
	stages := cmd.Flag("stages", "Path to the merged file").
		Short('s').Default("merged.json").String()
	workers := cmd.Flag("max_workers", "Number of workers").
		Short('j').Default("4").Int()
	only := cmd.Flag("only", "Run only a specific stage (format: target:stage)").
		String()
	withDeps := cmd.Flag("with-deps", "Also run dependent posts recursively").
		Bool()

	if isPost {
		PostStages, PostOnly, PostWithDeps, PostWorkers = stages, only, withDeps, workers
	} else {
		RunStages, RunOnly, RunWithDeps, RunWorkers = stages, only, withDeps, workers
	}
}
