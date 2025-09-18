package cli

import (
	"ppo/internal/delivery/cli/commands"

	"github.com/spf13/cobra"
	"ppo/internal/services"
)

func NewRootCommand(
	catSvc services.CategoryService,
	dsSvc services.DatasetService,
	notifSvc services.NotificationService,
	revSvc services.ReviewService,
	userSvc services.UserService,
	subSvc services.SubscriptionService,
) *cobra.Command {
	root := &cobra.Command{
		Use:   "techui",
		Short: "TechUI CLI for ML Dataset Catalog",
	}

	root.AddCommand(
		commands.NewCategoryCommand(catSvc),
		commands.NewDatasetCommand(dsSvc),
		commands.NewNotificationCommand(notifSvc),
		commands.NewReviewCommand(revSvc),
		commands.NewUserCommand(userSvc),
		commands.NewSubscriptionCommand(subSvc),
	)

	return root
}
