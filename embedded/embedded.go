package embedded

import "embed"

//go:embed assets/* templates/* templates/*/* skill/*
var Content embed.FS
