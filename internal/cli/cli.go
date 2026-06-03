package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nex1gen/sendix/internal/packer"
	"github.com/nex1gen/sendix/internal/password"
	"github.com/nex1gen/sendix/internal/unpacker"
)

// Run parses arguments and dispatches to pack or unpack.
func Run(args []string) int {
	if len(args) < 1 {
		printUsage()
		return 0
	}

	switch args[0] {
	case "pack":
		return runPack(args[1:])
	case "unpack":
		return runUnpack(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
		printUsage()
		return 2
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Description: Encrypted Git repository packaging tool")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Usage: sendix <command> [options]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  pack   [-o <name>.sdx] [--all-branches] [directory]")
	fmt.Fprintln(os.Stderr, "  unpack [-d destination] archive.sdx")
}

func runPack(args []string) int {
	fs := flag.NewFlagSet("pack", flag.ContinueOnError)
	output := fs.String("o", "", "Output archive filename (default: <dirname>_YYYY-MM-DD_HH-MM-SS.sdx)")
	allBranches := fs.Bool("all-branches", false, "Bundle all branches instead of just the current one")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	dir := fs.Arg(0)
	if dir == "" {
		dir = "."
	}

	out := *output
	if out == "" {
		dirname := filepath.Base(filepath.Clean(dir))
		if dirname == "." {
			wd, _ := os.Getwd()
			dirname = filepath.Base(wd)
		}
		out = fmt.Sprintf("%s_%s.sdx", dirname, time.Now().Format("2006-01-02_15-04-05"))
	}

	pwd, err := password.ReadConfirmed()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 3
	}

	if err := packer.Pack(dir, out, pwd, *allBranches); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "Packed to %s\n", out)
	return 0
}

func runUnpack(args []string) int {
	fs := flag.NewFlagSet("unpack", flag.ContinueOnError)
	dest := fs.String("d", ".", "Destination directory")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: archive path required")
		return 2
	}
	archive := fs.Arg(0)

	pwd, err := password.Read("Enter password: ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 3
	}

	if err := unpacker.Unpack(archive, *dest, pwd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "Unpacked to %s\n", *dest)
	return 0
}
