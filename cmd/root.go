package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/adrg/xdg"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type MarkData struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

var rootCmd = &cobra.Command{
	Use:               "fmark <name>",
	Short:             "Exec a command",
	SilenceUsage:      true,
	ValidArgsFunction: getBookmarkValidArgs,
	Args:              cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if jsonData, err := loadMarks(); err != nil {
			return err
		} else {
			i := slices.IndexFunc(jsonData, func(m MarkData) bool {
				return name == m.Name
			})
			if i != -1 {
				parts := strings.Fields(jsonData[i].Value)
				command := exec.Command(parts[0], parts[1:]...)
				command.Stdout = os.Stdout
				command.Stderr = os.Stderr
				command.Stdin = os.Stdin
				return command.Run()
			} else {
				return fmt.Errorf("bookmark %q not found", name)
			}
		}
	},
}
var fs = afero.NewOsFs()

func init() {
	rootCmd.PersistentFlags().Bool("verbose", false, "verbose log")
}

func Execute() {
	if xdg.DataHome == "" {
		fmt.Fprintln(os.Stderr, "Environment variable $XDG_DATA_HOME is empty")
		os.Exit(1)
	} else if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func loadMarks() ([]MarkData, error) {
	path, _ := xdg.DataFile("fmark/commands.json")
	if exists, err := afero.Exists(fs, path); err != nil || !exists {
		afero.WriteFile(fs, path, []byte("[]"), 0644)
	}
	data, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, err
	}
	var marks []MarkData
	return marks, json.Unmarshal(data, &marks)
}

func getBookmarkValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	path, _ := xdg.DataFile("fmark/commands.json")
	if exists, err := afero.Exists(fs, path); err != nil {
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	} else if exists {
		var jsonData []MarkData
		data, _ := afero.ReadFile(fs, path)
		if err := json.Unmarshal(data, &jsonData); err != nil {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}
		var items []string
		for _, mark := range jsonData {
			items = append(items, mark.Name)
		}
		return items, cobra.ShellCompDirectiveNoFileComp
	}
	return []string{}, cobra.ShellCompDirectiveNoFileComp
}

func (m MarkData) String() string {
	return fmt.Sprintf("%s - %q", m.Name, m.Value)
}
