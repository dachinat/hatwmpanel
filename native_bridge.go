package main

import nativepanel "hatwmpanel/native"

type NativePanel = nativepanel.Panel

func NewNativePanel(cfg Config) (*NativePanel, error) {
	return nativepanel.NewPanel(cfg)
}
