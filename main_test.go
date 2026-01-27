package main

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"testing"
)

func initTestConfig(t *testing.T) {
	viper.SetConfigFile("testdata/config.yaml")
	err := viper.ReadInConfig()
	require.Nil(t, err)
}

func TestMain(t *testing.T) {
	initTestConfig(t)
}
