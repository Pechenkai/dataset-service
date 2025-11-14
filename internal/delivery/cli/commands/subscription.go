package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"ppo/internal/delivery/cli/api"
)

func NewSubscriptionCommand(client *api.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscription",
		Short: "Manage dataset subscriptions",
	}

	subscribeCmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe a user to a dataset",
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, _ := cmd.Flags().GetUint64("user")
			datasetID, _ := cmd.Flags().GetUint64("dataset")
			if err := client.Subscribe(context.Background(), userID, datasetID); err != nil {
				return err
			}
			fmt.Printf("User %d subscribed to dataset %d\n", userID, datasetID)
			return nil
		},
	}
	subscribeCmd.Flags().Uint64("user", 0, "User ID (required)")
	subscribeCmd.Flags().Uint64("dataset", 0, "Dataset ID (required)")
	subscribeCmd.MarkFlagRequired("user")
	subscribeCmd.MarkFlagRequired("dataset")

	unsubscribeCmd := &cobra.Command{
		Use:   "unsubscribe",
		Short: "Unsubscribe a user from a dataset",
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, _ := cmd.Flags().GetUint64("user")
			datasetID, _ := cmd.Flags().GetUint64("dataset")
			if err := client.Unsubscribe(context.Background(), userID, datasetID); err != nil {
				return err
			}
			fmt.Printf("User %d unsubscribed from dataset %d\n", userID, datasetID)
			return nil
		},
	}
	unsubscribeCmd.Flags().Uint64("user", 0, "User ID (required)")
	unsubscribeCmd.Flags().Uint64("dataset", 0, "Dataset ID (required)")
	unsubscribeCmd.MarkFlagRequired("user")
	unsubscribeCmd.MarkFlagRequired("dataset")

	listSubsCmd := &cobra.Command{
		Use:   "list-subscriptions",
		Short: "List dataset IDs a user is subscribed to",
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, _ := cmd.Flags().GetUint64("user")
			list, err := client.ListSubscriptions(context.Background(), userID)
			if err != nil {
				return err
			}
			if len(list) == 0 {
				fmt.Println("No subscriptions found.")
				return nil
			}
			for _, sub := range list {
				fmt.Printf("Dataset %d (since %s)\n", sub.DatasetID, sub.CreatedAt.Format(time.RFC822))
			}
			return nil
		},
	}
	listSubsCmd.Flags().Uint64("user", 0, "User ID (required)")
	listSubsCmd.MarkFlagRequired("user")

	listUsersCmd := &cobra.Command{
		Use:   "list-subscribers",
		Short: "List user IDs subscribed to a dataset",
		RunE: func(cmd *cobra.Command, args []string) error {
			datasetID, _ := cmd.Flags().GetUint64("dataset")
			list, err := client.ListSubscribers(context.Background(), datasetID)
			if err != nil {
				return err
			}
			if len(list) == 0 {
				fmt.Println("No subscribers found.")
				return nil
			}
			for _, sub := range list {
				fmt.Printf("User %d\n", sub.UserID)
			}
			return nil
		},
	}
	listUsersCmd.Flags().Uint64("dataset", 0, "Dataset ID (required)")
	listUsersCmd.MarkFlagRequired("dataset")

	cmd.AddCommand(subscribeCmd, unsubscribeCmd, listSubsCmd, listUsersCmd)
	return cmd
}
