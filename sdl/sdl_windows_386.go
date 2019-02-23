//+build windows,386

package sdl

import (
	"syscall"
	"unsafe"
)

// GetPerformanceCounter returns the current value of the high resolution counter.
// (https://wiki.libsdl.org/SDL_GetPerformanceCounter)
func GetPerformanceCounter() uint64 {
	r1, r2, _ := getPerformanceCounter.Call()
	return uint64(r2)<<32 + uint64(r1)
}

// GetPerformanceFrequency returns the count per second of the high resolution counter.
// (https://wiki.libsdl.org/SDL_GetPerformanceFrequency)
func GetPerformanceFrequency() uint64 {
	r1, r2, _ := getPerformanceFrequency.Call()
	return uint64(r2)<<32 + uint64(r1)
}

// WriteBE64 writes 64 bits in native format to the RWops as big-endian data.
// (https://wiki.libsdl.org/SDL_WriteBE64)
func (rwops *RWops) WriteBE64(value uint64) uint {
	if rwops == nil {
		return 0
	}
	a := uint32(value)
	b := uint32(value >> 32)
	ret, _, _ := writeBE64.Call(
		uintptr(unsafe.Pointer(rwops)),
		uintptr(a),
		uintptr(b),
	)
	return uint(ret)
}

// WriteLE64 writes 64 bits in native format to the RWops as little-endian data.
// (https://wiki.libsdl.org/SDL_WriteLE64)
func (rwops *RWops) WriteLE64(value uint64) uint {
	if rwops == nil {
		return 0
	}
	a := uint32(value)
	b := uint32(value >> 32)
	ret, _, _ := writeLE64.Call(
		uintptr(unsafe.Pointer(rwops)),
		uintptr(a),
		uintptr(b),
	)
	return uint(ret)
}

// Seek seeks within the RWops data stream.
// (https://wiki.libsdl.org/SDL_RWseek)
func (rwops *RWops) Seek(offset int64, whence int) (int64, error) {
	if rwops == nil {
		return -1, ErrInvalidParameters
	}
	a := uint32(offset)
	b := uint32(offset >> 32)
	ret, _, _ := syscall.Syscall6(
		rwops.seek,
		4,
		uintptr(unsafe.Pointer(rwops)),
		uintptr(a),
		uintptr(b),
		uintptr(whence),
		0,
		0,
	)
	if ret < 0 {
		return int64(ret), GetError()
	}
	return int64(ret), nil
}

// Size returns the size of the data stream in the RWops.
// (https://wiki.libsdl.org/SDL_RWsize)
func (rwops *RWops) Size() (int64, error) {
	r1, r2, _ := syscall.Syscall(
		rwops.size,
		1,
		uintptr(unsafe.Pointer(rwops)),
		0,
		0,
	)
	n := int64(uint64(r2)<<32 + uint64(r1))
	if n < 0 {
		return n, GetError()
	}
	return n, nil
}

// ReadBE64 reads 64 bits of big-endian data from the RWops and returns in native format.
// (https://wiki.libsdl.org/SDL_ReadBE64)
func (rwops *RWops) ReadBE64() uint64 {
	if rwops == nil {
		return 0
	}
	r1, r2, _ := readBE64.Call(uintptr(unsafe.Pointer(rwops)))
	return uint64(r2)<<32 + uint64(r1)
}

// ReadLE64 reads 64 bits of little-endian data from the RWops and returns in native format.
// (https://wiki.libsdl.org/SDL_ReadLE64)
func (rwops *RWops) ReadLE64() uint64 {
	if rwops == nil {
		return 0
	}
	r1, r2, _ := readLE64.Call(uintptr(unsafe.Pointer(rwops)))
	return uint64(r2)<<32 + uint64(r1)
}

// GetNumTouchFingers returns the number of active fingers for a given touch device.
// (https://wiki.libsdl.org/SDL_GetNumTouchFingers)
func GetNumTouchFingers(t TouchID) int {
	a := uint32(t)
	b := uint32(t >> 32)
	ret, _, _ := getNumTouchFingers.Call(uintptr(a), uintptr(b))
	return int(ret)
}

// LoadDollarTemplates loads Dollar Gesture templates from a file.
// (https://wiki.libsdl.org/SDL_LoadDollarTemplates)
func LoadDollarTemplates(t TouchID, src *RWops) int {
	a := uint32(t)
	b := uint32(t >> 32)
	ret, _, _ := loadDollarTemplates.Call(
		uintptr(a),
		uintptr(b),
		uintptr(unsafe.Pointer(src)),
	)
	return int(ret)
}

// GameControllerMappingForGUID returns the game controller mapping string for a
// given GUID.
// (https://wiki.libsdl.org/SDL_GameControllerMappingForGUID)
func GameControllerMappingForGUID(guid JoystickGUID) string {
	// JoystickGUID contains
	// 	data [16]byte
	// that we need to pass in 32 bit uintptrs
	ret, _, _ := gameControllerMappingForGUID.Call(
		uintptr(*((*uint32)(unsafe.Pointer(&guid.data[0])))),
		uintptr(*((*uint32)(unsafe.Pointer(&guid.data[4])))),
		uintptr(*((*uint32)(unsafe.Pointer(&guid.data[8])))),
		uintptr(*((*uint32)(unsafe.Pointer(&guid.data[12])))),
	)
	return sdlToGoString(ret)
}

// JoystickGetGUIDString returns an ASCII string representation for a given JoystickGUID.
// (https://wiki.libsdl.org/SDL_JoystickGetGUIDString)
func JoystickGetGUIDString(guid JoystickGUID) string {
	buf := make([]byte, 1024)
	joystickGetGUIDString.Call(
		uintptr(*((*uint64)(unsafe.Pointer(&guid.data[0])))),
		uintptr(*((*uint64)(unsafe.Pointer(&guid.data[4])))),
		uintptr(*((*uint64)(unsafe.Pointer(&guid.data[8])))),
		uintptr(*((*uint64)(unsafe.Pointer(&guid.data[12])))),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	return sdlToGoString(uintptr(unsafe.Pointer(&buf[0])))
}
