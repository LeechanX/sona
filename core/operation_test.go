package core

import (
	"testing"
	_ "fmt"
)

func TestConfigController_Set(t *testing.T) {
	setter, err := GetConfigController()
	if err != nil {
		t.Error(err)
	}
	setter.Set("lebron.james.number", "23")

	value, err := setter.Get("lebron.james.number")
	if err != nil {
		t.Error(err)
	}
	if value != "23" {
		t.Error("I want 23, actual", value)
	}
}

func TestConfigGetter_Get(t *testing.T) {
	getter, err := GetConfigGetter()
	if err != nil {
		t.Error(err)
	}
	value, err := getter.Get("lebron.james.number")
	if err != nil {
		t.Error(err)
	}
	if value != "23" {
		t.Error("I want 23, actual", value)
	}
}
