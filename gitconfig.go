package gogitconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"gopkg.in/ini.v1"
)

var LookupStartLocation string

// BUG(gekkowrld): Can't handle windows paths yet
func gogitconfig(configKey string, startingPoint ...string) (string, error) {
	// Instantiate the configLocation
	configLocation, _ := os.Getwd()

	// Overide it if there is the startingPoint is actually set
	if len(startingPoint) >= 1 {
		// Take only the fist part of the input incase the user passed excess
		configLocation = startingPoint[0]
	}

	LookupStartLocation = configLocation
	return callConfRecursively(configKey)
}

/* The following methods are not exported, for internal usage */

// configLevel is used to return the file(s) to be used in each level.
// If it fails to get the file at a specified level, it returns an empty string
// No error is returned, consider an empty sting an error.
// BUG(gekkowrld): Can't get system configuration, I have to figure out how to get $(prefix)
func configLevelFile(level string) string {
	configStartingLocation := LookupStartLocation

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
