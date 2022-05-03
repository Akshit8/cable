package cable

import (
	"testing"
)

func TestLogger(t *testing.T) {
	t.Parallel()

	logger := NewLogger()
	logger.Info("Hello")
	logger.Infof("Hello %s", "World")
	logger.Error("Hello")
	logger.Errorf("Hello %s", "World")
}
