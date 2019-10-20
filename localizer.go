package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vc2402/localizer/engine"
)

func main() {

	fs := flag.NewFlagSet("main", flag.ExitOnError)
	fs.Usage = func() {
		fs.Output().Write([]byte(fmt.Sprintf("Usage: %s -export|-import [other-flags] androidProjectPath\n", filepath.Base(os.Args[0]))))
		fs.PrintDefaults()
	}
	expF := fs.String("export", "", "`path` to csv-file to export values to")
	impF := fs.String("import", "", "`path` to csv-file to import values from")
	// locales := fs.String("locales", "", "coma-separated names of required locales (may be defined automatically)")
	fs.Parse(os.Args[1:])

	if fs.NArg() == 0 {
		fs.Usage()
		return
	}
	ap := fs.Arg(0)
	eng := engine.New(ap).Load()

	var err error
	if *expF != "" {
		err = eng.Export(*expF)
	} else if *impF != "" {
		eng.Import(*impF)
		err = eng.Save()
	} else {
		fs.Usage()
	}
	if err != nil {
		fs.Output().Write([]byte(fmt.Sprintln(err)))
		return
	}
}
