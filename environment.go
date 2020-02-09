package main

import (
	"fmt"
	"os"
	"strings"
)

type duplicateValueError struct {
	description string
	lastkey     string
	currentkey  string
}

func (d *duplicateValueError) Error() string {
	return fmt.Sprintf("value for %q already provided by %q but also found in %q", d.description, "$"+d.lastkey, "$"+d.currentkey)
}

type envValueRequired struct {
	description string
	keyNames    []string
}

func (er *envValueRequired) Error() string {
	return fmt.Sprintf("must provide a value for %s using any of the following: %s", er.description, strings.Join(er.keyNames, ", "))
}

func getAnyEnvironment(description string, envName ...string) (string, error) {
	var value, lastKeyFound string

	for _, v := range envName {
		if s := os.Getenv(v); s != "" {
			if value != "" {
				return "", &duplicateValueError{description: description, lastkey: lastKeyFound, currentkey: v}
			}

			value = s
			lastKeyFound = v
		}
	}

	if value == "" {
		return "", &envValueRequired{description: description, keyNames: envName}
	}

	return value, nil
}
