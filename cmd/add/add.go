package add

import "github.com/spf13/cobra"

var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add projects to the index",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
