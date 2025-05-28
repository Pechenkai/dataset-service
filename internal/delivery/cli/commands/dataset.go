package commands

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"ppo/internal/services"
)

func NewDatasetCommand(svc services.DatasetService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataset",
		Short: "Dataset operations",
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new dataset",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			desc, _ := cmd.Flags().GetString("desc")
			catID, _ := cmd.Flags().GetUint64("category")
			isPublic, _ := cmd.Flags().GetBool("public")
			filePath, _ := cmd.Flags().GetString("file")

			f, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer f.Close()
			fi, _ := f.Stat()

			cmdSvc := services.CreateDatasetCmd{
				Name:        name,
				Description: desc,
				CategoryID:  catID,
				IsPublic:    isPublic,
				FileName:    fi.Name(),
				MetaFormat:  "",
				MetaTags:    "",
				MetaSize:    0,
			}
			id, err := svc.CreateDataset(context.Background(), cmdSvc, f, fi.Size())
			if err != nil {
				return err
			}
			fmt.Printf("Dataset created with ID: %d\n", id)
			return nil
		},
	}
	createCmd.Flags().StringP("name", "n", "", "Dataset name (required)")
	createCmd.Flags().StringP("desc", "d", "", "Dataset description")
	createCmd.Flags().Uint64P("category", "c", 0, "Category ID (required)")
	createCmd.Flags().BoolP("public", "p", false, "Make dataset public")
	createCmd.Flags().StringP("file", "f", "", "Path to data file (required)")
	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("category")
	createCmd.MarkFlagRequired("file")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List datasets (all, by user, or public)",
		RunE: func(cmd *cobra.Command, args []string) error {
			userIDVal, _ := cmd.Flags().GetUint64("user")
			publicOnly, _ := cmd.Flags().GetBool("public")

			var ownerID *uint64
			if userIDVal != 0 {
				ownerID = &userIDVal
			}

			list, err := svc.ListDatasets(context.Background(), publicOnly, ownerID)
			if err != nil {
				return err
			}

			if len(list) == 0 {
				fmt.Println("No datasets found.")
				return nil
			}

			for _, d := range list {
				fmt.Printf("%d: %s (cat %d) by user %d public=%v\n",
					d.ID, d.Name, d.CategoryID, d.OwnerID, d.IsPublic)
			}
			return nil
		},
	}
	listCmd.Flags().Uint64("user", 0, "Filter by owner user ID")
	listCmd.Flags().Bool("public", false, "List only public datasets")

	addVerCmd := &cobra.Command{
		Use:   "add-version [datasetID]",
		Short: "Add a new version to a dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dsID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			changelog, _ := cmd.Flags().GetString("changelog")
			filePath, _ := cmd.Flags().GetString("file")
			format, _ := cmd.Flags().GetString("format")
			tags, _ := cmd.Flags().GetString("tags")
			sizeFlag, _ := cmd.Flags().GetUint64("size")

			f, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer f.Close()
			fi, _ := f.Stat()

			verID, err := svc.AddDatasetVersion(context.Background(),
				services.AddVersionCmd{
					DatasetID:  dsID,
					ChangeLog:  changelog,
					FileName:   fi.Name(),
					MetaFormat: format,
					MetaTags:   tags,
					MetaSize:   sizeFlag,
				},
				f, fi.Size(),
			)
			if err != nil {
				return err
			}
			fmt.Printf("Added version %d to dataset %d\n", verID, dsID)
			return nil
		},
	}
	addVerCmd.Flags().StringP("file", "f", "", "Path to version file (required)")
	addVerCmd.Flags().StringP("changelog", "c", "", "Changelog for this version")
	addVerCmd.Flags().String("format", "", "Metadata format")
	addVerCmd.Flags().String("tags", "", "Metadata tags")
	addVerCmd.Flags().Uint64("size", 0, "Metadata size")
	addVerCmd.MarkFlagRequired("file")

	cmd.AddCommand(createCmd, listCmd)
	return cmd
}
