package commands

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"ppo/internal/services"
)

func NewSubscriptionCommand(svc services.SubscriptionService) *cobra.Command {
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
			if err := svc.Subscribe(context.Background(), userID, datasetID); err != nil {
				return err
			}
			fmt.Printf("User %d subscribed to dataset %d\n", userID, datasetID)
			return nil
		},
	}
	subscribeCmd.Flags().Uint64("user", 0, "User ID to subscribe (required)")
	subscribeCmd.Flags().Uint64("dataset", 0, "Dataset ID to subscribe to (required)")
	subscribeCmd.MarkFlagRequired("user")
	subscribeCmd.MarkFlagRequired("dataset")

	unsubscribeCmd := &cobra.Command{
		Use:   "unsubscribe",
		Short: "Unsubscribe a user from a dataset",
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, _ := cmd.Flags().GetUint64("user")
			datasetID, _ := cmd.Flags().GetUint64("dataset")
			if err := svc.Unsubscribe(context.Background(), userID, datasetID); err != nil {
				return err
			}
			fmt.Printf("User %d unsubscribed from dataset %d\n", userID, datasetID)
			return nil
		},
	}
	unsubscribeCmd.Flags().Uint64("user", 0, "User ID to unsubscribe (required)")
	unsubscribeCmd.Flags().Uint64("dataset", 0, "Dataset ID to unsubscribe from (required)")
	unsubscribeCmd.MarkFlagRequired("user")
	unsubscribeCmd.MarkFlagRequired("dataset")

	listSubsCmd := &cobra.Command{
		Use:   "list-subscriptions",
		Short: "List dataset IDs a user is subscribed to",
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, _ := cmd.Flags().GetUint64("user")
			dsIDs, err := svc.ListSubscriptions(context.Background(), userID)
			if err != nil {
				return err
			}
			if len(dsIDs) == 0 {
				fmt.Println("No subscriptions found.")
				return nil
			}
			fmt.Printf("User %d is subscribed to datasets: %v\n", userID, dsIDs)
			return nil
		},
	}
	listSubsCmd.Flags().Uint64("user", 0, "User ID to list subscriptions for (required)")
	listSubsCmd.MarkFlagRequired("user")

	listUsersCmd := &cobra.Command{
		Use:   "list-subscribers",
		Short: "List user IDs subscribed to a dataset",
		RunE: func(cmd *cobra.Command, args []string) error {
			datasetID, _ := cmd.Flags().GetUint64("dataset")
			userIDs, err := svc.ListSubscribers(context.Background(), datasetID)
			if err != nil {
				return err
			}
			if len(userIDs) == 0 {
				fmt.Println("No subscribers found.")
				return nil
			}
			fmt.Printf("Dataset %d has subscribers: %v\n", datasetID, userIDs)
			return nil
		},
	}
	listUsersCmd.Flags().Uint64("dataset", 0, "Dataset ID to list subscribers for (required)")
	listUsersCmd.MarkFlagRequired("dataset")

	cmd.AddCommand(subscribeCmd, unsubscribeCmd, listSubsCmd, listUsersCmd)
	return cmd
}
