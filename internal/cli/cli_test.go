package cli

import "testing"

func TestCommandsExist(t *testing.T) {
	cmds := map[string]*struct{ use string }{
		"init":     {InitCmd.Use},
		"add":      {AddCmd.Use},
		"install":  {InstallCmd.Use},
		"resolve":  {ResolveCmd.Use},
		"show":     {ShowCmd.Use},
		"graph":    {GraphCmd.Use},
		"remove":   {RemoveCmd.Use},
		"validate": {ValidateCmd.Use},
	}
	for name, c := range cmds {
		if c.use == "" {
			t.Errorf("command %s has empty Use", name)
		}
	}
}

func TestAddFlags(t *testing.T) {
	flags := []string{"version", "path", "reason"}
	for _, name := range flags {
		if AddCmd.Flags().Lookup(name) == nil {
			t.Errorf("add command missing flag --%s", name)
		}
	}
}

func TestResolveFlags(t *testing.T) {
	if ResolveCmd.Flags().Lookup("dry-run") == nil {
		t.Error("resolve command missing --dry-run flag")
	}
}

func TestAddRequiresArg(t *testing.T) {
	if AddCmd.Args == nil {
		t.Error("add command has no arg validation")
	}
}

func TestRemoveRequiresArg(t *testing.T) {
	if RemoveCmd.Args == nil {
		t.Error("remove command has no arg validation")
	}
}

func TestShowRequiresArg(t *testing.T) {
	if ShowCmd.Args == nil {
		t.Error("show command has no arg validation")
	}
}
