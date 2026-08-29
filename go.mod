module github.com/arr-ai/frozen

go 1.25.0

require golang.org/x/sys v0.47.0

require github.com/arr-ai/hash v1.2.0

// v1.8.0–v1.9.0 broke reflect.DeepEqual for Map/Set due to an unexported
// func field in the tree struct. Fixed in v1.10.0.
retract [v1.8.0, v1.9.0]
