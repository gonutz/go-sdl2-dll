//+build windows,amd64

package sdl

import (
	"math"
	"syscall"
	"unsafe"
)

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

// WriteBE64 writes 64 bits in native format to the RWops as big-endian data.
// (https://wiki.libsdl.org/SDL_WriteBE64)
func (rwops *RWops) WriteBE64(value uint64) uint {
	if rwops == nil {
		return 0
	}
	ret, _, _ := writeBE64.Call(uintptr(unsafe.Pointer(rwops)), uintptr(value))
	return uint(ret)
}

// WriteLE64 writes 64 bits in native format to the RWops as little-endian data.
// (https://wiki.libsdl.org/SDL_WriteLE64)
func (rwops *RWops) WriteLE64(value uint64) uint {
	if rwops == nil {
		return 0
	}
	ret, _, _ := writeLE64.Call(uintptr(unsafe.Pointer(rwops)), uintptr(value))
	return uint(ret)
}

// Seek seeks within the RWops data stream.
// (https://wiki.libsdl.org/SDL_RWseek)
func (rwops *RWops) Seek(offset int64, whence int) (int64, error) {
	if rwops == nil {
		return -1, ErrInvalidParameters
	}
	ret, _, _ := syscall.Syscall(
		rwops.seek,
		3,
		uintptr(unsafe.Pointer(rwops)),
		uintptr(offset), // TODO what about 32 bit systems? a uintptr is only 32 bytes there
		uintptr(whence),
	)
	if ret < 0 {
		return int64(ret), GetError()
	}
	return int64(ret), nil
}

// Size returns the size of the data stream in the RWops.
// (https://wiki.libsdl.org/SDL_RWsize)
func (rwops *RWops) Size() (int64, error) {
	ret, _, _ := syscall.Syscall(
		rwops.size,
		1,
		uintptr(unsafe.Pointer(rwops)),
		0,
		0,
	)
	n := int64(ret)
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
	ret, _, _ := readBE64.Call(uintptr(unsafe.Pointer(rwops)))
	return uint64(ret)
}

// ReadLE64 reads 64 bits of little-endian data from the RWops and returns in native format.
// (https://wiki.libsdl.org/SDL_ReadLE64)
func (rwops *RWops) ReadLE64() uint64 {
	if rwops == nil {
		return 0
	}
	ret, _, _ := readLE64.Call(uintptr(unsafe.Pointer(rwops)))
	return uint64(ret)
}

// GetNumTouchFingers returns the number of active fingers for a given touch device.
// (https://wiki.libsdl.org/SDL_GetNumTouchFingers)
func GetNumTouchFingers(t TouchID) int {
	ret, _, _ := getNumTouchFingers.Call(uintptr(t))
	return int(ret)
}

// LoadDollarTemplates loads Dollar Gesture templates from a file.
// (https://wiki.libsdl.org/SDL_LoadDollarTemplates)
func LoadDollarTemplates(t TouchID, src *RWops) int {
	ret, _, _ := loadDollarTemplates.Call(
		uintptr(t),
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
	// that we need to pass in 64 bit uintptrs
	ret, _, _ := gameControllerMappingForGUID.Call(
		uintptr(*((*uint64)(unsafe.Pointer(&guid.data[0])))),
		uintptr(*((*uint64)(unsafe.Pointer(&guid.data[8])))),
	)
	return sdlToGoString(ret)
}

// JoystickGetGUIDString returns an ASCII string representation for a given JoystickGUID.
// (https://wiki.libsdl.org/SDL_JoystickGetGUIDString)
func JoystickGetGUIDString(guid JoystickGUID) string {
	buf := make([]byte, 1024)
	joystickGetGUIDString.Call(
		uintptr(*((*uint64)(unsafe.Pointer(&guid.data[0])))),
		uintptr(*((*uint64)(unsafe.Pointer(&guid.data[8])))),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	return sdlToGoString(uintptr(unsafe.Pointer(&buf[0])))
}

// CopyEx copies a portion of the texture to the current rendering target, optionally rotating it by angle around the given center and also flipping it top-bottom and/or left-right.
// (https://wiki.libsdl.org/SDL_RenderCopyEx)
func (renderer *Renderer) CopyEx(texture *Texture, src, dst *Rect, angle float64, center *Point, flip RendererFlip) error {
	ret, _, _ := renderCopyEx.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(texture)),
		uintptr(unsafe.Pointer(src)),
		uintptr(unsafe.Pointer(dst)),
		uintptr(math.Float64bits(angle)),
		uintptr(unsafe.Pointer(center)),
		uintptr(flip),
	)
	return errorFromInt(int(ret))
}

// CopyExF copies a portion of the texture to the current rendering target, optionally rotating it by angle around the given center and also flipping it top-bottom and/or left-right.
// TODO: (https://wiki.libsdl.org/SDL_RenderCopyExF)
func (renderer *Renderer) CopyExF(texture *Texture, src, dst *FRect, angle float64, center *FPoint, flip RendererFlip) error {
	ret, _, _ := renderCopyExF.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(texture)),
		uintptr(unsafe.Pointer(src)),
		uintptr(unsafe.Pointer(dst)),
		uintptr(math.Float64bits(angle)),
		uintptr(unsafe.Pointer(center)),
		uintptr(flip),
	)
	return errorFromInt(int(ret))
}

// RecordGesture begins recording a gesture on a specified touch device or all touch devices.
// (https://wiki.libsdl.org/SDL_RecordGesture)
func RecordGesture(t TouchID) int {
	ret, _, _ := recordGesture.Call(uintptr(t))
	return int(ret)
}

// GetTouchDevice returns the touch ID with the given index.
// (https://wiki.libsdl.org/SDL_GetTouchDevice)
func GetTouchDevice(index int) TouchID {
	ret, _, _ := getTouchDevice.Call(uintptr(index))
	return TouchID(ret)
}

// SaveDollarTemplate saves a currently loaded Dollar Gesture template.
// (https://wiki.libsdl.org/SDL_SaveDollarTemplate)
func SaveDollarTemplate(g GestureID, src *RWops) int {
	ret, _, _ := saveDollarTemplate.Call(uintptr(g), uintptr(unsafe.Pointer(src)))
	return int(ret)
}

// GetTouchFinger returns the finger object for specified touch device ID and finger index.
// (https://wiki.libsdl.org/SDL_GetTouchFinger)
func GetTouchFinger(t TouchID, index int) *Finger {
	ret, _, _ := getTouchFinger.Call(uintptr(t), uintptr(index))
	return (*Finger)(unsafe.Pointer(ret))
}
