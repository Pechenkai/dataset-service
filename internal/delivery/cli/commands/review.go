package commands

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"ppo/internal/delivery/cli/api"
	"ppo/internal/entities"
	"ppo/internal/services"
)

func NewReviewCommand(client *api.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Review operations",
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new review",
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, _ := cmd.Flags().GetUint64("user")
			datasetID, _ := cmd.Flags().GetUint64("dataset")
			ratingVal, _ := cmd.Flags().GetInt("rating")
			text, _ := cmd.Flags().GetString("text")

			if ratingVal < int(entities.Rating1) || ratingVal > int(entities.Rating5) {
				return errors.New("rating must be between 1 and 5")
			}

			crcmd := services.CreateReviewCmd{
				UserID:    userID,
				DatasetID: datasetID,
				Rating:    entities.Rating(ratingVal),
				Text:      text,
			}
			revID, err := client.CreateReview(context.Background(), crcmd)
			if err != nil {
				return err
			}
			fmt.Printf("Review created with ID: %d\n", revID)
			return nil
		},
	}
	createCmd.Flags().Uint64("user", 0, "User ID (required)")
	createCmd.Flags().Uint64("dataset", 0, "Dataset ID (required)")
	createCmd.Flags().Int("rating", 0, "Rating (1-5) (required)")
	createCmd.Flags().String("text", "", "Review text")
	createCmd.MarkFlagRequired("user")
	createCmd.MarkFlagRequired("dataset")
	createCmd.MarkFlagRequired("rating")

	updateCmd := &cobra.Command{
		Use:   "update [id]",
		Short: "Update an existing review",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			revID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			ratingVal, _ := cmd.Flags().GetInt("rating")
			text, _ := cmd.Flags().GetString("text")

			if ratingVal != 0 && (ratingVal < int(entities.Rating1) || ratingVal > int(entities.Rating5)) {
				return errors.New("rating must be between 1 and 5")
			}

			updcmd := services.UpdateReviewCmd{
				ReviewID: revID,
				Rating:   entities.Rating(ratingVal),
				Text:     text,
			}
			return client.UpdateReview(context.Background(), updcmd)
		},
	}
	updateCmd.Flags().Int("rating", 0, "New rating (1-5)")
	updateCmd.Flags().String("text", "", "New review text")

	deleteCmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a review by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			revID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			return client.DeleteReview(context.Background(), revID)
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List reviews by dataset or user",
		RunE: func(cmd *cobra.Command, args []string) error {
			dsID, _ := cmd.Flags().GetUint64("dataset")
			usrID, _ := cmd.Flags().GetUint64("user")

			if dsID != 0 && usrID != 0 {
				return errors.New("use only one of --dataset or --user")
			}
			var reviews []*entities.Review
			var err error
			if dsID != 0 {
				reviews, err = client.ListReviewsByDataset(context.Background(), dsID)
			} else if usrID != 0 {
				reviews, err = client.ListReviewsByUser(context.Background(), usrID)
			} else {
				return errors.New("need --dataset or --user flag")
			}
			if err != nil {
				return err
			}
			for _, r := range reviews {
				fmt.Printf("%d: user %d, dataset %d, rating %d, text: %s\n",
					r.ID, r.UserID, r.DatasetID, r.Rating, r.Text)
			}
			return nil
		},
	}
	listCmd.Flags().Uint64("dataset", 0, "Dataset ID to list reviews for")
	listCmd.Flags().Uint64("user", 0, "User ID to list reviews by")

	summaryCmd := &cobra.Command{
		Use:   "summary [datasetID]",
		Short: "Show rating summary for a dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dsID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			summary, err := client.GetRatingSummary(context.Background(), dsID)
			if err != nil {
				return err
			}
			fmt.Printf("Average rating: %.2f, Count: %d\n", summary.Average, summary.Count)
			return nil
		},
	}

	cmd.AddCommand(createCmd, updateCmd, deleteCmd, listCmd, summaryCmd)
	return cmd
}
