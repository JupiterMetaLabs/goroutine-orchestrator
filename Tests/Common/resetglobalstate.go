package Common

import "github.com/neerajchowdary889/GoRoutinesManager/types"

// resetGlobalState resets the global singleton for testing
func ResetGlobalState() {
	types.Global = nil
}
