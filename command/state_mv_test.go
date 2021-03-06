package command

import (
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/cli"
)

func TestStateMv(t *testing.T) {
	state := &terraform.State{
		Modules: []*terraform.ModuleState{
			&terraform.ModuleState{
				Path: []string{"root"},
				Resources: map[string]*terraform.ResourceState{
					"test_instance.foo": &terraform.ResourceState{
						Type: "test_instance",
						Primary: &terraform.InstanceState{
							ID: "bar",
							Attributes: map[string]string{
								"foo": "value",
								"bar": "value",
							},
						},
					},

					"test_instance.baz": &terraform.ResourceState{
						Type: "test_instance",
						Primary: &terraform.InstanceState{
							ID: "foo",
							Attributes: map[string]string{
								"foo": "value",
								"bar": "value",
							},
						},
					},
				},
			},
		},
	}

	statePath := testStateFile(t, state)

	p := testProvider()
	ui := new(cli.MockUi)
	c := &StateMvCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	args := []string{
		"-state", statePath,
		"test_instance.foo",
		"test_instance.bar",
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	// Test it is correct
	testStateOutput(t, statePath, testStateMvOutput)

	// Test we have backups
	backups := testStateBackups(t, filepath.Dir(statePath))
	if len(backups) != 1 {
		t.Fatalf("bad: %#v", backups)
	}
	testStateOutput(t, backups[0], testStateMvOutputOriginal)
}

func TestStateMv_stateOutNew(t *testing.T) {
	state := &terraform.State{
		Modules: []*terraform.ModuleState{
			&terraform.ModuleState{
				Path: []string{"root"},
				Resources: map[string]*terraform.ResourceState{
					"test_instance.foo": &terraform.ResourceState{
						Type: "test_instance",
						Primary: &terraform.InstanceState{
							ID: "bar",
							Attributes: map[string]string{
								"foo": "value",
								"bar": "value",
							},
						},
					},
				},
			},
		},
	}

	statePath := testStateFile(t, state)
	stateOutPath := statePath + ".out"

	p := testProvider()
	ui := new(cli.MockUi)
	c := &StateMvCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	args := []string{
		"-state", statePath,
		"-state-out", stateOutPath,
		"test_instance.foo",
		"test_instance.bar",
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	// Test it is correct
	testStateOutput(t, stateOutPath, testStateMvOutput_stateOut)
	testStateOutput(t, statePath, testStateMvOutput_stateOutSrc)

	// Test we have backups
	backups := testStateBackups(t, filepath.Dir(statePath))
	if len(backups) != 1 {
		t.Fatalf("bad: %#v", backups)
	}
	testStateOutput(t, backups[0], testStateMvOutput_stateOutOriginal)
}

func TestStateMv_stateOutExisting(t *testing.T) {
	stateSrc := &terraform.State{
		Modules: []*terraform.ModuleState{
			&terraform.ModuleState{
				Path: []string{"root"},
				Resources: map[string]*terraform.ResourceState{
					"test_instance.foo": &terraform.ResourceState{
						Type: "test_instance",
						Primary: &terraform.InstanceState{
							ID: "bar",
							Attributes: map[string]string{
								"foo": "value",
								"bar": "value",
							},
						},
					},
				},
			},
		},
	}

	statePath := testStateFile(t, stateSrc)

	stateDst := &terraform.State{
		Modules: []*terraform.ModuleState{
			&terraform.ModuleState{
				Path: []string{"root"},
				Resources: map[string]*terraform.ResourceState{
					"test_instance.qux": &terraform.ResourceState{
						Type: "test_instance",
						Primary: &terraform.InstanceState{
							ID: "bar",
						},
					},
				},
			},
		},
	}

	stateOutPath := testStateFile(t, stateDst)

	p := testProvider()
	ui := new(cli.MockUi)
	c := &StateMvCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	args := []string{
		"-state", statePath,
		"-state-out", stateOutPath,
		"test_instance.foo",
		"test_instance.bar",
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	// Test it is correct
	testStateOutput(t, stateOutPath, testStateMvExisting_stateDst)
	testStateOutput(t, statePath, testStateMvExisting_stateSrc)

	// Test we have backups
	backups := testStateBackups(t, filepath.Dir(statePath))
	if len(backups) != 1 {
		t.Fatalf("bad: %#v", backups)
	}
	testStateOutput(t, backups[0], testStateMvExisting_stateSrcOriginal)

	backups = testStateBackups(t, filepath.Dir(stateOutPath))
	if len(backups) != 1 {
		t.Fatalf("bad: %#v", backups)
	}
	testStateOutput(t, backups[0], testStateMvExisting_stateDstOriginal)
}

func TestStateMv_noState(t *testing.T) {
	tmp, cwd := testCwd(t)
	defer testFixCwd(t, tmp, cwd)

	p := testProvider()
	ui := new(cli.MockUi)
	c := &StateMvCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	args := []string{"from", "to"}
	if code := c.Run(args); code != 1 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}
}

func TestStateMv_stateOutNew_count(t *testing.T) {
	state := &terraform.State{
		Modules: []*terraform.ModuleState{
			&terraform.ModuleState{
				Path: []string{"root"},
				Resources: map[string]*terraform.ResourceState{
					"test_instance.foo.0": &terraform.ResourceState{
						Type: "test_instance",
						Primary: &terraform.InstanceState{
							ID: "foo",
							Attributes: map[string]string{
								"foo": "value",
								"bar": "value",
							},
						},
					},

					"test_instance.foo.1": &terraform.ResourceState{
						Type: "test_instance",
						Primary: &terraform.InstanceState{
							ID: "bar",
							Attributes: map[string]string{
								"foo": "value",
								"bar": "value",
							},
						},
					},

					"test_instance.bar": &terraform.ResourceState{
						Type: "test_instance",
						Primary: &terraform.InstanceState{
							ID: "bar",
							Attributes: map[string]string{
								"foo": "value",
								"bar": "value",
							},
						},
					},
				},
			},
		},
	}

	statePath := testStateFile(t, state)
	stateOutPath := statePath + ".out"

	p := testProvider()
	ui := new(cli.MockUi)
	c := &StateMvCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	args := []string{
		"-state", statePath,
		"-state-out", stateOutPath,
		"test_instance.foo",
		"test_instance.bar",
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	// Test it is correct
	testStateOutput(t, stateOutPath, testStateMvCount_stateOut)
	testStateOutput(t, statePath, testStateMvCount_stateOutSrc)

	// Test we have backups
	backups := testStateBackups(t, filepath.Dir(statePath))
	if len(backups) != 1 {
		t.Fatalf("bad: %#v", backups)
	}
	testStateOutput(t, backups[0], testStateMvCount_stateOutOriginal)
}

func TestStateMv_stateOutNew_nestedModule(t *testing.T) {
	state := &terraform.State{
		Modules: []*terraform.ModuleState{
			&terraform.ModuleState{
				Path:      []string{"root"},
				Resources: map[string]*terraform.ResourceState{},
			},

			&terraform.ModuleState{
				Path:      []string{"root", "foo"},
				Resources: map[string]*terraform.ResourceState{},
			},

			&terraform.ModuleState{
				Path: []string{"root", "foo", "child1"},
				Resources: map[string]*terraform.ResourceState{
					"test_instance.foo": &terraform.ResourceState{
						Type: "test_instance",
						Primary: &terraform.InstanceState{
							ID: "bar",
							Attributes: map[string]string{
								"foo": "value",
								"bar": "value",
							},
						},
					},
				},
			},

			&terraform.ModuleState{
				Path: []string{"root", "foo", "child2"},
				Resources: map[string]*terraform.ResourceState{
					"test_instance.foo": &terraform.ResourceState{
						Type: "test_instance",
						Primary: &terraform.InstanceState{
							ID: "bar",
							Attributes: map[string]string{
								"foo": "value",
								"bar": "value",
							},
						},
					},
				},
			},
		},
	}

	statePath := testStateFile(t, state)
	stateOutPath := statePath + ".out"

	p := testProvider()
	ui := new(cli.MockUi)
	c := &StateMvCommand{
		Meta: Meta{
			ContextOpts: testCtxConfig(p),
			Ui:          ui,
		},
	}

	args := []string{
		"-state", statePath,
		"-state-out", stateOutPath,
		"module.foo",
		"module.bar",
	}
	if code := c.Run(args); code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, ui.ErrorWriter.String())
	}

	// Test it is correct
	testStateOutput(t, stateOutPath, testStateMvNestedModule_stateOut)
	testStateOutput(t, statePath, testStateMvNestedModule_stateOutSrc)

	// Test we have backups
	backups := testStateBackups(t, filepath.Dir(statePath))
	if len(backups) != 1 {
		t.Fatalf("bad: %#v", backups)
	}
	testStateOutput(t, backups[0], testStateMvNestedModule_stateOutOriginal)
}

const testStateMvOutputOriginal = `
test_instance.baz:
  ID = foo
  bar = value
  foo = value
test_instance.foo:
  ID = bar
  bar = value
  foo = value
`

const testStateMvOutput = `
test_instance.bar:
  ID = bar
  bar = value
  foo = value
test_instance.baz:
  ID = foo
  bar = value
  foo = value
`

const testStateMvCount_stateOut = `
test_instance.bar.0:
  ID = foo
  bar = value
  foo = value
test_instance.bar.1:
  ID = bar
  bar = value
  foo = value
`

const testStateMvCount_stateOutSrc = `
test_instance.bar:
  ID = bar
  bar = value
  foo = value
`

const testStateMvCount_stateOutOriginal = `
test_instance.bar:
  ID = bar
  bar = value
  foo = value
test_instance.foo.0:
  ID = foo
  bar = value
  foo = value
test_instance.foo.1:
  ID = bar
  bar = value
  foo = value
`

const testStateMvNestedModule_stateOut = `
<no state>
module.bar:
  <no state>
module.bar.child1:
  test_instance.foo:
    ID = bar
    bar = value
    foo = value
module.bar.child2:
  test_instance.foo:
    ID = bar
    bar = value
    foo = value
`

const testStateMvNestedModule_stateOutSrc = `
<no state>
`

const testStateMvNestedModule_stateOutOriginal = `
<no state>
module.foo:
  <no state>
module.foo.child1:
  test_instance.foo:
    ID = bar
    bar = value
    foo = value
module.foo.child2:
  test_instance.foo:
    ID = bar
    bar = value
    foo = value
`

const testStateMvOutput_stateOut = `
test_instance.bar:
  ID = bar
  bar = value
  foo = value
`

const testStateMvOutput_stateOutSrc = `
<no state>
`

const testStateMvOutput_stateOutOriginal = `
test_instance.foo:
  ID = bar
  bar = value
  foo = value
`

const testStateMvExisting_stateSrc = `
<no state>
`

const testStateMvExisting_stateDst = `
test_instance.bar:
  ID = bar
  bar = value
  foo = value
test_instance.qux:
  ID = bar
`

const testStateMvExisting_stateSrcOriginal = `
test_instance.foo:
  ID = bar
  bar = value
  foo = value
`

const testStateMvExisting_stateDstOriginal = `
test_instance.qux:
  ID = bar
`
