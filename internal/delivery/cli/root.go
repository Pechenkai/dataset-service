package cli

import (
	"context"
	"fmt"
	"os"
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

	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
	}

	cobra.OnInitialize(func() {
		if err := root.ExecuteContext(context.Background()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	})

	return root
}
