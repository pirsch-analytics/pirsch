package pirsch

// This is a list of all storage backends to be used in tests.
// We test against real databases, so for testing all storage solutions must be installed an configured.
func testStorageBackends() []Store {
	return []Store{
		NewPostgresStore(db),
	}
}
