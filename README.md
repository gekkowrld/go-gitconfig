# Git Config

This is a tool to extract git configurations in go.

It reads the data directly from `git config` files and returns the values.
It works using key value pair, if the key isn't set, then it returns an Error String.
The value of the key is set to empty and an error message set accordingly.

Only use the values if the value of `error` code is `nil`.

The configurations are returned in order of their precedence, the first one to be found is returned.

Look at the [git-config FILES section](https://git-scm.com/docs/git-config#FILES) about precidence

## Usage

You should pass in the key of the git value you want to obtain.
This follows the git config model:

```ini
[user]
    name = Your Name
    email = your@email.com
; Others
```

To get the username, you should pass in "user.name" and so on.
The starting point (where you'd like the tool to start git config lookup) can be left blank if the starting point is the current directory or you don't require local config.

Example:

```go
import "github.com/gekkowrld/go-gitconfig"

// The struct to be passed.
// The required field is ConfigKey
// all the others (well 2) are optional
myConfig := OptionsPassed{
    LookupStartLocation: "/path/to/repo", // Adjust this path as needed
    ConfigLevel:         0, 
    ConfigKey:           "user.name",
}

// Call the function with the test case
username, err := gogitconfig.GetValue(myConfig)

fmt.Printf("Your git username is %s\n", username)
```
