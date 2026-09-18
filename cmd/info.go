package cmd

import (
	"encoding/json"
	"fmt"
	"umaru/internal/templates"
	"umaru/internal/ui"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var infoJSONFlag bool

var infoCmd = &cobra.Command{
	Use:     "info [template-id]",
	Short:   "Inspect template architecture, directory tree, ports, and metadata",
	GroupID: "core",
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
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedID string

		if len(args) > 0 {
			selectedID = args[0]
		} else {
			// Interactive Selection
			allTemplates, err := templates.GetAvailableTemplates()
			if err != nil {
				return fmt.Errorf("failed to load templates: %w", err)
			}

			if len(allTemplates) == 0 {
				fmt.Println("No templates found.")
				return nil
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
				return nil
			}
		}

		info, err := templates.GetTemplateInfo(selectedID)
		if err != nil {
			return fmt.Errorf("template '%s' not found. Run 'umaru list' to see all available templates", selectedID)
		}

		if infoJSONFlag {
			data, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to serialize template info to json: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		ui.PrintTemplateInfoCard(info)
		return nil
	},
}

func init() {
	infoCmd.Flags().BoolVar(&infoJSONFlag, "json", false, "Output template details in JSON format")
	rootCmd.AddCommand(infoCmd)
}
