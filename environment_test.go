package main

import (
	"fmt"
	"os"
	"testing"
)

func TestEnvErrAssertions(t *testing.T) {
	var _ error = &duplicateValueError{}
	var _ error = &envValueRequired{}
}

func TestEnvironmentAnyValue(t *testing.T) {
	cases := []struct {
		description string
		envvars     map[string]string
		err         error
		expected    string
	}{
		{
			description: "env var value set in multiple places",
			envvars: map[string]string{
				"INPUT_ABC": "abc",
				"ABC":       "abc",
			},
			err: &duplicateValueError{},
		},
		{
			description: "correct fetch",
			envvars: map[string]string{
				"NAME": "Peter",
			},
			expected: "Peter",
		},
		{
			description: "not found value",
			envvars:     nil,
			err:         &envValueRequired{},
		},
	}

	for _, v := range cases {
		t.Run(v.description, func(tt *testing.T) {
			setEnvVars(v.envvars)
			defer unsetEnvVars(v.envvars)

			value, err := getAnyEnvironment(v.description, getKeys(v.envvars)...)
			if err != nil {
				if fmt.Sprintf("%T", err) != fmt.Sprintf("%T", v.err) {
					tt.Fatalf("expecting error to be of type %T, but got %T: %s", v.err, err, err.Error())
				}
			}

			if err == nil && v.err != nil {
				tt.Fatalf("expecting function to fail with error of type %T but got no error", v.err)
			}

			if value != v.expected {
				tt.Fatalf("expecting value to be %q, but got %q", v.expected, value)
			}
		})
	}
}

func setEnvVars(m map[string]string) {
	for k, v := range m {
		os.Setenv(k, v)
	}
}

func unsetEnvVars(m map[string]string) {
	for k := range m {
		os.Unsetenv(k)
	}
}

func getKeys(m map[string]string) []string {
	s := make([]string, 0, len(m))

	for k := range m {
		s = append(s, k)
	}

	return s
}
