//+build windows,amd64

package sdl

// GetPerformanceCounter returns the current value of the high resolution counter.
// (https://wiki.libsdl.org/SDL_GetPerformanceCounter)
func GetPerformanceCounter() uint64 {
	ret, _, _ := getPerformanceCounter.Call()
	return uint64(ret)
}

// GetPerformanceFrequency returns the count per second of the high resolution counter.
// (https://wiki.libsdl.org/SDL_GetPerformanceFrequency)
func GetPerformanceFrequency() uint64 {
	ret, _, _ := getPerformanceFrequency.Call()
	return uint64(ret)
}
