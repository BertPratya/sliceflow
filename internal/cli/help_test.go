package cli

import "testing"

func TestGroupHelpListsEveryCommand(t *testing.T) {
	g := group{
		path:  "pet",
		about: "pet looks after your pet.",
		commands: []command{
			{name: "feed", args: "<food>", about: "Feed the pet"},
			{name: "walk", about: "Walk the pet"},
		},
	}

	want := `pet looks after your pet.

Usage:
  pet <command> [arguments]

Commands:
  feed <food>    Feed the pet
  walk           Walk the pet

Flags:
  -h, --help    Show this help

Run "pet <command> --help" for more about a command.
`
	if got := g.help(); got != want {
		t.Errorf("help() =\n%s\nwant:\n%s", got, want)
	}
}
