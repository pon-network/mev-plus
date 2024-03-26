package config

import (
	"encoding"
	"flag"
	"os"
	"os/user"
	"strings"

	"github.com/urfave/cli/v2"
)

type TextMarshaler interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

// textMarshalerVal turns a TextMarshaler into a flag.Value
type textMarshalerVal struct {
	v TextMarshaler
}

func (v textMarshalerVal) String() string {
	if v.v == nil {
		return ""
	}
	text, _ := v.v.MarshalText()
	return string(text)
}

func (v textMarshalerVal) Set(s string) error {
	return v.v.UnmarshalText([]byte(s))
}

var (
	_ cli.Flag              = (*TextMarshalerFlag)(nil)
	_ cli.RequiredFlag      = (*TextMarshalerFlag)(nil)
	_ cli.VisibleFlag       = (*TextMarshalerFlag)(nil)
	_ cli.DocGenerationFlag = (*TextMarshalerFlag)(nil)
	_ cli.CategorizableFlag = (*TextMarshalerFlag)(nil)
)

// TextMarshalerFlag wraps a TextMarshaler value.
type TextMarshalerFlag struct {
	Name string

	Category    string
	DefaultText string
	Usage       string

	Required   bool
	Hidden     bool
	HasBeenSet bool

	Value TextMarshaler

	Aliases []string
}

// For cli.Flag:

func (f *TextMarshalerFlag) Names() []string { return append([]string{f.Name}, f.Aliases...) }
func (f *TextMarshalerFlag) IsSet() bool     { return f.HasBeenSet }
func (f *TextMarshalerFlag) String() string  { return cli.FlagStringer(f) }

func (f *TextMarshalerFlag) Apply(set *flag.FlagSet) error {
	eachName(f, func(name string) {
		set.Var(textMarshalerVal{f.Value}, f.Name, f.Usage)
	})
	return nil
}

// For cli.RequiredFlag:

func (f *TextMarshalerFlag) IsRequired() bool { return f.Required }

// For cli.VisibleFlag:

func (f *TextMarshalerFlag) IsVisible() bool { return !f.Hidden }

// For cli.CategorizableFlag:

func (f *TextMarshalerFlag) GetCategory() string { return f.Category }

// For cli.DocGenerationFlag:

func (f *TextMarshalerFlag) TakesValue() bool     { return true }
func (f *TextMarshalerFlag) GetUsage() string     { return f.Usage }
func (f *TextMarshalerFlag) GetEnvVars() []string { return nil } // env not supported

func (f *TextMarshalerFlag) GetValue() string {
	t, err := f.Value.MarshalText()
	if err != nil {
		return "(ERR: " + err.Error() + ")"
	}
	return string(t)
}

func (f *TextMarshalerFlag) GetDefaultText() string {
	if f.DefaultText != "" {
		return f.DefaultText
	}
	return f.GetValue()
}

func HomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

func eachName(f cli.Flag, fn func(string)) {
	for _, name := range f.Names() {
		name = strings.Trim(name, " ")
		fn(name)
	}
}
