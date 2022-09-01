## Instructions

1. Copy `env.secret.tpl` to a new file called `env.secret`. Edit the new file to include the required variables. This file should be sourced before running the `update-issues` script.

2. Modify `index.tpl.html` to include desired formatting

3. Build the `update-issues` script
```
$> go build -o update-issues{,.go}
```

4. Run the `update-issues` script. As an example, for the repo at https://github.com/eugeneswalker/github-issue-tracker, use `eugeneswalker` for the namespace and `github-issue-tracker` for the repo name.
```
$> . ./env.secret
$> ./update-issues -n <repo-namespace> -r <repo-name>
```

5. Examine generated `index.html`
