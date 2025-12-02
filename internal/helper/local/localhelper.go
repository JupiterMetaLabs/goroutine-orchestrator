package local

import "github.com/JupiterMetaLabs/goroutine-orchestrator/types"

type LocalHelper struct{}

func NewLocalHelper() *LocalHelper {
	return &LocalHelper{}
}

// Convert Map to Slice
func (LH *LocalHelper) MapToSlice(m map[string]*types.LocalManager) []*types.LocalManager {
	slice := make([]*types.LocalManager, 0, len(m))
	for _, v := range m {
		slice = append(slice, v)
	}
	return slice
}

func (LH *LocalHelper) RoutinesMapToSlice(m map[string]*types.Routine) []*types.Routine {
	slice := make([]*types.Routine, 0, len(m))
	for _, v := range m {
		slice = append(slice, v)
	}
	return slice
}