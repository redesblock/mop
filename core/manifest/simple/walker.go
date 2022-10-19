package simple

// WalkEntryFunc is the type of the function called for each entry visited
// by WalkEntry.
type WalkEntryFunc func(path string, entry Entry, err error) error

func (m *manifest) WalkEntry(root string, walkFn WalkEntryFunc) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, v := range m.Entries {
		entry := newEntry(v.Ref, v.Meta)
		err = walkFn(k, entry, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
