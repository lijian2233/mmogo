package logger

var DummyLogger = &dummyLogger{}

type dummyLogger struct{}

func (d *dummyLogger) Debug(args ...interface{})                   {}
func (d *dummyLogger) Debugf(format string, args ...interface{})   {}
func (d *dummyLogger) Info(args ...interface{})                    {}
func (d *dummyLogger) Infof(format string, args ...interface{})    {}
func (d *dummyLogger) Warning(args ...interface{})                 {}
func (d *dummyLogger) Warningf(format string, args ...interface{}) {}
func (d *dummyLogger) Error(args ...interface{})                   {}
func (d *dummyLogger) Errorf(format string, args ...interface{})   {}
func (d *dummyLogger) Fatal(args ...interface{})                   {}
func (d *dummyLogger) Fatalf(format string, args ...interface{})   {}
