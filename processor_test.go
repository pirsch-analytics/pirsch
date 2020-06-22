package pirsch

import "testing"

func TestProcessor_Process(t *testing.T) {
	createTestdata(t)
	processor := NewProcessor(NewPostgresStore(db))
	processor.Process()
}

func createTestdata(t *testing.T) {
	cleanupDB(t)
}
