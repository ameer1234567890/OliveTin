package executor

import (
	"github.com/stretchr/testify/assert"
	"testing"

	acl "github.com/OliveTin/OliveTin/internal/acl"
	config "github.com/OliveTin/OliveTin/internal/config"
)

func TestSanitizeUnsafe(t *testing.T) {
	assert.Nil(t, TypeSafetyCheck("", "_zomg_ c:/ haxxor ' bobby tables && rm -rf ", "very_dangerous_raw_string"))
}

func testingExecutor() (*Executor, *config.Config) {
	e := DefaultExecutor()

	cfg := config.DefaultConfig()

	a1 := config.Action{
		Title: "Do some tickles",
		Shell: "echo 'Tickling {{ person }}'",
		Arguments: []config.ActionArgument{
			{
				Name: "person",
				Type: "ascii",
			},
		},
	}

	cfg.Actions = append(cfg.Actions, a1)
	cfg.Sanitize()

	return e, cfg
}

func TestCreateExecutorAndExec(t *testing.T) {
	e, cfg := testingExecutor()

	req := ExecutionRequest{
		ActionName:        "Do some tickles",
		AuthenticatedUser: &acl.AuthenticatedUser{Username: "Mr Tickle"},
		Cfg:               cfg,
		Arguments: map[string]string{
			"person": "yourself",
		},
	}

	e.ExecRequest(&req)

	assert.NotNil(t, e, "Create an executor")

	assert.NotNil(t, e.ExecRequest(&req), "Execute a request")
	assert.Equal(t, int32(0), req.logEntry.ExitCode, "Exit code is zero")
}

func TestExecNonExistant(t *testing.T) {
	e, cfg := testingExecutor()

	req := ExecutionRequest{
		ActionName: "Waffles",
		logEntry:   &InternalLogEntry{},
		Cfg:        cfg,
	}

	e.ExecRequest(&req)

	assert.Equal(t, int32(-1337), req.logEntry.ExitCode, "Log entry is set to an internal error code")
	assert.Equal(t, "", req.logEntry.ActionIcon, "Log entry icon wasnt found")
}

func TestArgumentNameCamelCase(t *testing.T) {
	a1 := config.Action{
		Title: "Do some tickles",
		Shell: "echo 'Tickling {{ personName }}'",
		Arguments: []config.ActionArgument{
			{
				Name: "personName",
				Type: "ascii",
			},
		},
	}

	values := map[string]string{
		"personName": "Fred",
	}

	out, err := parseActionArguments(a1.Shell, values, &a1)

	assert.Equal(t, "echo 'Tickling Fred'", out)
	assert.Nil(t, err)
}

func TestArgumentNameSnakeCase(t *testing.T) {
	a1 := config.Action{
		Title: "Do some tickles",
		Shell: "echo 'Tickling {{ person_name }}'",
		Arguments: []config.ActionArgument{
			{
				Name: "person_name",
				Type: "ascii",
			},
		},
	}

	values := map[string]string{
		"person_name": "Fred",
	}

	out, err := parseActionArguments(a1.Shell, values, &a1)

	assert.Equal(t, "echo 'Tickling Fred'", out)
	assert.Nil(t, err)
}

func TestArgumentNameNumbers(t *testing.T) {
	a1 := config.Action{
		Title: "Do some tickles",
		Shell: "echo 'Tickling {{ person1name }}'",
		Arguments: []config.ActionArgument{
			{
				Name: "person1name",
				Type: "ascii",
			},
		},
	}

	values := map[string]string{
		"person1name": "Fred",
	}

	out, err := parseActionArguments(a1.Shell, values, &a1)

	assert.Equal(t, "echo 'Tickling Fred'", out)
	assert.Nil(t, err)
}

func TestArgumentNotProvided(t *testing.T) {
	a1 := config.Action{
		Title: "Do some tickles",
		Shell: "echo 'Tickling {{ personName }}'",
		Arguments: []config.ActionArgument{
			{
				Name: "person",
				Type: "ascii",
			},
		},
	}

	values := map[string]string{}

	out, err := parseActionArguments(a1.Shell, values, &a1)

	assert.Equal(t, "", out)
	assert.Equal(t, err.Error(), "Required arg not provided: personName")
}
