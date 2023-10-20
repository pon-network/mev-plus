package coreCli

import (
	"fmt"
	"strings"
	"github.com/urfave/cli/v2"
)

func init() {
	cli.FlagStringer = FlagString
}

// FlagString prints a single flag in help.
func FlagString(f cli.Flag) string {
	df, ok := f.(cli.DocGenerationFlag)
	if !ok {
		return ""
	}

	needsPlaceholder := df.TakesValue()
	placeholder := ""
	if needsPlaceholder {
		placeholder = "value"
	}

	namesText := pad(cli.FlagNamePrefixer(df.Names(), placeholder), 30)

	defaultValueString := ""
	if s := df.GetDefaultText(); s != "" {
		defaultValueString = " (default: " + s + ")"
	}

	usage := strings.TrimSpace(df.GetUsage())
	envHint := strings.TrimSpace(cli.FlagEnvHinter(df.GetEnvVars(), ""))
	if len(envHint) > 0 {
		usage += " " + envHint
	}

	usage = wordWrap(usage, 80)
	usage = indent(usage, 10)

	return fmt.Sprintf("\n    %s%s\n%s", namesText, defaultValueString, usage)
}

func pad(s string, length int) string {
	if len(s) < length {
		s += strings.Repeat(" ", length-len(s))
	}
	return s
}

func indent(s string, nspace int) string {
	ind := strings.Repeat(" ", nspace)
	return ind + strings.ReplaceAll(s, "\n", "\n"+ind)
}

func wordWrap(s string, width int) string {
	var (
		output     strings.Builder
		lineLength = 0
	)

	for {
		sp := strings.IndexByte(s, ' ')
		var word string
		if sp == -1 {
			word = s
		} else {
			word = s[:sp]
		}
		wlen := len(word)
		over := lineLength+wlen >= width
		if over {
			output.WriteByte('\n')
			lineLength = 0
		} else {
			if lineLength != 0 {
				output.WriteByte(' ')
				lineLength++
			}
		}

		output.WriteString(word)
		lineLength += wlen

		if sp == -1 {
			break
		}
		s = s[wlen+1:]
	}

	return output.String()
}

// Merge merges the given flag slices.
func Merge(groups ...[]cli.Flag) []cli.Flag {
	var ret []cli.Flag
	for _, group := range groups {
		ret = append(ret, group...)
	}
	return ret
}