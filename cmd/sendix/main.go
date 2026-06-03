package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/nex1gen/sendix/internal/cli"
)

// version is set at build time via -ldflags; otherwise falls back to module version from debug.BuildInfo.
var version = "dev"

const banner = `[93m ███████╗ █[91m██████╗ ███╗   ██╗ ██████╗ [93m ██╗ ██╗  ██╗
[93m ██╔════╝ █[91m█╔════╝ ████╗  ██║ ██╔══██╗[93m ██║ ╚██╗██╔╝
[93m ███████╗ █[91m████╗   ██╔██╗ ██║ ██║  ██║[93m ██║  ╚███╔╝ 
[93m ╚════██║ █[91m█╔══╝   ██║╚██╗██║ ██║  ██║[93m ██║  ██╔██╗ 
[93m ███████║ █[91m██████╗ ██║ ╚████║ ██████╔╝[93m ██║ ██╔╝ ██╗
[93m ╚══════╝ ╚[91m══════╝ ╚═╝  ╚═══╝ ╚═════╝ [93m ╚═╝ ╚═╝  ╚═╝
    %s
`

func getVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func main() {
	fmt.Fprintf(os.Stderr, banner, getVersion())
	os.Exit(cli.Run(os.Args[1:]))
}
