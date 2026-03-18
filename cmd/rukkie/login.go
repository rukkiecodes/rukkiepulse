package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/rukkiecodes/rukkiepulse/internal/auth"
	"github.com/rukkiecodes/rukkiepulse/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to RukkiePulse",
	RunE:  runLogin,
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of RukkiePulse",
	RunE:  runLogout,
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	output.PrintPasswordPrompt()

	var password string
	if term.IsTerminal(int(syscall.Stdin)) {
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			output.PrintError("failed to read password")
			return nil
		}
		password = string(passwordBytes)
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		password = strings.TrimSpace(scanner.Text())
		fmt.Println()
	}

	if err := auth.Login(password); err != nil {
		output.PrintError(err.Error())
		return nil
	}

	output.PrintLoginSuccess()
	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	if err := auth.Logout(); err != nil {
		output.PrintError(err.Error())
		return nil
	}
	output.PrintLogout()
	return nil
}
