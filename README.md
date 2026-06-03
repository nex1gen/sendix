# sendix

Encrypted Git repository packaging tool. Pack a repo into a single `.sdx` archive, transfer it safely, and unpack it on another machine while preserving Git history, working tree files, untracked files, and local Git configuration.

## Install

### Using `go install`

```bash
go install github.com/nex1gen/sendix/cmd/sendix@latest
```

Make sure `$GOPATH/bin` (or `$HOME/go/bin`) is in your `$PATH`.

### Build from source

```bash
git clone https://github.com/nex1gen/sendix.git
cd sendix
go build -ldflags "-X main.version=$(git describe --tags --always)" -o sendix ./cmd/sendix
mv sendix /usr/local/bin/ # or any directory in your $PATH
```

## Usage

```bash
# Pack the current directory (default filename: <dirname>_YYYY-MM-DD_HH-MM-SS.sdx)
sendix pack .

# Pack with a custom output name
sendix pack -o myrepo.sdx /path/to/repo

# Pack all branches (default: current branch only)
sendix pack --all-branches -o myrepo.sdx /path/to/repo

# Unpack into the current directory
sendix unpack myrepo.sdx

# Unpack into a specific directory
sendix unpack -d ./restored myrepo.sdx
```

### Password input

By default `sendix` prompts for the password interactively with hidden input.

For automation or CI/CD you can set the environment variable:

```bash
export SENDIX_PASSWORD="your-secret-password"
sendix pack -o myrepo.sdx .
sendix unpack -d ./restored myrepo.sdx
```

## How it works

`sendix` creates a hybrid encrypted archive:

- **`repo.bundle`** — a [Git bundle](https://git-scm.com/docs/git-bundle) containing the repository history
- **`workspace.tar.gz`** — the working tree, `.git/config`, hooks, and untracked files
- **AES-256-GCM encryption** — archive is encrypted with a key derived from your password via PBKDF2

The final `.sdx` file format:

```
[SALT (16 bytes)] [NONCE (12 bytes)] [AES-256-GCM ciphertext]
```

## Archive contents

The workspace part includes:

- All tracked files
- All untracked files
- `.git/config`
- `.git/hooks/*`
- `.git/info/*`
- Local branches (stored in the bundle)

Excluded from the workspace (handled by the bundle or ignored):

- `.git/objects/*`
- `.git/logs/*`
- Existing `.sdx` archives in the repo directory

## Safety features

- **Self-bloat protection** — any existing `.sdx` files inside the repo are automatically excluded from the archive
- **Safe unpack** — works in existing non-empty directories using `git init` + `git fetch` + `git checkout`
- **Authenticated encryption** — AES-256-GCM with random salt and nonce per archive

## Requirements

- Go 1.25+
- Git installed on both pack and unpack machines

## License

MIT
