package cli

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// row is one line of a help list: what to type, and what it does.
type row struct {
	left, right string
}

// guideText returns the guide that sliceflow --help shows, built from
// registry: the intro, the usage line, one line for each group in
// alphabetical order, and the help option.
func guideText(registry CommandRegistry) string {
	var rows []row
	for _, name := range slices.Sorted(maps.Keys(registry.Groups)) {
		rows = append(rows, row{name, registry.Groups[name].Description})
	}

	var b strings.Builder
	b.WriteString(intro)
	b.WriteString("\n\n")
	b.WriteString("Usage:\n  sliceflow <command> [options]\n\n")
	b.WriteString("Commands:\n")
	writeRows(&b, rows)
	b.WriteString("\nOptions:\n")
	writeRows(&b, []row{helpOption})
	return b.String()
}

// groupHelpText returns the help of the group name, e.g. what
// sliceflow slice --help shows: the group's description, its usage line,
// one line for each of its commands in alphabetical order, with their
// arguments, and the help option.
func groupHelpText(name string, group CommandGroup) string {
	var rows []row
	for _, actionName := range slices.Sorted(maps.Keys(group.Actions)) {
		action := group.Actions[actionName]
		rows = append(rows, row{actionName + argsUsage(action.Args), action.Description})
	}

	var b strings.Builder
	b.WriteString(group.Description)
	b.WriteString(".\n\n")
	fmt.Fprintf(&b, "Usage:\n  sliceflow %s <command> [options]\n\n", name)
	b.WriteString("Commands:\n")
	writeRows(&b, rows)
	b.WriteString("\nOptions:\n")
	writeRows(&b, []row{helpOption})
	return b.String()
}

// argsUsage returns the arguments of a command as they appear in a usage
// line, each after a space, e.g. " <name>". It returns "" for a command
// with no arguments.
func argsUsage(args []CommandArg) string {
	var b strings.Builder
	for _, arg := range args {
		b.WriteString(" ")
		b.WriteString(argUsage(arg))
	}
	return b.String()
}

// argUsage returns one argument as it appears in help: <name> when it is
// required, [name] when it is not.
func argUsage(arg CommandArg) string {
	if arg.Required {
		return "<" + arg.Name + ">"
	}
	return "[" + arg.Name + "]"
}

// writeRows writes each row on its own line, indented by two spaces,
// with the right parts lined up two spaces after the longest left part.
func writeRows(b *strings.Builder, rows []row) {
	width := 0
	for _, r := range rows {
		width = max(width, len(r.left))
	}
	for _, r := range rows {
		fmt.Fprintf(b, "  %-*s  %s\n", width, r.left, r.right)
	}
}

// commandHelpText returns the help of one command, e.g. what
// sliceflow slice create --help shows: its long description, its usage
// line, its arguments in their order on the command line, when it has
// any, and its options in alphabetical order, with -h, --help last.
func commandHelpText(groupName, actionName string, action CommandAction) string {
	var rows []row
	for _, name := range slices.Sorted(maps.Keys(action.Options)) {
		rows = append(rows, row{optionUsage(name, action.Options[name]), action.Options[name].Description})
	}
	rows = append(rows, helpOption)

	var b strings.Builder
	b.WriteString(action.LongDescription)
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "Usage:\n  %s\n\n", usageLine(groupName, actionName, action))
	if len(action.Args) > 0 {
		b.WriteString("Arguments:\n")
		writeRows(&b, argRows(action.Args))
		b.WriteString("\n")
	}
	b.WriteString("Options:\n")
	writeRows(&b, rows)
	return b.String()
}

// optionUsage returns an option as it appears in help: --name, then
// <value> when the option takes a value, e.g. --title <text>.
func optionUsage(name string, option CommandOption) string {
	if option.ValueName == "" {
		return "--" + name
	}
	return "--" + name + " <" + option.ValueName + ">"
}

// argRows returns one help row for each argument, in their order on the
// command line: <name> or [name], and its description.
func argRows(args []CommandArg) []row {
	var rows []row
	for _, arg := range args {
		rows = append(rows, row{argUsage(arg), arg.Description})
	}
	return rows
}

// usageLine returns how a command is typed, e.g.
// "sliceflow slice create <name> [options]". Help and some errors show it.
func usageLine(groupName, actionName string, action CommandAction) string {
	return "sliceflow " + groupName + " " + actionName + argsUsage(action.Args) + " [options]"
}
