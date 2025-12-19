package logging

import (
	"context"
	"fmt"
	"math"
	"os"

	"github.com/rs/zerolog"
)

type Type string

const (
	TypeService   Type = "service"
	TypeComponent Type = "component"
)

type zerologHook struct{}

func (z zerologHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	ctx := e.GetCtx()
	if ctx == nil {
		return
	}

	if service, ok := ctx.Value(TypeService).(string); ok {
		e.Str("service", service)
	}

	if component, ok := ctx.Value(TypeComponent).(string); ok {
		e.Str("component", component)
	}
}

// New initializes global log settings then puts the settings into the context object.
func New(ctx context.Context, logLevel int) (context.Context, error) {

	if logLevel > math.MaxInt8 || logLevel < math.MinInt8 {
		return nil, fmt.Errorf("conversion overflow on log level %v, expected within int8", logLevel)
	}
	level := int8(logLevel)
	zerolog.SetGlobalLevel(zerolog.Level(level))
	newCtx := zerolog.New(os.Stdout).Hook(&zerologHook{}).With().Timestamp().Caller().Logger().WithContext(ctx)

	return newCtx, nil
}
