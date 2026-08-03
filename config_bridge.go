package main

import configpkg "hatwmpanel/config"

type Config = configpkg.Config
type Module = configpkg.Module

var (
	GetConfigPath = configpkg.GetConfigPath
	LoadConfig    = configpkg.LoadConfig
)
