package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"ppo/internal/services"
)

func NewUserCommand(svc services.UserService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "User operations",
	}

	registerCmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			username, _ := cmd.Flags().GetString("username")
			email, _ := cmd.Flags().GetString("email")
			password, _ := cmd.Flags().GetString("password")
			country, _ := cmd.Flags().GetString("country")
			role, _ := cmd.Flags().GetString("role")

			newID, err := svc.Register(context.Background(), services.RegisterUserCmd{
				Username: username,
				Email:    email,
				Password: password,
				Country:  country,
				Role:     role,
			})
			if err != nil {
				return err
			}
			fmt.Printf("User registered with ID: %d\n", newID)
			return nil
		},
	}
	registerCmd.Flags().String("username", "", "Username (required)")
	registerCmd.Flags().String("email", "", "Email (required)")
	registerCmd.Flags().String("password", "", "Password (required)")
	registerCmd.Flags().String("country", "", "Country")
	registerCmd.Flags().String("role", "user", "Role (guest, user, admin)")
	registerCmd.MarkFlagRequired("username")
	registerCmd.MarkFlagRequired("email")
	registerCmd.MarkFlagRequired("password")

	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			email, _ := cmd.Flags().GetString("email")
			password, _ := cmd.Flags().GetString("password")

			user, err := svc.Authenticate(context.Background(), services.AuthenticateUserCmd{
				Email:    email,
				Password: password,
			})
			if err != nil {
				return err
			}
			fmt.Printf("Authenticated user: ID=%d, Username=%s, Email=%s, Role=%s\n",
				user.ID, user.Username, user.Email, user.Role)
			return nil
		},
	}
	loginCmd.Flags().String("email", "", "Email (required)")
	loginCmd.Flags().String("password", "", "Password (required)")
	loginCmd.MarkFlagRequired("email")
	loginCmd.MarkFlagRequired("password")

	getCmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get user by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			user, err := svc.GetUserByID(context.Background(), id)
			if err != nil {
				return err
			}
			fmt.Printf("User ID=%d\n  Username: %s\n  Email: %s\n  Country: %s\n  Role: %s\n  Registered: %s\n",
				user.ID, user.Username, user.Email, user.Country, user.Role, user.RegistrationDate)
			return nil
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update [id]",
		Short: "Update an existing user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			username, _ := cmd.Flags().GetString("username")
			email, _ := cmd.Flags().GetString("email")
			password, _ := cmd.Flags().GetString("password")
			country, _ := cmd.Flags().GetString("country")
			role, _ := cmd.Flags().GetString("role")
			blocked, _ := cmd.Flags().GetBool("blocked")

			return svc.UpdateUser(context.Background(), services.UpdateUserCmd{
				ID:        id,
				Username:  username,
				Email:     email,
				Password:  password,
				Country:   country,
				Role:      role,
				IsBlocked: blocked,
			})
		},
	}
	updateCmd.Flags().String("username", "", "New username")
	updateCmd.Flags().String("email", "", "New email")
	updateCmd.Flags().String("password", "", "New password")
	updateCmd.Flags().String("country", "", "New country")
	updateCmd.Flags().String("role", "", "New role (guest, user, admin)")
	updateCmd.Flags().Bool("blocked", false, "Block or unblock user")

	deleteCmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a user by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			return svc.DeleteUser(context.Background(), id)
		},
	}

	cmd.AddCommand(registerCmd, loginCmd, getCmd, updateCmd, deleteCmd)
	return cmd
}
