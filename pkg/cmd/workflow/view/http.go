package view

import (
	"fmt"

	"github.com/cli/cli/api"
	"github.com/cli/cli/internal/ghrepo"
	"github.com/cli/cli/pkg/cmd/workflow/shared"
)

// func workflowByName(client *api.Client, name string)

func getWorkflow(client *api.Client, repo ghrepo.Interface, workflowID string) (*shared.Workflow, error) {
	var workflow shared.Workflow

	err := client.REST(repo.RepoHost(), "GET",
		fmt.Sprintf("repos/%s/actions/workflows/%s", ghrepo.FullName(repo), workflowID),
		nil, &workflow)

	if err != nil {
		return nil, err
	}

	return &workflow, nil
}
