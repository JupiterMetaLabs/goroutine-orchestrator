package Global

import (
	"errors"
	"time"

	"github.com/neerajchowdary889/GoRoutinesManager/types"
)

const (
	SET_METRICS_URL = "SET_METRICS_URL"
	SET_SHUTDOWN_TIMEOUT = "SET_SHUTDOWN_TIMEOUT"
	SET_MAX_ROUTINES = "SET_MAX_ROUTINES"
)

type metricsConfig struct {
	Enabled bool
	URL     string
}

func (GM *GlobalManager) UpdateGlobalMetadata(flag string, value interface{}) (*types.Metadata, error) {
	// Get the global manager first
	g, err := types.GetGlobalManager()
	if err != nil {
		return nil, err
	}

	metadata := g.NewMetadata()

	switch flag {
	case SET_METRICS_URL:
		// Accept several input shapes:
		//  - string -> URL (enabled = true)
		//  - metricsConfig
		//  - []interface{}{bool, string} or [2]interface{}{bool, string}
		switch v := value.(type) {
		case string:
			metadata.SetMetrics(true, v)
		case metricsConfig:
			metadata.SetMetrics(v.Enabled, v.URL)
		case *metricsConfig:
			metadata.SetMetrics(v.Enabled, v.URL)
		case []interface{}:
			if len(v) != 2 {
				return nil, errors.New("metrics: expected slice of length 2: [enabled(bool), url(string)]")
			}
			enabled, ok1 := v[0].(bool)
			url, ok2 := v[1].(string)
			if !ok1 || !ok2 {
				return nil, errors.New("metrics: expected [bool, string] in slice")
			}
			metadata.SetMetrics(enabled, url)
		case [2]interface{}:
			enabled, ok1 := v[0].(bool)
			url, ok2 := v[1].(string)
			if !ok1 || !ok2 {
				return nil, errors.New("metrics: expected [bool, string] array")
			}
			metadata.SetMetrics(enabled, url)
		default:
			return nil, errors.New("metrics: unsupported value type; expected string, metricsConfig, or [bool,string]")
		}

	case SET_SHUTDOWN_TIMEOUT:
		switch t := value.(type) {
		case time.Duration:
			metadata.SetShutdownTimeout(t)
		case *time.Duration:
			metadata.SetShutdownTimeout(*t)
		default:
			return nil, errors.New("shutdown timeout: expected time.Duration")
		}

	case SET_MAX_ROUTINES:
		switch n := value.(type) {
		case int:
			metadata.SetMaxRoutines(n)
		case int32:
			metadata.SetMaxRoutines(int(n))
		case int64:
			metadata.SetMaxRoutines(int(n))
		case *int:
			metadata.SetMaxRoutines(*n)
		default:
			return nil, errors.New("max routines: expected integer type")
		}

	default:
		return nil, errors.New("unknown update flag")
	}

	return metadata, nil
}

func (GM *GlobalManager) GetGlobalMetadata() (*types.Metadata, error) {
	g, err := types.GetGlobalManager()
	if err != nil {
		return nil, err
	}
	return g.GetMetadata(), nil
}
