package app

import "wholesale.local/orbit/internal/store"

func stableID(name string) string { return store.ID(name) }
