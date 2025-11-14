package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"ppo/internal/delivery/cli/api"
)

func NewNotificationCommand(client *api.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notification",
		Short: "Notification operations",
	}

	notifyCmd := &cobra.Command{
		Use:   "notify",
		Short: "Notify all subscribers of a dataset",
		RunE: func(cmd *cobra.Command, args []string) error {
			dsID, _ := cmd.Flags().GetUint64("dataset")
			message, _ := cmd.Flags().GetString("message")
			count, err := client.NotifySubscribers(context.Background(), dsID, message)
			if err != nil {
				return err
			}
			fmt.Printf("Sent notifications to %d subscribers of dataset %d\n", count, dsID)
			return nil
		},
	}
	notifyCmd.Flags().Uint64("dataset", 0, "Dataset ID to notify (required)")
	notifyCmd.Flags().String("message", "", "Notification message (required)")
	notifyCmd.MarkFlagRequired("dataset")
	notifyCmd.MarkFlagRequired("message")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List notifications for a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, _ := cmd.Flags().GetUint64("user")
			notifs, err := client.ListNotifications(context.Background(), userID)
			if err != nil {
				return err
			}
			if len(notifs) == 0 {
				fmt.Println("No notifications found.")
				return nil
			}
			for _, n := range notifs {
				status := "unread"
				if n.IsRead {
					status = "read"
				}
				fmt.Printf("%d: dataset %d — %s [%s] at %s\n",
					n.ID, n.DatasetID, n.Message, status, n.CreatedAt.Format("2006-01-02 15:04:05"))
			}
			return nil
		},
	}
	listCmd.Flags().Uint64("user", 0, "User ID (required)")
	listCmd.MarkFlagRequired("user")

	markCmd := &cobra.Command{
		Use:   "mark-read [id]",
		Short: "Mark a notification as read",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			if err := client.MarkNotification(context.Background(), id, true); err != nil {
				return err
			}
			fmt.Printf("Notification %d marked as read\n", id)
			return nil
		},
	}

	cmd.AddCommand(notifyCmd, listCmd, markCmd)
	return cmd
}
