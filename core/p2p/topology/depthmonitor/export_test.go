package depthmonitor

var (
	ManageWait    = &manageWait
	MinimumRadius = &minimumRadius
)

func (s *Service) StorageDepth() uint8 {
	return s.bs.GetReserveState().StorageRadius
}

func (s *Service) SetStorageRadius(r uint8) {
	_ = s.bs.SetStorageRadius(func(_ uint8) uint8 {
		return r
	})
}
