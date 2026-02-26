package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kfroemming/ralphkit/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)

	// Load viper config.
	home, err := os.UserHomeDir()
	if err == nil {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(filepath.Join(home, ".ralphkit"))
		viper.ReadInConfig()
	}
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage ralphkit configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		keys := viper.AllKeys()
		if len(keys) == 0 {
			ui.Dim("No configuration set. Run 'ralphkit install' to create defaults.")
			return nil
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("%s = %v\n", k, viper.Get(k))
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		viper.Set(key, value)

		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configDir := filepath.Join(home, ".ralphkit")
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return err
		}

		configFile := filepath.Join(configDir, "config.yaml")

		// Read existing config, update key, write back.
		lines, _ := readLines(configFile)
		updated := false
		for i, line := range lines {
			if strings.HasPrefix(line, key+":") || strings.HasPrefix(line, key+" :") {
				lines[i] = fmt.Sprintf("%s: %s", key, value)
				updated = true
				break
			}
		}
		if !updated {
			lines = append(lines, fmt.Sprintf("%s: %s", key, value))
		}

		content := strings.Join(lines, "\n") + "\n"
		if err := os.WriteFile(configFile, []byte(content), 0o644); err != nil {
			return err
		}
		ui.Success(fmt.Sprintf("Set %s = %s", key, value))
		return nil
	},
}

func readLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	return lines, nil
}
