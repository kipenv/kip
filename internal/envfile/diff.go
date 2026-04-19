package envfile

// DiffResult contains the result of comparing two env maps.
type DiffResult struct {
	Missing []string            // Keys in remote but not in local
	Extra   []string            // Keys in local but not in remote
	Changed []DiffChange        // Keys present in both but with different values
	Same    []string            // Keys with identical values
}

// DiffChange represents a key whose value differs between local and remote.
type DiffChange struct {
	Key        string
	LocalValue string
	RemoteValue string
}

// Diff compares local and remote env maps and returns the differences.
func Diff(local, remote map[string]string) DiffResult {
	var result DiffResult

	// Find missing and changed keys
	for key, remoteVal := range remote {
		localVal, exists := local[key]
		if !exists {
			result.Missing = append(result.Missing, key)
		} else if localVal != remoteVal {
			result.Changed = append(result.Changed, DiffChange{
				Key:         key,
				LocalValue:  localVal,
				RemoteValue: remoteVal,
			})
		} else {
			result.Same = append(result.Same, key)
		}
	}

	// Find extra keys (in local but not remote)
	for key := range local {
		if _, exists := remote[key]; !exists {
			result.Extra = append(result.Extra, key)
		}
	}

	return result
}

// HasDifferences returns true if there are any differences.
func (d DiffResult) HasDifferences() bool {
	return len(d.Missing) > 0 || len(d.Extra) > 0 || len(d.Changed) > 0
}
