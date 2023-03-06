package log

import "testing"

func TestNewZeroLog(t *testing.T) {
	opts := NewOptions(WithDirectory("./logs/"),
		WithFilename("app.log"),
		WithLevel("info"),
		WithRotate(true),
		WithRotateKeepAge(86400),
		WithRotateMaxSize(1024),
		WithRotateKeepNumber(10),
		WithConsoleLoggingEnabled(true),
		WithReportCaller(true),
		WithReportCaller(true),
	)
	err := Register("zlog", opts)

	if err != nil {
		panic(err)
	}

	GetLogger("zlog").Info().Msg("test2")
}
