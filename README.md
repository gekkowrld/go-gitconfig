# Git Config

This is a tool to extract git configurations in go.

It reads the data directly from `git config` files and returns the values.
It works using key value pair, if the key isn't set, then it returns an Error String.
Along with the value of the key, there is a err value, this is an int representation.

Only use the values if the value of `error` code is `nil`.

The configurations are returned in order of their precedence, the first one to be found is returned.

Look at the [git-config FILES section](https://git-scm.com/docs/git-config#FILES) about precidence


