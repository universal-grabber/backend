package config

type Config struct {
	ListenAddr string
	ListenPort uint8
	LogLevel   LogLevel

	DbConnection struct {
		Host           string
		Port           uint8
		DbName         string
		CollectionName string
	}
}

type LogLevel string

const (
	PanicLevel LogLevel = "Panic"
	FatalLevel          = "Fatal"
	ErrorLevel          = "Error"
	WarnLevel           = "Warn"
	InfoLevel           = "Info"
	DebugLevel          = "Debug"
	TraceLevel          = "Trace"
)
