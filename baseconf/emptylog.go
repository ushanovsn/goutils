package baseconf

// Empty type for interface compatible
type elog struct{}

// Empty plug for interface compatible
func (obj *elog) Debug(string) {}

// Empty plug for interface compatible
func (obj *elog) Info(string) {}

// Empty plug for interface compatible
func (obj *elog) Warn(string) {}

// Empty plug for interface compatible
func (obj *elog) Error(string) {}
