package cli

// init registers all AI-agent setup commands with the root command in canonical order.
// To add support for a new agent, add an entry to the agents map in agent_config.go;
// registration happens automatically here without touching this file.
func init() {
	for _, key := range agentOrder {
		rootCmd.AddCommand(makeSetupCmd(agents[key]))
	}
}
