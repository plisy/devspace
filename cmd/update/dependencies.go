package update

import (
	"github.com/loft-sh/devspace/cmd/flags"
	"github.com/loft-sh/devspace/pkg/devspace/dependency"
	"github.com/loft-sh/devspace/pkg/util/factory"
	"github.com/loft-sh/devspace/pkg/util/message"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// dependenciesCmd holds the cmd flags
type dependenciesCmd struct {
	*flags.GlobalFlags
}

// newDependenciesCmd creates a new command
func newDependenciesCmd(f factory.Factory, globalFlags *flags.GlobalFlags) *cobra.Command {
	cmd := &dependenciesCmd{GlobalFlags: globalFlags}

	dependenciesCmd := &cobra.Command{
		Use:   "dependencies",
		Short: "Updates the git repositories of the dependencies defined in the devspace.yaml",
		Long: `
#######################################################
############ devspace update dependencies #############
#######################################################
Updates the git repositories of the dependencies defined
in the devspace.yaml
#######################################################
	`,
		Args: cobra.NoArgs,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmd.RunDependencies(f, cobraCmd, args)
		},
	}

	return dependenciesCmd
}

// RunDependencies executes the functionality "devspace update dependencies"
func (cmd *dependenciesCmd) RunDependencies(f factory.Factory, cobraCmd *cobra.Command, args []string) error {
	// Set config root
	log := f.GetLog()
	configOptions := cmd.ToConfigOptions(log)
	configLoader := f.NewConfigLoader(cmd.ConfigPath)
	configExists, err := configLoader.SetDevSpaceRoot(log)
	if err != nil {
		return err
	}
	if !configExists {
		return errors.New(message.ConfigNotFound)
	}

	// Get the config
	config, err := configLoader.Load(configOptions, log)
	if err != nil {
		return err
	}

	// Update dependencies
	err = dependency.NewManager(config, nil, configOptions, log).UpdateAll()
	if err != nil {
		return err
	}

	log.Donef("Successfully updated all dependencies")
	return nil
}
