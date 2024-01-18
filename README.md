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

username, err := gogitconfig.GetValue("user.name", "/path/to/repo") // For a specific starting point
if err != nil {
    fmt.Print(err)
}

email, err := gogitconfig.GetValue("user.email") // For when no specific location is required
if err != nil {
    fmt.Print(err)
}

fmt.Printf("Your git username is %s\nYour git Email is %s", username, email)
```
