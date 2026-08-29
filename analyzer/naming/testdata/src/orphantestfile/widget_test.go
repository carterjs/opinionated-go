package orphantestfile // want "widget_test.go has no corresponding widget.go"

import "testing"

// Bad: no widget.go anywhere in this package.
func TestWidget(t *testing.T) {}
