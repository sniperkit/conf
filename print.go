package conf

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"
)

// PrintError outputs the error message for err to stderr.
func (ld Loader) PrintError(err error) {
	w := bufio.NewWriter(os.Stderr)
	ld.FprintError(w, err)
	w.Flush()
}

// FprintError outputs the error message for err to w.
func (ld Loader) FprintError(w io.Writer, err error) {
	fmt.Fprintf(w, "Error:\n  %v\n", err)
}

// PrintHelp outputs the help message for cfg to stderr.
func (ld Loader) PrintHelp(cfg interface{}) {
	w := bufio.NewWriter(os.Stderr)
	ld.FprintHelp(w, cfg)
	w.Flush()
}

// FprintHelp outputs the help message for cfg to w.
func (ld Loader) FprintHelp(w io.Writer, cfg interface{}) {
	v := reflect.ValueOf(cfg)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		panic(fmt.Sprintf("cannot load configuration into %T", cfg))
	}

	set := newFlagSet(makeConfValue(v), ld.Program)

	if len(ld.FileFlag) != 0 {
		addFileFlag(set, nil, ld.FileFlag)
	}

	fmt.Fprintf(w, "Usage of %s:\n", ld.Program)

	// Outputs the flags following the same format than the standard flag
	// package. The main difference is in the type names which are set to
	// values returned by prettyType.
	set.VisitAll(func(f *flag.Flag) {
		v := f.Value.(value)
		h := []string{}

		fmt.Fprintf(w, "  -%s", f.Name)

		switch {
		case !v.IsBoolFlag():
			fmt.Fprintf(w, " %s\n", prettyType(v.v.Type()))
		case len(f.Name) > 4: // put help message inline for boolean flags
			fmt.Fprint(w, "\n")
		}

		if s := f.Usage; len(s) != 0 {
			h = append(h, s)
		}
		if s := f.DefValue; len(s) != 0 && !v.IsBoolFlag() && !isZeroValue(v.v) {
			h = append(h, "(default "+s+")")
		}

		if len(h) != 0 {
			if !v.IsBoolFlag() || len(f.Name) > 4 {
				fmt.Fprint(w, "    ")
			}
			fmt.Fprintf(w, "\t%s\n", strings.Join(h, " "))
		}
	})
}

func prettyType(t reflect.Type) string {
	if t == nil {
		return "unknown"
	}

	switch {
	case t == reflect.TypeOf(time.Duration(0)):
		return "duration"

	case t == reflect.TypeOf(duration(0)):
		return "duration"

	case t == reflect.TypeOf(time.Time{}):
		return "time"
	}

	switch t.Kind() {
	case reflect.Struct, reflect.Map:
		return "object"

	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			return "base64"
		}
		return "list"

	case reflect.Ptr:
		return prettyType(t.Elem())

	default:
		return t.String()
	}
}