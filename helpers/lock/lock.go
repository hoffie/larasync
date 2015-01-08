package lock

var currentManager Manager

func CurrentManager() Manager {
	if currentManager == nil {
		currentManager = newProcessManager()
	}
	return currentManager
}
