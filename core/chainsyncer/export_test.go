package chainsyncer

func SetNotifyHook(f func()) func() {
	var cleanup func()

	func(f func()) {
		cleanup = func() {
			notifyHook = f
		}
	}(notifyHook)
	notifyHook = f
	return cleanup
}
