// A package to retrive git configuration values.
// Doesn't use git commands, instead reads the files directly.
// You are free to audit and use the code as you please, just follow the LICENSE agreement when doing that.
// NOTE: Don't fully trust the results returned, do your own error checking.

package gogitconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"gopkg.in/ini.v1"
)

// Options Passed
// Set:
//     - the startiing location
//     - the config key e.g user.key
//     - the config level
// The config level is one of this values:
//   - 1 -> "local"
//   - 2 -> "global"
//   - 3 -> "system"
//   - not set -> whichever comes first
//   - 0 and any other number -> whichever comes first
type OptionsPassed struct {
    LookupStartLocation string
    ConfigLevel         int
    ConfigKey           string
}

// The var for use as "normal"
var optsEntered struct {
    LookupStartLocation string
    ConfigLevel         int
    ConfigKey           string
}

// BUG(gekkowrld): Can't handle windows paths yet.

// GetValue returns the value of the git keys as specified by git config files.
// If the key is not set, an error is returned with an empty string
// The startingPoint is optional and should only be set if there is need for constant starting point.
// The value is searched from the local configuration file upwards.
// No error is returned if any of the configuration files don't exist or have an error.
// An error is only returned when the key is not found.
func GetValue(optsPassed OptionsPassed) (string, error) {
    // Instantiate the configLocation
    configLocation, _ := os.Getwd()

    // Override it if there is the startingPoint is actually set
    if optsPassed.LookupStartLocation != ""{
        // Take only the first part of the input in case the user passed excess
        configLocation = optsPassed.LookupStartLocation
    }

    // Check if a user passed in a file or a directory.
    // If not directory, return the parent directory
    // Don't do any error checking for now.
    isUserDir := directoryExists(configLocation)
    if !isUserDir {
        configLocation = filepath.Dir(configLocation)
    }

    // BUG: Not my bug but worth mentioning:
    // The filepath.Dir() doesn't play nice with Windows.
    // So there may be some problems with that if it happens
    // to run on windows.

    optsEntered.LookupStartLocation = configLocation
    optsEntered.ConfigLevel = optsPassed.ConfigLevel
    optsEntered.ConfigKey = optsPassed.ConfigKey
    return dealWithLevels()
}

// Check the user entered number and use it to
// call the relevant function(s)
func dealWithLevels() (string, error){
    var level int
    var confKey string
    level = optsEntered.ConfigLevel
    confKey = optsEntered.ConfigKey

    if level == 1 {
	    local_file := configLevelFile("local")
    	return  getConfig(confKey, local_file)
    } else if level == 2 {
	    global_file := configLevelFile("global")
    	return  getConfig(confKey, global_file)
    } else if level == 3 {
	    system_file := configLevelFile("system")
    	return getConfig(confKey, system_file)
    } else {
        return callConfRecursively(confKey)
    }
}

/* The following methods are not exported, for internal usage */

// configLevel is used to return the file(s) to be used in each level.
// If it fails to get the file at a specified level, it returns an empty string
// No error is returned, consider an empty sting an error.
// BUG(gekkowrld): Can't get system configuration, I have to figure out how to get $(prefix)
func configLevelFile(level string) string {
	configStartingLocation := optsEntered.LookupStartLocation

	// For files to be used/returned
	var localLocation string
	var globalLocation string
	var systemLocation string

	switch level {
	case "local":
		// For local level, there is only one file, $GITDIR/.git/config file
		gitroot, err := findGitRoot(configStartingLocation)
		if err != nil {
			return fmt.Sprintf("%v", err)
		}

		localLocation = filepath.Join(gitroot, ".git/config")
		if !fileExists(localLocation) {
			localLocation = ""
		}
		return localLocation
	case "global":
		// For global, there are two possible locations:
		// $HOME/.gitconfig or $XDG_CONFIG_HOME/git/config
		xdg_file, _ := xdg.ConfigFile("git/config")
		home_file := filepath.Join(os.Getenv("HOME"), ".gitconfig")
		if fileExists(xdg_file) {
			globalLocation = xdg_file
		} else if fileExists(home_file) {
			globalLocation = home_file
		}

		return globalLocation
	case "system":
		return systemLocation
	default:
		return ""
	}
}

// directoryExists checks if a directory exists, it must be a directory
// for it to return true.
func directoryExists(dirName string) bool {
	info, err := os.Stat(dirName)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// fileExists checks if a file exists, if not return false
func fileExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}


// findGitRoot recursively works its way up from the location specified.
// If it reaches the root directory before it gets a hit, it returns an error.
func findGitRoot(currentPath string) (string, error) {
	currentPath, err := filepath.Abs(currentPath)
	if err != nil {
		return "", err
	}

	// Loop over from the starting point to root, if there is a location where
	//   there is a match e.g $LOCATION/.git, then that is a git directory,
	//   return the location in absolute path form.
	for {
		gitPath := filepath.Join(currentPath, ".git")
		if directoryExists(gitPath) {
			return currentPath, nil
		}

		// Move up one level in the directory hierarchy
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached the root without finding .git directory
			break
		}
		currentPath = parentPath
	}

	return "", fmt.Errorf("fatal: not a git repository or any parent up to mount point")
}

// getConfig takes in a key that can be in the form [section].[key] format.
// It then splits it and gives it to an ini processor.
// The value is returned "AS IS", no further modification is done with the value
// Be sure to check the error value and the value returned (e.g for empty string)
func getConfig(gitkey string, filename string) (string, error) {
	splitString := strings.SplitN(gitkey, ".", 2)
	var parent string
	var sibling string

	if len(splitString) == 2 {
		parent = splitString[0]
		sibling = splitString[1]
	} else if len(splitString) == 1 {
		sibling = splitString[0]
	}

	cfg, err := ini.Load(filename)
	if err != nil {
		return "", err
	}

	return cfg.Section(parent).Key(sibling).String(), nil
}

// callConfRecursively just calls getConfig on all three known git config levels
func callConfRecursively(confKey string) (string, error) {
	// Call from local, then global and finally system.

	local_file := configLevelFile("local")
	tmpVal, _ := getConfig(confKey, local_file)
	if tmpVal != "" {
		return tmpVal, nil
	}
	global_file := configLevelFile("global")
	tmpVal, _ = getConfig(confKey, global_file)
	if tmpVal != "" {
		return tmpVal, nil
	}

	sytem_file := configLevelFile("system")
	tmpVal, _ = getConfig(confKey, sytem_file)
	if tmpVal != "" {
		return tmpVal, nil
	}

	return "", fmt.Errorf("Couldn't get the value of the requested key")
}
