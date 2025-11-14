package cli

import (
	"ppo/internal/delivery/cli/api"
	"ppo/internal/delivery/cli/commands"

	"github.com/spf13/cobra"
)

func NewRootCommand(apiClient *api.Client) *cobra.Command {
	root := &cobra.Command{
		Use:   "techui",
		Short: "TechUI CLI for ML Dataset Catalog",
	}

	root.AddCommand(
		commands.NewCategoryCommand(apiClient),
		commands.NewDatasetCommand(apiClient),
		commands.NewNotificationCommand(apiClient),
		commands.NewReviewCommand(apiClient),
		commands.NewUserCommand(apiClient),
		commands.NewSubscriptionCommand(apiClient),
	)

	return root
}
