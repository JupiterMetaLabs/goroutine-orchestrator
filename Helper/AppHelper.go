package Helper

import "github.com/neerajchowdary889/GoRoutinesManager/types"

// Helper function for AppManager
type AppHelper struct{}

func NewAppHelper() *AppHelper {
	return &AppHelper{}
}

// Convert Map to Slice
func (AH *AppHelper) MapToSlice(m map[string]*types.AppManager) []*types.AppManager {
	slice := make([]*types.AppManager, 0, len(m))
	for _, v := range m {
		slice = append(slice, v)
	}
	return slice
}