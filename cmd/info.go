package cmd

import (
	"fmt"
	"os"
	"umaru/internal/templates"
	"umaru/internal/ui"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info [template-id]",
	Short: "Inspect template architecture, directory tree, ports, and metadata",
	Long: `Inspect provides in-depth technical details about any starter template,
including its default ports, install/run commands, supported addons, and a full
ASCII directory tree of the generated project structure.`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		tmpls, err := templates.GetAvailableTemplates()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		var ids []string
		for _, t := range tmpls {
			ids = append(ids, t.ID+"\t"+t.Name)
		}
		return ids, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		var selectedID string

		if len(args) > 0 {
			selectedID = args[0]
		} else {
			// Interactive Selection
			allTemplates, err := templates.GetAvailableTemplates()
			if err != nil {
				fmt.Printf("❌ Failed to load templates: %v\n", err)
				os.Exit(1)
			}

			if len(allTemplates) == 0 {
				fmt.Println("No templates found.")
				return
			}

			var options []huh.Option[string]
			for _, tmpl := range allTemplates {
				label := fmt.Sprintf("%-28s [%s]", tmpl.Name, tmpl.ID)
				options = append(options, huh.NewOption(label, tmpl.ID))
			}

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("🔍 Select a template to inspect").
						Description("Explore directory trees, default ports, and architecture").
						Options(options...).
						Value(&selectedID),
				),
			)

			err = form.Run()
			if err != nil {
				// User cancelled prompt
				return
			}
		}

		info, err := templates.GetTemplateInfo(selectedID)
		if err != nil {
			fmt.Printf("❌ Template '%s' not found.\nRun 'umaru list' to see all available templates.\n", selectedID)
			os.Exit(1)
		}

		ui.PrintTemplateInfoCard(info)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
