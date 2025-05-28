package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"ppo/internal/services"
)

func NewCategoryCommand(svc services.CategoryService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "category",
		Short: "Category operations",
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new category",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			desc, _ := cmd.Flags().GetString("desc")
			cmdCat := services.CreateCategoryCmd{
				Name:        name,
				Description: desc,
			}
			id, err := svc.CreateCategory(context.Background(), cmdCat)
			if err != nil {
				return err
			}
			fmt.Printf("Category created with ID: %d\n", id)
			return nil
		},
	}
	createCmd.Flags().StringP("name", "n", "", "Category name (required)")
	createCmd.Flags().StringP("desc", "d", "", "Category description")
	createCmd.MarkFlagRequired("name")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all categories",
		RunE: func(cmd *cobra.Command, args []string) error {
			cats, err := svc.ListCategories(context.Background())
			if err != nil {
				return err
			}
			for _, c := range cats {
				fmt.Printf("%d: %s - %s\n", c.ID, c.Name, c.Description)
			}
			return nil
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update [id]",
		Short: "Update a category by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			name, _ := cmd.Flags().GetString("name")
			desc, _ := cmd.Flags().GetString("desc")
			return svc.UpdateCategory(context.Background(), services.UpdateCategoryCmd{
				ID:          id,
				Name:        name,
				Description: desc,
			})
		},
	}
	updateCmd.Flags().StringP("name", "n", "", "New category name")
	updateCmd.Flags().StringP("desc", "d", "", "New category description")

	deleteCmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a category by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			if err := svc.DeleteCategory(context.Background(), id); err != nil {
				return err
			}
			fmt.Printf("Category %d deleted\n", id)
			return nil
		},
	}

	cmd.AddCommand(createCmd, listCmd, deleteCmd)
	return cmd
}
