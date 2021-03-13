package view

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/internal/ghrepo"
	"github.com/cli/cli/pkg/cmdutil"
	"github.com/cli/cli/pkg/iostreams"
	"github.com/spf13/cobra"
)

type ViewOptions struct {
	HttpClient func() (*http.Client, error)
	IO         *iostreams.IOStreams
	BaseRepo   func() (ghrepo.Interface, error)

	WorkflowID string
	YAML       string

	Prompt bool
}

func NewCmdView(f *cmdutil.Factory, runF func(*ViewOptions) error) *cobra.Command {
	opts := &ViewOptions{
		IO:         f.IOStreams,
		HttpClient: f.HttpClient,
	}
	cmd := &cobra.Command{
		Use:    "view [<run-id>]",
		Short:  "View a summary of a workflow run",
		Args:   cobra.MaximumNArgs(1),
		Hidden: true,
		Example: heredoc.Doc(`
		  # Interactively select a workflow to view
		  $ gh workflow view

		  # View a specific workflow
		  $ gh workflow view 0451
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			// support `-R, --repo` override
			opts.BaseRepo = f.BaseRepo

			if len(args) > 0 {
				opts.WorkflowID = args[0]
			} else if !opts.IO.CanPrompt() {
				return &cmdutil.FlagError{Err: errors.New("workflow ID required when not running interactively")}
			} else {
				opts.Prompt = true
			}

			if runF != nil {
				return runF(opts)
			}
			return runView(opts)
		},
	}

	return cmd
}

func runView(opts *ViewOptions) error {
	// TODO

	mock := heredoc.Doc(`
		%s %s
		%s

		%s
		%s add even more tests        actions       pull_request 1m ago %s
		%s start on workflow comma... workflow-view push         2h ago %s
		%s add stubs                  workflow-list push         2d ago %s

		To see more runs for this workflow, try: gh run list -w"goreleaser"
		To see the YAML for this workflow, try: gh workflow view "goreleaser" --yaml
  `)

	cs := opts.IO.ColorScheme()

	fmt.Fprintf(opts.IO.Out, mock,
		cs.Bold("goreleaser"),
		cs.Cyan("447058545"),
		cs.Gray("releases.yml"),
		cs.Bold("Recent Runs"),
		cs.Yellow("-"),
		cs.Cyan("647332420"),
		cs.Green("✓"),
		cs.Cyan("647332419"),
		cs.Red("X"),
		cs.Cyan("647332418"),
	)

	return nil
}
