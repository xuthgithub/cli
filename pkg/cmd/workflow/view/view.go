package view

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/api"
	"github.com/cli/cli/internal/ghrepo"
	"github.com/cli/cli/pkg/cmd/workflow/shared"
	"github.com/cli/cli/pkg/cmdutil"
	"github.com/cli/cli/pkg/iostreams"
	"github.com/cli/cli/pkg/markdown"
	"github.com/cli/cli/pkg/prompt"
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

			// TODO support --web

			if runF != nil {
				return runF(opts)
			}
			return runView(opts)
		},
	}

	return cmd
}

func runView(opts *ViewOptions) error {
	c, err := opts.HttpClient()
	if err != nil {
		return fmt.Errorf("could not build http client: %w", err)
	}
	client := api.NewClientFromHTTP(c)

	repo, err := opts.BaseRepo()
	if err != nil {
		return fmt.Errorf("could not determine base repo: %w", err)
	}

	var workflow *shared.Workflow

	if opts.Prompt {
		workflow, err = promptWorkflows(client, repo)
		if err != nil {
			return err
		}
	}

	if workflow == nil {
		workflow, err = getWorkflow(client, repo, opts.WorkflowID)
		if err != nil {
			return err
		}
	}

	fmt.Printf("DBG %#v\n", workflow)

	// TODO figure out how to get the yaml at default branch of remote
	// TODO lay out yaml with syntax highlighting
	// TODO consider attempting to figure out from what project a workflow originates

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

func promptWorkflows(client *api.Client, repo ghrepo.Interface) (*shared.Workflow, error) {
	workflows, err := shared.GetWorkflows(client, repo, 10)
	if len(workflows) == 0 {
		err = errors.New("no workflows are enabled")
	}

	if err != nil {
		var httpErr api.HTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode == 404 {
			err = errors.New("no workflows are enabled")
		}

		return nil, fmt.Errorf("could not fetch workflows for %s: %w", ghrepo.FullName(repo), err)
	}

	filtered := []shared.Workflow{}
	candidates := []string{}
	for _, workflow := range workflows {
		if !workflow.Disabled() {
			filtered = append(filtered, workflow)
			candidates = append(candidates, workflow.Name)
		}
	}

	var selected int

	err = prompt.SurveyAskOne(&survey.Select{
		Message:  "Select a workflow",
		Options:  candidates,
		PageSize: 10,
	}, &selected)
	if err != nil {
		return nil, err
	}

	return &filtered[selected], nil
}
