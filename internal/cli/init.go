package cli

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/zulfikawr/gosp/pkg/config"
	"github.com/zulfikawr/gosp/pkg/tokens"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize GOSP node and configuration",
	Run: func(cmd *cobra.Command, args []string) {
		runInit()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit() {
	fmt.Println("🦞 Welcome to the GOSP Initialization Wizard!")
	fmt.Println("--------------------------------------------")

	cfg := &config.Config{}

	// 1. Role Selection
	rolePrompt := &survey.Select{
		Message: "What is this node's role?",
		Options: []string{"master", "worker", "both"},
		Default: "worker",
	}
	survey.AskOne(rolePrompt, &cfg.Role)

	// 2. Node Identity
	idPrompt := &survey.Input{
		Message: "Enter a unique name for this node (leave blank for random):",
	}
	survey.AskOne(idPrompt, &cfg.NodeID)
	if cfg.NodeID == "" {
		cfg.NodeID = "node-" + uuid.New().String()[:8]
	}

	regionPrompt := &survey.Input{
		Message: "Which region are you in? (e.g., US-East, ID-Jakarta):",
		Default: "ID-Jakarta",
	}
	survey.AskOne(regionPrompt, &cfg.Region)

	// 3. Network Configuration
	if cfg.Role == "master" || cfg.Role == "both" {
		httpPrompt := &survey.Input{
			Message: "Enter HTTP API port:",
			Default: "19000",
		}
		survey.AskOne(httpPrompt, &cfg.HTTPPort)

		grpcPrompt := &survey.Input{
			Message: "Enter gRPC port:",
			Default: "19004",
		}
		survey.AskOne(grpcPrompt, &cfg.GRPCPort)

		genToken := false
		tokenGenPrompt := &survey.Confirm{
			Message: "Would you like to generate a secure Join Token for new workers?",
			Default: true,
		}
		survey.AskOne(tokenGenPrompt, &genToken)
		if genToken {
			cfg.JoinToken, _ = tokens.Generate()
			fmt.Printf("✅ Generated Join Token: %s\n", cfg.JoinToken)
			fmt.Println("   (Save this! New workers will need it to join your cluster)")
		}
	}

	if cfg.Role == "worker" || cfg.Role == "both" {
		masterPrompt := &survey.Input{
			Message: "Enter Master gRPC address (e.g., localhost:19004):",
			Default: "localhost:19004",
		}
		survey.AskOne(masterPrompt, &cfg.MasterURL)

		tokenPrompt := &survey.Input{
			Message: "Enter Join Token (if required):",
		}
		survey.AskOne(tokenPrompt, &cfg.JoinToken)
	}

	// 4. Save Configuration
	err := config.Save(cfg)
	if err != nil {
		fmt.Printf("❌ Failed to save configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✨ GOSP successfully initialized!")
	fmt.Printf("Config saved to: %s\n", config.GetConfigPath())
	fmt.Println("You can now run 'gosp master' or 'gosp worker' to start your node.")
}
