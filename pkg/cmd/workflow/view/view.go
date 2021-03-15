package view

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/internal/ghrepo"
	"github.com/cli/cli/pkg/cmd/workflow/shared"
	"github.com/cli/cli/pkg/cmdutil"
	"github.com/cli/cli/pkg/iostreams"
	"github.com/cli/cli/pkg/markdown"
	"github.com/spf13/cobra"
)

type ViewOptions struct {
	HttpClient func() (*http.Client, error)
	IO         *iostreams.IOStreams
	BaseRepo   func() (ghrepo.Interface, error)

	WorkflowID string
	YAML       string

	Prompt bool
	Raw    bool
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

			opts.Raw = !opts.IO.CanPrompt()

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
	workflowID := opts.WorkflowID
	var workflow *shared.Workflow

	fmt.Printf("DBG %#v\n", workflowID)
	fmt.Printf("DBG %#v\n", workflow)

	if opts.Prompt {
		// TODO grab workflows
		// TODO prompt
	}

	if workflow == nil {
		// TODO REST GET workflowID
	}

	// TODO figure out how to get the yaml at default branch of remote
	// TODO lay out yaml with syntax highlighting

	yaml := "# TODO YAML"

	theme := opts.IO.DetectTerminalTheme()
	markdownStyle := markdown.GetStyle(theme)
	if err := opts.IO.StartPager(); err != nil {
		fmt.Fprintf(opts.IO.ErrOut, "starting pager failed: %v\n", err)
	}
	defer opts.IO.StopPager()

	if !opts.Raw {
		// TODO print little header
		codeBlock := fmt.Sprintf("```yaml\n%s\n```", yaml)
		// TODO fix indentation
		rendered, err := markdown.Render(codeBlock, markdownStyle, "")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(opts.IO.Out, rendered)
		return err
	}

	if _, err := fmt.Fprint(opts.IO.Out, yaml); err != nil {
		return err
	}

	if !strings.HasSuffix(yaml, "\n") {
		_, err := fmt.Fprint(opts.IO.Out, "\n")
		return err
	}

	return nil
}
