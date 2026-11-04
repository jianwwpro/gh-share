package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yuler/gh-share/internal/skill"
)

var forceInstall bool

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage the AI/Agent share skill globally",
	Long:  "Install, verify status, or uninstall the share skill across agent systems (~/.agents, ~/.gemini, ~/.claude).",
}

var skillInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the share skill globally for AI coding assistants",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🚀 Installing share skill globally...")
		statuses, err := skill.Install(forceInstall)
		if err != nil {
			return err
		}

		printSkillStatus(statuses)
		fmt.Println("🎉 Skill installed! Your agents (Antigravity, Gemini CLI, Claude Code) can now use the 'share' skill.")
		return nil
	},
}

var skillStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check installation status of the share skill",
	RunE: func(cmd *cobra.Command, args []string) error {
		statuses, err := skill.GetLocations()
		if err != nil {
			return err
		}
		printSkillStatus(statuses)
		return nil
	},
}

var skillUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the share skill",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🧹 Uninstalling share skill...")
		statuses, err := skill.Uninstall()
		if err != nil {
			return err
		}
		printSkillStatus(statuses)
		return nil
	},
}

func printSkillStatus(statuses []skill.LocationStatus) {
	fmt.Printf("\n%-40s  %-10s  %s\n", "TARGET LOCATION", "STATUS", "DETAILS")
	fmt.Printf("%-40s  %-10s  %s\n", "---------------", "------", "-------")
	for _, st := range statuses {
		statStr := "❌ Missing"
		if st.Installed {
			statStr = "✅ Active"
		}
		fmt.Printf("%-40s  %-10s  %s\n", st.Name, statStr, st.Note)
	}
	fmt.Println()
}

func init() {
	skillInstallCmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "Overwrite existing symlinks or files")

	skillCmd.AddCommand(skillInstallCmd)
	skillCmd.AddCommand(skillStatusCmd)
	skillCmd.AddCommand(skillUninstallCmd)

	rootCmd.AddCommand(skillCmd)
}
