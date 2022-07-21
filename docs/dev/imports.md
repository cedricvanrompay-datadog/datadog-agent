# Goimports

In the `imports ( .. )` section of each Go file, imports should be separated into at least three sections:

1. standard library packages
1. external packages
1. local packages

This is verified both by a pre-commit hook and in CI.

## Editors

The `goimports` tool can do this for you automatically, with the `-local github.com/DataDog/datadog-agent` flag.

(Please feel free to add instructions for your favorite editor here!)

### Vim

In vim, using `vim-go`, add

```vim
let g:go_fmt_options = {
\ 'goimports': '-local github.com/DataDog/datadog-agent',
\ }
```

### VSCode

```json
{
  "gopls": {
    "formatting.local": "github.com/DataDog/datadog-agent"
  } 
}
```

See https://github.com/golang/vscode-go/wiki/features#format-and-organize-imports.
