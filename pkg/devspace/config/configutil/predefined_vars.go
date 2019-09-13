package configutil

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	cloudconfig "github.com/devspace-cloud/devspace/pkg/devspace/cloud/config"
	cloudtoken "github.com/devspace-cloud/devspace/pkg/devspace/cloud/token"
	"github.com/devspace-cloud/devspace/pkg/util/git"
	"github.com/devspace-cloud/devspace/pkg/util/kubeconfig"
	"github.com/devspace-cloud/devspace/pkg/util/ptr"
	"github.com/devspace-cloud/devspace/pkg/util/randutil"
	"github.com/mgutz/ansi"
	"github.com/pkg/errors"
)

// PredefinedVars holds all predefined variables that can be used in the config
var PredefinedVars = map[string]*predefinedVarDefinition{
	"DEVSPACE_RANDOM": &predefinedVarDefinition{
		Fill: func(kubeContext string) (*string, error) {
			ret, err := randutil.GenerateRandomString(6)
			if err != nil {
				return nil, err
			}

			return &ret, nil
		},
	},
	"DEVSPACE_TIMESTAMP": &predefinedVarDefinition{
		Fill: func(kubeContext string) (*string, error) {
			return ptr.String(strconv.FormatInt(time.Now().Unix(), 10)), nil
		},
	},
	"DEVSPACE_GIT_COMMIT": &predefinedVarDefinition{
		ErrorMessage: "No git repository found, but predefined var DEVSPACE_GIT_COMMIT is used",
		Fill: func(kubeContext string) (*string, error) {
			gitRepo := git.NewGitRepository(".", "")

			hash, err := gitRepo.GetHash()
			if err != nil {
				return nil, nil
			}

			return ptr.String(hash[:8]), nil
		},
	},
	"DEVSPACE_SPACE": &predefinedVarDefinition{
		ErrorMessage: fmt.Sprintf("Current context is not a space, but predefined var DEVSPACE_SPACE is used.\n\nPlease run: \n- `%s` to create a new space\n- `%s` to use an existing space\n- `%s` to list existing spaces", ansi.Color("devspace create space [NAME]", "white+b"), ansi.Color("devspace use space [NAME]", "white+b"), ansi.Color("devspace list spaces", "white+b")),
		Fill: func(overrideKubeContext string) (*string, error) {
			kubeContext, err := kubeconfig.GetCurrentContext()
			if err != nil {
				return nil, nil
			}
			if overrideKubeContext != "" {
				kubeContext = overrideKubeContext
			}

			isSpace, err := kubeconfig.IsCloudSpace(kubeContext)
			if err != nil || !isSpace {
				return nil, nil
			}

			spaceID, providerName, err := kubeconfig.GetSpaceID(kubeContext)
			if err != nil {
				return nil, err
			}

			cloudConfigData, err := cloudconfig.ParseProviderConfig()
			if err != nil {
				return nil, nil
			}

			provider := cloudconfig.GetProvider(cloudConfigData, providerName)
			if provider == nil {
				return nil, nil
			}
			if provider.Spaces == nil {
				return nil, nil
			}
			if provider.Spaces[spaceID] == nil {
				return nil, nil
			}

			return &provider.Spaces[spaceID].Space.Name, nil
		},
	},
	"DEVSPACE_SPACE_NAMESPACE": &predefinedVarDefinition{
		ErrorMessage: fmt.Sprintf("Current context is not a space, but predefined var DEVSPACE_SPACE_NAMESPACE is used.\n\nPlease run: \n- `%s` to create a new space\n- `%s` to use an existing space\n- `%s` to list existing spaces", ansi.Color("devspace create space [NAME]", "white+b"), ansi.Color("devspace use space [NAME]", "white+b"), ansi.Color("devspace list spaces", "white+b")),
		Fill: func(overrideKubeContext string) (*string, error) {
			kubeContext, err := kubeconfig.GetCurrentContext()
			if err != nil {
				return nil, nil
			}
			if overrideKubeContext != "" {
				kubeContext = overrideKubeContext
			}

			isSpace, err := kubeconfig.IsCloudSpace(kubeContext)
			if err != nil || !isSpace {
				return nil, nil
			}

			spaceID, providerName, err := kubeconfig.GetSpaceID(kubeContext)
			if err != nil {
				return nil, err
			}

			cloudConfigData, err := cloudconfig.ParseProviderConfig()
			if err != nil {
				return nil, nil
			}

			provider := cloudconfig.GetProvider(cloudConfigData, providerName)
			if provider == nil {
				return nil, nil
			}
			if provider.Spaces == nil {
				return nil, nil
			}
			if provider.Spaces[spaceID] == nil {
				return nil, nil
			}

			return &provider.Spaces[spaceID].ServiceAccount.Namespace, nil
		},
	},
	"DEVSPACE_USERNAME": &predefinedVarDefinition{
		ErrorMessage: fmt.Sprintf("You are not logged into DevSpace Cloud, but predefined var DEVSPACE_USERNAME is used.\n\nPlease run: \n- `%s` to login into devspace cloud. Alternatively you can also remove the variable ${DEVSPACE_USERNAME} from your config", ansi.Color("devspace login", "white+b")),
		Fill: func(overrideKubeContext string) (*string, error) {
			kubeContext, err := kubeconfig.GetCurrentContext()
			if err != nil {
				return nil, err
			}
			if overrideKubeContext != "" {
				kubeContext = overrideKubeContext
			}

			cloudConfigData, err := cloudconfig.ParseProviderConfig()
			if err != nil {
				return nil, err
			}

			_, providerName, err := kubeconfig.GetSpaceID(kubeContext)
			if err != nil {
				// use global provider config as fallback
				if cloudConfigData.Default != "" {
					providerName = cloudConfigData.Default
				} else {
					providerName = cloudconfig.DevSpaceCloudProviderName
				}
			}

			provider := cloudconfig.GetProvider(cloudConfigData, providerName)
			if provider == nil {
				return nil, nil
			}
			if provider.Token == "" {
				return nil, nil
			}

			accountName, err := cloudtoken.GetAccountName(provider.Token)
			if err != nil {
				return nil, nil
			}

			return &accountName, nil
		},
	},
}

type predefinedVarDefinition struct {
	Value        *string
	ErrorMessage string
	Fill         func(string) (*string, error)
}

func fillPredefinedVars(overrideKubeContext string) error {
	for varName, predefinedVariable := range PredefinedVars {
		val, err := predefinedVariable.Fill(overrideKubeContext)
		if err != nil {
			return errors.Wrap(err, "fill predefined var "+varName)
		}

		predefinedVariable.Value = val
	}

	return nil
}

func getPredefinedVar(name, overrideKubeContext string) (bool, string, error) {
	if variable, ok := PredefinedVars[strings.ToUpper(name)]; ok {
		if variable.Value == nil {
			return false, "", errors.New(variable.ErrorMessage)
		}

		return true, *variable.Value, nil
	}

	// Load space domain environment variable
	if strings.HasPrefix(strings.ToUpper(name), "DEVSPACE_SPACE_DOMAIN") {
		idx, err := strconv.Atoi(name[len("DEVSPACE_SPACE_DOMAIN"):])
		if err != nil {
			return false, "", errors.Errorf("Error parsing variable %s: %v", name, err)
		}

		kubeContext, err := kubeconfig.GetCurrentContext()
		if err != nil {
			return false, "", errors.Wrap(err, "get current context")
		}
		if overrideKubeContext != "" {
			kubeContext = overrideKubeContext
		}

		spaceID, providerName, err := kubeconfig.GetSpaceID(kubeContext)
		if err != nil {
			return false, "", errors.Errorf("No space configured, but predefined var %s is used.\n\nPlease run: \n- `%s` to create a new space\n- `%s` to use an existing space\n- `%s` to list existing spaces", name, ansi.Color("devspace create space [NAME]", "white+b"), ansi.Color("devspace use space [NAME]", "white+b"), ansi.Color("devspace list spaces", "white+b"))
		}

		cloudConfigData, err := cloudconfig.ParseProviderConfig()
		if err != nil {
			return false, "", errors.Wrap(err, "parse provider config")
		}

		provider := cloudconfig.GetProvider(cloudConfigData, providerName)
		if provider == nil {
			return false, "", errors.Errorf("Couldn't find space provider: %s", providerName)
		}
		if provider.Spaces == nil {
			return false, "", errors.Errorf("No space configured, but predefined var %s is used.\n\nPlease run: \n- `%s` to create a new space\n- `%s` to use an existing space\n- `%s` to list existing spaces", name, ansi.Color("devspace create space [NAME]", "white+b"), ansi.Color("devspace use space [NAME]", "white+b"), ansi.Color("devspace list spaces", "white+b"))
		}
		if provider.Spaces[spaceID] == nil {
			return false, "", errors.Errorf("No space configured, but predefined var %s is used.\n\nPlease run: \n- `%s` to create a new space\n- `%s` to use an existing space\n- `%s` to list existing spaces", name, ansi.Color("devspace create space [NAME]", "white+b"), ansi.Color("devspace use space [NAME]", "white+b"), ansi.Color("devspace list spaces", "white+b"))
		}

		if len(provider.Spaces[spaceID].Space.Domains) <= idx-1 {
			return false, "", errors.Errorf("Error loading %s: Space has %d domains but domain with number %d was requested", name, len(provider.Spaces[spaceID].Space.Domains), idx)
		}

		return true, provider.Spaces[spaceID].Space.Domains[idx-1].URL, nil
	}

	return false, "", nil
}