//+build windows

/*
Package sdl is SDL2 wrapped for Go users. It enables interoperability between
Go and the SDL2 library which is written in C. That means the original SDL2
installation is required for this to work. SDL2 is a cross-platform
development library designed to provide low level access to audio, keyboard,
mouse, joystick, and graphics hardware via OpenGL and Direct3D.
*/
package sdl

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"reflect"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

func init() {
	// Make sure the main goroutine is bound to the main thread.
	runtime.LockOSThread()
}

// Audio format masks.
// (https://wiki.libsdl.org/SDL_AudioFormat)
const (
	AUDIO_MASK_BITSIZE  = 0xFF
	AUDIO_MASK_DATATYPE = 1 << 8
	AUDIO_MASK_ENDIAN   = 1 << 12
	AUDIO_MASK_SIGNED   = 1 << 15
)

// Audio format values.
// (https://wiki.libsdl.org/SDL_AudioFormat)
const (
	AUDIO_S8 = 0x8008 // unsigned 8-bit samples
	AUDIO_U8 = 0x0008 // signed 8-bit samples

	AUDIO_S16LSB = 0x8010       // signed 16-bit samples in little-endian byte order
	AUDIO_S16MSB = 0x9010       // signed 16-bit samples in big-endian byte order
	AUDIO_S16SYS = AUDIO_S16LSB // signed 16-bit samples in native byte order
	AUDIO_S16    = AUDIO_S16LSB // AUDIO_S16LSB
	AUDIO_U16LSB = 0x0010       // unsigned 16-bit samples in little-endian byte order
	AUDIO_U16MSB = 0x1010       // unsigned 16-bit samples in big-endian byte order
	AUDIO_U16SYS = AUDIO_U16LSB // unsigned 16-bit samples in native byte order
	AUDIO_U16    = AUDIO_U16LSB // AUDIO_U16LSB

	AUDIO_S32LSB = 0x8020       // 32-bit integer samples in little-endian byte order
	AUDIO_S32MSB = 0x9020       // 32-bit integer samples in big-endian byte order
	AUDIO_S32SYS = AUDIO_S32LSB // 32-bit integer samples in native byte order
	AUDIO_S32    = AUDIO_S32LSB // AUDIO_S32LSB

	AUDIO_F32LSB = 0x8120       // 32-bit floating point samples in little-endian byte order
	AUDIO_F32MSB = 0x9120       // 32-bit floating point samples in big-endian byte order
	AUDIO_F32SYS = AUDIO_F32LSB // 32-bit floating point samples in native byte order
	AUDIO_F32    = AUDIO_F32LSB // AUDIO_F32LSB
)

// AllowedChanges flags specify how SDL should behave when a device cannot offer a specific feature. If the application requests a feature that the hardware doesn't offer, SDL will always try to get the closest equivalent. Used in OpenAudioDevice().
// (https://wiki.libsdl.org/SDL_OpenAudioDevice)
const (
	AUDIO_ALLOW_FREQUENCY_CHANGE = 0x00000001
	AUDIO_ALLOW_FORMAT_CHANGE    = 0x00000002
	AUDIO_ALLOW_CHANNELS_CHANGE  = 0x00000004
	AUDIO_ALLOW_SAMPLES_CHANGE   = 0x00000008
	AUDIO_ALLOW_ANY_CHANGE       = (AUDIO_ALLOW_FREQUENCY_CHANGE | AUDIO_ALLOW_FORMAT_CHANGE | AUDIO_ALLOW_CHANNELS_CHANGE | AUDIO_ALLOW_SAMPLES_CHANGE)
)

// An enumeration of audio device states used in GetAudioDeviceStatus() and GetAudioStatus().
// (https://wiki.libsdl.org/SDL_AudioStatus)
const (
	AUDIO_STOPPED AudioStatus = iota // audio device is stopped
	AUDIO_PLAYING                    // audio device is playing
	AUDIO_PAUSED                     // audio device is paused
)

const (
	BLENDMODE_NONE    = 0x00000000 // no blending
	BLENDMODE_BLEND   = 0x00000001 // alpha blending
	BLENDMODE_ADD     = 0x00000002 // additive blending
	BLENDMODE_MOD     = 0x00000004 // color modulate
	BLENDMODE_INVALID = 0x7FFFFFFF
)

const (
	BLENDOPERATION_ADD          = 0x1 // dst + src: supported by all renderers
	BLENDOPERATION_SUBTRACT     = 0x2 // dst - src : supported by D3D9, D3D11, OpenGL, OpenGLES
	BLENDOPERATION_REV_SUBTRACT = 0x3 // src - dst : supported by D3D9, D3D11, OpenGL, OpenGLES
	BLENDOPERATION_MINIMUM      = 0x4 // min(dst, src) : supported by D3D11
	BLENDOPERATION_MAXIMUM      = 0x5 // max(dst, src) : supported by D3D11
)

const (
	BLENDFACTOR_ZERO                = 0x1 // 0, 0, 0, 0
	BLENDFACTOR_ONE                 = 0x2 // 1, 1, 1, 1
	BLENDFACTOR_SRC_COLOR           = 0x3 // srcR, srcG, srcB, srcA
	BLENDFACTOR_ONE_MINUS_SRC_COLOR = 0x4 // 1-srcR, 1-srcG, 1-srcB, 1-srcA
	BLENDFACTOR_SRC_ALPHA           = 0x5 // srcA, srcA, srcA, srcA
	BLENDFACTOR_ONE_MINUS_SRC_ALPHA = 0x6 // 1-srcA, 1-srcA, 1-srcA, 1-srcA
	BLENDFACTOR_DST_COLOR           = 0x7 // dstR, dstG, dstB, dstA
	BLENDFACTOR_ONE_MINUS_DST_COLOR = 0x8 // 1-dstR, 1-dstG, 1-dstB, 1-dstA
	BLENDFACTOR_DST_ALPHA           = 0x9 // dstA, dstA, dstA, dstA
	BLENDFACTOR_ONE_MINUS_DST_ALPHA = 0xA // 1-dstA, 1-dstA, 1-dstA, 1-dstA
)

// Endian-specific values.
// (https://wiki.libsdl.org/CategoryEndian)
const (
	BYTEORDER  = LIL_ENDIAN // macro that corresponds to the byte order used by the processor type it was compiled for
	LIL_ENDIAN = 1234       // byte order is 1234, where the least significant byte is stored first
	BIG_ENDIAN = 4321       // byte order is 4321, where the most significant byte is stored first
)

// SDL error codes with their corresponding predefined strings.
const (
	ENOMEM      ErrorCode = iota // out of memory
	EFREAD                       // error reading from datastream
	EFWRITE                      // error writing to datastream
	EFSEEK                       // error seeking in datastream
	UNSUPPORTED                  // that operation is not supported
	LASTERROR                    // the highest numbered predefined error
)

// Enumeration of the types of events that can be delivered.
// (https://wiki.libsdl.org/SDL_EventType)
const (
	FIRSTEVENT = 0 // do not remove (unused)

	// Application events
	QUIT = 0x100 // user-requested quit

	// Android, iOS and WinRT events
	APP_TERMINATING         = 0x100 + 1 // OS is terminating the application
	APP_LOWMEMORY           = 0x100 + 2 // OS is low on memory; free some
	APP_WILLENTERBACKGROUND = 0x100 + 3 // application is entering background
	APP_DIDENTERBACKGROUND  = 0x100 + 4 //application entered background
	APP_WILLENTERFOREGROUND = 0x100 + 5 // application is entering foreground
	APP_DIDENTERFOREGROUND  = 0x100 + 6 // application entered foreground

	// Window events
	WINDOWEVENT = 0x200     // window state change
	SYSWMEVENT  = 0x200 + 1 // system specific event

	// Keyboard events
	KEYDOWN       = 0x300     // key pressed
	KEYUP         = 0x300 + 1 // key released
	TEXTEDITING   = 0x300 + 2 // keyboard text editing (composition)
	TEXTINPUT     = 0x300 + 3 // keyboard text input
	KEYMAPCHANGED = 0x300 + 4 // keymap changed due to a system event such as an input language or keyboard layout change (>= SDL 2.0.4)

	// Mouse events
	MOUSEMOTION     = 0x400     // mouse moved
	MOUSEBUTTONDOWN = 0x400 + 1 // mouse button pressed
	MOUSEBUTTONUP   = 0x400 + 2 // mouse button released
	MOUSEWHEEL      = 0x400 + 3 // mouse wheel motion

	// Joystick events
	JOYAXISMOTION    = 0x600     // joystick axis motion
	JOYBALLMOTION    = 0x600 + 1 // joystick trackball motion
	JOYHATMOTION     = 0x600 + 2 // joystick hat position change
	JOYBUTTONDOWN    = 0x600 + 3 // joystick button pressed
	JOYBUTTONUP      = 0x600 + 4 // joystick button released
	JOYDEVICEADDED   = 0x600 + 5 // joystick connected
	JOYDEVICEREMOVED = 0x600 + 6 // joystick disconnected

	// Game controller events
	CONTROLLERAXISMOTION     = 0x650     // controller axis motion
	CONTROLLERBUTTONDOWN     = 0x650 + 1 // controller button pressed
	CONTROLLERBUTTONUP       = 0x650 + 2 // controller button released
	CONTROLLERDEVICEADDED    = 0x650 + 3 // controller connected
	CONTROLLERDEVICEREMOVED  = 0x650 + 4 // controller disconnected
	CONTROLLERDEVICEREMAPPED = 0x650 + 5 // controller mapping updated

	// Touch events
	FINGERDOWN   = 0x700     // user has touched input device
	FINGERUP     = 0x700 + 1 // user stopped touching input device
	FINGERMOTION = 0x700 + 3 // user is dragging finger on input device

	// Gesture events
	DOLLARGESTURE = 0x800
	DOLLARRECORD  = 0x800 + 1
	MULTIGESTURE  = 0x800 + 2

	// Clipboard events
	CLIPBOARDUPDATE = 0x900 // the clipboard changed

	// Drag and drop events
	DROPFILE     = 0x1000     // the system requests a file open
	DROPTEXT     = 0x1000 + 1 // text/plain drag-and-drop event
	DROPBEGIN    = 0x1000 + 2 // a new set of drops is beginning (NULL filename)
	DROPCOMPLETE = 0x1000 + 3 // current set of drops is now complete (NULL filename)

	// Audio hotplug events
	AUDIODEVICEADDED   = 0x1100     // a new audio device is available (>= SDL 2.0.4)
	AUDIODEVICEREMOVED = 0x1100 + 1 // an audio device has been removed (>= SDL 2.0.4)

	// Sensor events
	SENSORUPDATE = 0x1200 // a sensor was updated

	// Render events
	RENDER_TARGETS_RESET = 0x2000     // the render targets have been reset and their contents need to be updated (>= SDL 2.0.2)
	RENDER_DEVICE_RESET  = 0x2000 + 1 // the device has been reset and all textures need to be recreated (>= SDL 2.0.4)

	// These are for your use, and should be allocated with RegisterEvents()
	USEREVENT = 0x8000 // a user-specified event
	LASTEVENT = 0xFFFF // (only for bounding internal arrays)
)

// Actions for PeepEvents().
// (https://wiki.libsdl.org/SDL_PeepEvents)
const (
	ADDEVENT  = iota // up to numevents events will be added to the back of the event queue
	PEEKEVENT        // up to numevents events at the front of the event queue, within the specified minimum and maximum type, will be returned and will not be removed from the queue
	GETEVENT         // up to numevents events at the front of the event queue, within the specified minimum and maximum type, will be returned and will be removed from the queue
)

// Toggles for different event state functions.
const (
	QUERY   = -1
	IGNORE  = 0
	DISABLE = 0
	ENABLE  = 1
)

// Types of game controller inputs.
const (
	CONTROLLER_BINDTYPE_NONE = iota
	CONTROLLER_BINDTYPE_BUTTON
	CONTROLLER_BINDTYPE_AXIS
	CONTROLLER_BINDTYPE_HAT
)

// An enumeration of axes available from a controller.
// (https://wiki.libsdl.org/SDL_GameControllerAxis)
const (
	CONTROLLER_AXIS_INVALID = iota - 1
	CONTROLLER_AXIS_LEFTX
	CONTROLLER_AXIS_LEFTY
	CONTROLLER_AXIS_RIGHTX
	CONTROLLER_AXIS_RIGHTY
	CONTROLLER_AXIS_TRIGGERLEFT
	CONTROLLER_AXIS_TRIGGERRIGHT
	CONTROLLER_AXIS_MAX
)

// An enumeration of buttons available from a controller.
// (https://wiki.libsdl.org/SDL_GameControllerButton)
const (
	CONTROLLER_BUTTON_INVALID = iota - 1
	CONTROLLER_BUTTON_A
	CONTROLLER_BUTTON_B
	CONTROLLER_BUTTON_X
	CONTROLLER_BUTTON_Y
	CONTROLLER_BUTTON_BACK
	CONTROLLER_BUTTON_GUIDE
	CONTROLLER_BUTTON_START
	CONTROLLER_BUTTON_LEFTSTICK
	CONTROLLER_BUTTON_RIGHTSTICK
	CONTROLLER_BUTTON_LEFTSHOULDER
	CONTROLLER_BUTTON_RIGHTSHOULDER
	CONTROLLER_BUTTON_DPAD_UP
	CONTROLLER_BUTTON_DPAD_DOWN
	CONTROLLER_BUTTON_DPAD_LEFT
	CONTROLLER_BUTTON_DPAD_RIGHT
	CONTROLLER_BUTTON_MAX
)

// Haptic effects.
// (https://wiki.libsdl.org/SDL_HapticEffect)
const (
	HAPTIC_CONSTANT     = 1 << iota // constant haptic effect
	HAPTIC_SINE                     // periodic haptic effect that simulates sine waves
	HAPTIC_LEFTRIGHT                // haptic effect for direct control over high/low frequency motors
	HAPTIC_TRIANGLE                 // periodic haptic effect that simulates triangular waves
	HAPTIC_SAWTOOTHUP               // periodic haptic effect that simulates saw tooth up waves
	HAPTIC_SAWTOOTHDOWN             // periodic haptic effect that simulates saw tooth down waves
	HAPTIC_RAMP                     // ramp haptic effect
	HAPTIC_SPRING                   // condition haptic effect that simulates a spring.  Effect is based on the axes position
	HAPTIC_DAMPER                   // condition haptic effect that simulates dampening.  Effect is based on the axes velocity
	HAPTIC_INERTIA                  // condition haptic effect that simulates inertia.  Effect is based on the axes acceleration
	HAPTIC_FRICTION                 // condition haptic effect that simulates friction.  Effect is based on the axes movement
	HAPTIC_CUSTOM                   // user defined custom haptic effect
	HAPTIC_GAIN                     // device supports setting the global gain
	HAPTIC_AUTOCENTER               // device supports setting autocenter
	HAPTIC_STATUS                   // device can be queried for effect status
	HAPTIC_PAUSE                    // device can be paused
)

// Direction encodings.
// (https://wiki.libsdl.org/SDL_HapticDirection)
const (
	HAPTIC_POLAR     = 0          // uses polar coordinates for the direction
	HAPTIC_CARTESIAN = 1          // uses cartesian coordinates for the direction
	HAPTIC_SPHERICAL = 2          // uses spherical coordinates for the direction
	HAPTIC_INFINITY  = 4294967295 // used to play a device an infinite number of times
)

// Configuration hints
// (https://wiki.libsdl.org/CategoryHints)
const (
	HINT_FRAMEBUFFER_ACCELERATION                 = "SDL_FRAMEBUFFER_ACCELERATION"                 // specifies how 3D acceleration is used with Window.GetSurface()
	HINT_RENDER_DRIVER                            = "SDL_RENDER_DRIVER"                            // specifies which render driver to use
	HINT_RENDER_OPENGL_SHADERS                    = "SDL_RENDER_OPENGL_SHADERS"                    // specifies whether the OpenGL render driver uses shaders
	HINT_RENDER_DIRECT3D_THREADSAFE               = "SDL_RENDER_DIRECT3D_THREADSAFE"               // specifies whether the Direct3D device is initialized for thread-safe operations
	HINT_RENDER_DIRECT3D11_DEBUG                  = "SDL_RENDER_DIRECT3D11_DEBUG"                  // specifies a variable controlling whether to enable Direct3D 11+'s Debug Layer
	HINT_RENDER_SCALE_QUALITY                     = "SDL_RENDER_SCALE_QUALITY"                     // specifies scaling quality
	HINT_RENDER_VSYNC                             = "SDL_RENDER_VSYNC"                             // specifies whether sync to vertical refresh is enabled or disabled in CreateRenderer() to avoid tearing
	HINT_VIDEO_ALLOW_SCREENSAVER                  = "SDL_VIDEO_ALLOW_SCREENSAVER"                  // specifies whether the screensaver is enabled
	HINT_VIDEO_X11_NET_WM_PING                    = "SDL_VIDEO_X11_NET_WM_PING"                    // specifies whether the X11 _NET_WM_PING protocol should be supported
	HINT_VIDEO_X11_XVIDMODE                       = "SDL_VIDEO_X11_XVIDMODE"                       // specifies whether the X11 VidMode extension should be used
	HINT_VIDEO_X11_XINERAMA                       = "SDL_VIDEO_X11_XINERAMA"                       // specifies whether the X11 Xinerama extension should be used
	HINT_VIDEO_X11_XRANDR                         = "SDL_VIDEO_X11_XRANDR"                         // specifies whether the X11 XRandR extension should be used
	HINT_GRAB_KEYBOARD                            = "SDL_GRAB_KEYBOARD"                            // specifies whether grabbing input grabs the keyboard
	HINT_MOUSE_RELATIVE_MODE_WARP                 = "SDL_MOUSE_RELATIVE_MODE_WARP"                 // specifies whether relative mouse mode is implemented using mouse warping
	HINT_VIDEO_MINIMIZE_ON_FOCUS_LOSS             = "SDL_VIDEO_MINIMIZE_ON_FOCUS_LOSS"             // specifies if a Window is minimized if it loses key focus when in fullscreen mode
	HINT_IDLE_TIMER_DISABLED                      = "SDL_IOS_IDLE_TIMER_DISABLED"                  // specifies a variable controlling whether the idle timer is disabled on iOS
	HINT_IME_INTERNAL_EDITING                     = "SDL_IME_INTERNAL_EDITING"                     // specifies whether certain IMEs should handle text editing internally instead of sending TextEditingEvents
	HINT_ORIENTATIONS                             = "SDL_IOS_ORIENTATIONS"                         // specifies a variable controlling which orientations are allowed on iOS
	HINT_ACCELEROMETER_AS_JOYSTICK                = "SDL_ACCELEROMETER_AS_JOYSTICK"                // specifies whether the Android / iOS built-in accelerometer should be listed as a joystick device, rather than listing actual joysticks only
	HINT_XINPUT_ENABLED                           = "SDL_XINPUT_ENABLED"                           // specifies if Xinput gamepad devices are detected
	HINT_XINPUT_USE_OLD_JOYSTICK_MAPPING          = "SDL_XINPUT_USE_OLD_JOYSTICK_MAPPING"          // specifies that SDL should use the old axis and button mapping for XInput devices
	HINT_GAMECONTROLLERCONFIG                     = "SDL_GAMECONTROLLERCONFIG"                     // specifies extra gamecontroller db entries
	HINT_JOYSTICK_ALLOW_BACKGROUND_EVENTS         = "SDL_JOYSTICK_ALLOW_BACKGROUND_EVENTS"         // specifies if joystick (and gamecontroller) events are enabled even when the application is in the background
	HINT_ALLOW_TOPMOST                            = "SDL_ALLOW_TOPMOST"                            // specifies if top most bit on an SDL Window can be set
	HINT_THREAD_STACK_SIZE                        = "SDL_THREAD_STACK_SIZE"                        // specifies a variable specifying SDL's threads stack size in bytes or "0" for the backend's default size
	HINT_TIMER_RESOLUTION                         = "SDL_TIMER_RESOLUTION"                         // specifies the timer resolution in milliseconds
	HINT_VIDEO_HIGHDPI_DISABLED                   = "SDL_VIDEO_HIGHDPI_DISABLED"                   // specifies if high-DPI windows ("Retina" on Mac and iOS) are not allowed
	HINT_MAC_BACKGROUND_APP                       = "SDL_MAC_BACKGROUND_APP"                       // specifies if the SDL app should not be forced to become a foreground process on Mac OS X
	HINT_MAC_CTRL_CLICK_EMULATE_RIGHT_CLICK       = "SDL_MAC_CTRL_CLICK_EMULATE_RIGHT_CLICK"       // specifies whether ctrl+click should generate a right-click event on Mac
	HINT_VIDEO_WIN_D3DCOMPILER                    = "SDL_VIDEO_WIN_D3DCOMPILER"                    // specifies which shader compiler to preload when using the Chrome ANGLE binaries
	HINT_VIDEO_WINDOW_SHARE_PIXEL_FORMAT          = "SDL_VIDEO_WINDOW_SHARE_PIXEL_FORMAT"          // specifies the address of another Window* (as a hex string formatted with "%p")
	HINT_WINRT_PRIVACY_POLICY_URL                 = "SDL_WINRT_PRIVACY_POLICY_URL"                 // specifies a URL to a WinRT app's privacy policy
	HINT_WINRT_PRIVACY_POLICY_LABEL               = "SDL_WINRT_PRIVACY_POLICY_LABEL"               // specifies a label text for a WinRT app's privacy policy link
	HINT_WINRT_HANDLE_BACK_BUTTON                 = "SDL_WINRT_HANDLE_BACK_BUTTON"                 // specifies a variable to allow back-button-press events on Windows Phone to be marked as handled
	HINT_VIDEO_MAC_FULLSCREEN_SPACES              = "SDL_VIDEO_MAC_FULLSCREEN_SPACES"              // specifies policy for fullscreen Spaces on Mac OS X
	HINT_NO_SIGNAL_HANDLERS                       = "SDL_NO_SIGNAL_HANDLERS"                       // specifies not to catch the SIGINT or SIGTERM signals
	HINT_WINDOW_FRAME_USABLE_WHILE_CURSOR_HIDDEN  = "SDL_WINDOW_FRAME_USABLE_WHILE_CURSOR_HIDDEN"  // specifies whether the window frame and title bar are interactive when the cursor is hidden
	HINT_WINDOWS_ENABLE_MESSAGELOOP               = "SDL_WINDOWS_ENABLE_MESSAGELOOP"               // specifies whether the windows message loop is processed by SDL
	HINT_WINDOWS_NO_CLOSE_ON_ALT_F4               = "SDL_WINDOWS_NO_CLOSE_ON_ALT_F4"               // specifies that SDL should not to generate WINDOWEVENT_CLOSE events for Alt+F4 on Microsoft Windows
	HINT_ANDROID_SEPARATE_MOUSE_AND_TOUCH         = "SDL_ANDROID_SEPARATE_MOUSE_AND_TOUCH"         // specifies a variable to control whether mouse and touch events are to be treated together or separately
	HINT_ANDROID_APK_EXPANSION_MAIN_FILE_VERSION  = "SDL_ANDROID_APK_EXPANSION_MAIN_FILE_VERSION"  // specifies the Android APK expansion main file version
	HINT_ANDROID_APK_EXPANSION_PATCH_FILE_VERSION = "SDL_ANDROID_APK_EXPANSION_PATCH_FILE_VERSION" // specifies the Android APK expansion patch file version
	HINT_AUDIO_RESAMPLING_MODE                    = "SDL_AUDIO_RESAMPLING_MODE"                    // specifies a variable controlling speed/quality tradeoff of audio resampling
	HINT_RENDER_LOGICAL_SIZE_MODE                 = "SDL_RENDER_LOGICAL_SIZE_MODE"                 // specifies a variable controlling the scaling policy for SDL_RenderSetLogicalSize
	HINT_MOUSE_NORMAL_SPEED_SCALE                 = "SDL_MOUSE_NORMAL_SPEED_SCALE"                 // specifies a variable setting the speed scale for mouse motion, in floating point, when the mouse is not in relative mode
	HINT_MOUSE_RELATIVE_SPEED_SCALE               = "SDL_MOUSE_RELATIVE_SPEED_SCALE"               // specifies a variable setting the scale for mouse motion, in floating point, when the mouse is in relative mode
	HINT_TOUCH_MOUSE_EVENTS                       = "SDL_TOUCH_MOUSE_EVENTS"                       // specifies a variable controlling whether touch events should generate synthetic mouse events
	HINT_WINDOWS_INTRESOURCE_ICON                 = "SDL_WINDOWS_INTRESOURCE_ICON"                 // specifies a variable to specify custom icon resource id from RC file on Windows platform
	HINT_WINDOWS_INTRESOURCE_ICON_SMALL           = "SDL_WINDOWS_INTRESOURCE_ICON_SMALL"           // specifies a variable to specify custom icon resource id from RC file on Windows platform
	HINT_IOS_HIDE_HOME_INDICATOR                  = "SDL_IOS_HIDE_HOME_INDICATOR"                  // specifies a variable controlling whether the home indicator bar on iPhone X should be hidden.
	HINT_RETURN_KEY_HIDES_IME                     = "SDL_RETURN_KEY_HIDES_IME"                     // specifies a variable to control whether the return key on the soft keyboard should hide the soft keyboard on Android and iOS.
	HINT_TV_REMOTE_AS_JOYSTICK                    = "SDL_TV_REMOTE_AS_JOYSTICK"                    // specifies a variable controlling whether the Android / tvOS remotes  should be listed as joystick devices, instead of sending keyboard events.
	HINT_VIDEO_X11_NET_WM_BYPASS_COMPOSITOR       = "SDL_VIDEO_X11_NET_WM_BYPASS_COMPOSITOR"       // specifies a variable controlling whether the X11 _NET_WM_BYPASS_COMPOSITOR hint should be used.
	HINT_VIDEO_DOUBLE_BUFFER                      = "SDL_VIDEO_DOUBLE_BUFFER"                      // specifies a variable that tells the video driver that we only want a double buffer.
)

// An enumeration of hint priorities.
// (https://wiki.libsdl.org/SDL_HintPriority)
const (
	HINT_DEFAULT  = iota // low priority, used for default values
	HINT_NORMAL          // medium priority
	HINT_OVERRIDE        // high priority
)

// Hat positions.
// (https://wiki.libsdl.org/SDL_JoystickGetHat)
const (
	HAT_CENTERED  = 0x00
	HAT_UP        = 0x01
	HAT_RIGHT     = 0x02
	HAT_DOWN      = 0x04
	HAT_LEFT      = 0x08
	HAT_RIGHTUP   = HAT_RIGHT | HAT_UP
	HAT_RIGHTDOWN = HAT_RIGHT | HAT_DOWN
	HAT_LEFTUP    = HAT_LEFT | HAT_UP
	HAT_LEFTDOWN  = HAT_LEFT | HAT_DOWN
)

// Types of a joystick.
const (
	JOYSTICK_TYPE_UNKNOWN = iota
	JOYSTICK_TYPE_GAMECONTROLLER
	JOYSTICK_TYPE_WHEEL
	JOYSTICK_TYPE_ARCADE_STICK
	JOYSTICK_TYPE_FLIGHT_STICK
	JOYSTICK_TYPE_DANCE_PAD
	JOYSTICK_TYPE_GUITAR
	JOYSTICK_TYPE_DRUM_KIT
	JOYSTICK_TYPE_ARCADE_PAD
	JOYSTICK_TYPE_THROTTLE
)

// An enumeration of battery levels of a joystick.
// (https://wiki.libsdl.org/SDL_JoystickPowerLevel)
const (
	JOYSTICK_POWER_UNKNOWN = iota - 1
	JOYSTICK_POWER_EMPTY
	JOYSTICK_POWER_LOW
	JOYSTICK_POWER_MEDIUM
	JOYSTICK_POWER_FULL
	JOYSTICK_POWER_WIRED
	JOYSTICK_POWER_MAX
)

// The SDL virtual key representation.
// (https://wiki.libsdl.org/SDL_Keycode)
// (https://wiki.libsdl.org/SDLKeycodeLookup)
const (
	K_UNKNOWN = 0 // "" (no name, empty string)

	K_RETURN     = '\r'   // "Return" (the Enter key (main keyboard))
	K_ESCAPE     = '\033' // "Escape" (the Esc key)
	K_BACKSPACE  = '\b'   // "Backspace"
	K_TAB        = '\t'   // "Tab" (the Tab key)
	K_SPACE      = ' '    // "Space" (the Space Bar key(s))
	K_EXCLAIM    = '!'    // "!"
	K_QUOTEDBL   = '"'    // """
	K_HASH       = '#'    // "#"
	K_PERCENT    = '%'    // "%"
	K_DOLLAR     = '$'    // "$"
	K_AMPERSAND  = '&'    // "&"
	K_QUOTE      = '\''   // "'"
	K_LEFTPAREN  = '('    // "("
	K_RIGHTPAREN = ')'    // ")"
	K_ASTERISK   = '*'    // "*"
	K_PLUS       = '+'    // "+"
	K_COMMA      = ','    // ","
	K_MINUS      = '-'    // "-"
	K_PERIOD     = '.'    // "."
	K_SLASH      = '/'    // "/"
	K_0          = '0'    // "0"
	K_1          = '1'    // "1"
	K_2          = '2'    // "2"
	K_3          = '3'    // "3"
	K_4          = '4'    // "4"
	K_5          = '5'    // "5"
	K_6          = '6'    // "6"
	K_7          = '7'    // "7"
	K_8          = '8'    // "8"
	K_9          = '9'    // "9"
	K_COLON      = ':'    // ":"
	K_SEMICOLON  = ';'    // ";"
	K_LESS       = '<'    // "<"
	K_EQUALS     = '='    // "="
	K_GREATER    = '>'    // ">"
	K_QUESTION   = '?'    // "?"
	K_AT         = '@'    // "@"
	/*
	   Skip uppercase letters
	*/
	K_LEFTBRACKET  = '['  // "["
	K_BACKSLASH    = '\\' // "\"
	K_RIGHTBRACKET = ']'  // "]"
	K_CARET        = '^'  // "^"
	K_UNDERSCORE   = '_'  // "_"
	K_BACKQUOTE    = '`'  // "`"
	K_a            = 'a'  // "A"
	K_b            = 'b'  // "B"
	K_c            = 'c'  // "C"
	K_d            = 'd'  // "D"
	K_e            = 'e'  // "E"
	K_f            = 'f'  // "F"
	K_g            = 'g'  // "G"
	K_h            = 'h'  // "H"
	K_i            = 'i'  // "I"
	K_j            = 'j'  // "J"
	K_k            = 'k'  // "K"
	K_l            = 'l'  // "L"
	K_m            = 'm'  // "M"
	K_n            = 'n'  // "N"
	K_o            = 'o'  // "O"
	K_p            = 'p'  // "P"
	K_q            = 'q'  // "Q"
	K_r            = 'r'  // "R"
	K_s            = 's'  // "S"
	K_t            = 't'  // "T"
	K_u            = 'u'  // "U"
	K_v            = 'v'  // "V"
	K_w            = 'w'  // "W"
	K_x            = 'x'  // "X"
	K_y            = 'y'  // "Y"
	K_z            = 'z'  // "Z"

	K_CAPSLOCK = SCANCODE_CAPSLOCK | K_SCANCODE_MASK // "CapsLock"

	K_F1  = SCANCODE_F1 | K_SCANCODE_MASK  // "F1"
	K_F2  = SCANCODE_F2 | K_SCANCODE_MASK  // "F2"
	K_F3  = SCANCODE_F3 | K_SCANCODE_MASK  // "F3"
	K_F4  = SCANCODE_F4 | K_SCANCODE_MASK  // "F4"
	K_F5  = SCANCODE_F5 | K_SCANCODE_MASK  // "F5"
	K_F6  = SCANCODE_F6 | K_SCANCODE_MASK  // "F6"
	K_F7  = SCANCODE_F7 | K_SCANCODE_MASK  // "F7"
	K_F8  = SCANCODE_F8 | K_SCANCODE_MASK  // "F8"
	K_F9  = SCANCODE_F9 | K_SCANCODE_MASK  // "F9"
	K_F10 = SCANCODE_F10 | K_SCANCODE_MASK // "F10"
	K_F11 = SCANCODE_F11 | K_SCANCODE_MASK // "F11"
	K_F12 = SCANCODE_F12 | K_SCANCODE_MASK // "F12"

	K_PRINTSCREEN = SCANCODE_PRINTSCREEN | K_SCANCODE_MASK // "PrintScreen"
	K_SCROLLLOCK  = SCANCODE_SCROLLLOCK | K_SCANCODE_MASK  // "ScrollLock"
	K_PAUSE       = SCANCODE_PAUSE | K_SCANCODE_MASK       // "Pause" (the Pause / Break key)
	K_INSERT      = SCANCODE_INSERT | K_SCANCODE_MASK      // "Insert" (insert on PC, help on some Mac keyboards (but does send code 73, not 117))
	K_HOME        = SCANCODE_HOME | K_SCANCODE_MASK        // "Home"
	K_PAGEUP      = SCANCODE_PAGEUP | K_SCANCODE_MASK      // "PageUp"
	K_DELETE      = '\177'                                 // "Delete"
	K_END         = SCANCODE_END | K_SCANCODE_MASK         // "End"
	K_PAGEDOWN    = SCANCODE_PAGEDOWN | K_SCANCODE_MASK    // "PageDown"
	K_RIGHT       = SCANCODE_RIGHT | K_SCANCODE_MASK       // "Right" (the Right arrow key (navigation keypad))
	K_LEFT        = SCANCODE_LEFT | K_SCANCODE_MASK        // "Left" (the Left arrow key (navigation keypad))
	K_DOWN        = SCANCODE_DOWN | K_SCANCODE_MASK        // "Down" (the Down arrow key (navigation keypad))
	K_UP          = SCANCODE_UP | K_SCANCODE_MASK          // "Up" (the Up arrow key (navigation keypad))

	K_NUMLOCKCLEAR = SCANCODE_NUMLOCKCLEAR | K_SCANCODE_MASK // "Numlock" (the Num Lock key (PC) / the Clear key (Mac))
	K_KP_DIVIDE    = SCANCODE_KP_DIVIDE | K_SCANCODE_MASK    // "Keypad /" (the / key (numeric keypad))
	K_KP_MULTIPLY  = SCANCODE_KP_MULTIPLY | K_SCANCODE_MASK  // "Keypad *" (the * key (numeric keypad))
	K_KP_MINUS     = SCANCODE_KP_MINUS | K_SCANCODE_MASK     // "Keypad -" (the - key (numeric keypad))
	K_KP_PLUS      = SCANCODE_KP_PLUS | K_SCANCODE_MASK      // "Keypad +" (the + key (numeric keypad))
	K_KP_ENTER     = SCANCODE_KP_ENTER | K_SCANCODE_MASK     // "Keypad Enter" (the Enter key (numeric keypad))
	K_KP_1         = SCANCODE_KP_1 | K_SCANCODE_MASK         // "Keypad 1" (the 1 key (numeric keypad))
	K_KP_2         = SCANCODE_KP_2 | K_SCANCODE_MASK         // "Keypad 2" (the 2 key (numeric keypad))
	K_KP_3         = SCANCODE_KP_3 | K_SCANCODE_MASK         // "Keypad 3" (the 3 key (numeric keypad))
	K_KP_4         = SCANCODE_KP_4 | K_SCANCODE_MASK         // "Keypad 4" (the 4 key (numeric keypad))
	K_KP_5         = SCANCODE_KP_5 | K_SCANCODE_MASK         // "Keypad 5" (the 5 key (numeric keypad))
	K_KP_6         = SCANCODE_KP_6 | K_SCANCODE_MASK         // "Keypad 6" (the 6 key (numeric keypad))
	K_KP_7         = SCANCODE_KP_7 | K_SCANCODE_MASK         // "Keypad 7" (the 7 key (numeric keypad))
	K_KP_8         = SCANCODE_KP_8 | K_SCANCODE_MASK         // "Keypad 8" (the 8 key (numeric keypad))
	K_KP_9         = SCANCODE_KP_9 | K_SCANCODE_MASK         // "Keypad 9" (the 9 key (numeric keypad))
	K_KP_0         = SCANCODE_KP_0 | K_SCANCODE_MASK         // "Keypad 0" (the 0 key (numeric keypad))
	K_KP_PERIOD    = SCANCODE_KP_PERIOD | K_SCANCODE_MASK    // "Keypad ." (the . key (numeric keypad))

	K_APPLICATION    = SCANCODE_APPLICATION | K_SCANCODE_MASK    // "Application" (the Application / Compose / Context Menu (Windows) key)
	K_POWER          = SCANCODE_POWER | K_SCANCODE_MASK          // "Power" (The USB document says this is a status flag, not a physical key - but some Mac keyboards do have a power key.)
	K_KP_EQUALS      = SCANCODE_EQUALS | K_SCANCODE_MASK         // "Keypad =" (the = key (numeric keypad))
	K_F13            = SCANCODE_F13 | K_SCANCODE_MASK            // "F13"
	K_F14            = SCANCODE_F14 | K_SCANCODE_MASK            // "F14"
	K_F15            = SCANCODE_F15 | K_SCANCODE_MASK            // "F15"
	K_F16            = SCANCODE_F16 | K_SCANCODE_MASK            // "F16"
	K_F17            = SCANCODE_F17 | K_SCANCODE_MASK            // "F17"
	K_F18            = SCANCODE_F18 | K_SCANCODE_MASK            // "F18"
	K_F19            = SCANCODE_F19 | K_SCANCODE_MASK            // "F19"
	K_F20            = SCANCODE_F20 | K_SCANCODE_MASK            // "F20"
	K_F21            = SCANCODE_F21 | K_SCANCODE_MASK            // "F21"
	K_F22            = SCANCODE_F22 | K_SCANCODE_MASK            // "F22"
	K_F23            = SCANCODE_F23 | K_SCANCODE_MASK            // "F23"
	K_F24            = SCANCODE_F24 | K_SCANCODE_MASK            // "F24"
	K_EXECUTE        = SCANCODE_EXECUTE | K_SCANCODE_MASK        // "Execute"
	K_HELP           = SCANCODE_HELP | K_SCANCODE_MASK           // "Help"
	K_MENU           = SCANCODE_MENU | K_SCANCODE_MASK           // "Menu"
	K_SELECT         = SCANCODE_SELECT | K_SCANCODE_MASK         // "Select"
	K_STOP           = SCANCODE_STOP | K_SCANCODE_MASK           // "Stop"
	K_AGAIN          = SCANCODE_AGAIN | K_SCANCODE_MASK          // "Again" (the Again key (Redo))
	K_UNDO           = SCANCODE_UNDO | K_SCANCODE_MASK           // "Undo"
	K_CUT            = SCANCODE_CUT | K_SCANCODE_MASK            // "Cut"
	K_COPY           = SCANCODE_COPY | K_SCANCODE_MASK           // "Copy"
	K_PASTE          = SCANCODE_PASTE | K_SCANCODE_MASK          // "Paste"
	K_FIND           = SCANCODE_FIND | K_SCANCODE_MASK           // "Find"
	K_MUTE           = SCANCODE_MUTE | K_SCANCODE_MASK           // "Mute"
	K_VOLUMEUP       = SCANCODE_VOLUMEUP | K_SCANCODE_MASK       // "VolumeUp"
	K_VOLUMEDOWN     = SCANCODE_VOLUMEDOWN | K_SCANCODE_MASK     // "VolumeDown"
	K_KP_COMMA       = SCANCODE_KP_COMMA | K_SCANCODE_MASK       // "Keypad ," (the Comma key (numeric keypad))
	K_KP_EQUALSAS400 = SCANCODE_KP_EQUALSAS400 | K_SCANCODE_MASK // "Keypad = (AS400)" (the Equals AS400 key (numeric keypad))

	K_ALTERASE   = SCANCODE_ALTERASE | K_SCANCODE_MASK   // "AltErase" (Erase-Eaze)
	K_SYSREQ     = SCANCODE_SYSREQ | K_SCANCODE_MASK     // "SysReq" (the SysReq key)
	K_CANCEL     = SCANCODE_CANCEL | K_SCANCODE_MASK     // "Cancel"
	K_CLEAR      = SCANCODE_CLEAR | K_SCANCODE_MASK      // "Clear"
	K_PRIOR      = SCANCODE_PRIOR | K_SCANCODE_MASK      // "Prior"
	K_RETURN2    = SCANCODE_RETURN2 | K_SCANCODE_MASK    // "Return"
	K_SEPARATOR  = SCANCODE_SEPARATOR | K_SCANCODE_MASK  // "Separator"
	K_OUT        = SCANCODE_OUT | K_SCANCODE_MASK        // "Out"
	K_OPER       = SCANCODE_OPER | K_SCANCODE_MASK       // "Oper"
	K_CLEARAGAIN = SCANCODE_CLEARAGAIN | K_SCANCODE_MASK // "Clear / Again"
	K_CRSEL      = SCANCODE_CRSEL | K_SCANCODE_MASK      // "CrSel"
	K_EXSEL      = SCANCODE_EXSEL | K_SCANCODE_MASK      // "ExSel"

	K_KP_00              = SCANCODE_KP_00 | K_SCANCODE_MASK              // "Keypad 00" (the 00 key (numeric keypad))
	K_KP_000             = SCANCODE_KP_000 | K_SCANCODE_MASK             // "Keypad 000" (the 000 key (numeric keypad))
	K_THOUSANDSSEPARATOR = SCANCODE_THOUSANDSSEPARATOR | K_SCANCODE_MASK // "ThousandsSeparator" (the Thousands Separator key)
	K_DECIMALSEPARATOR   = SCANCODE_DECIMALSEPARATOR | K_SCANCODE_MASK   // "DecimalSeparator" (the Decimal Separator key)
	K_CURRENCYUNIT       = SCANCODE_CURRENCYUNIT | K_SCANCODE_MASK       // "CurrencyUnit" (the Currency Unit key)
	K_CURRENCYSUBUNIT    = SCANCODE_CURRENCYSUBUNIT | K_SCANCODE_MASK    // "CurrencySubUnit" (the Currency Subunit key)
	K_KP_LEFTPAREN       = SCANCODE_KP_LEFTPAREN | K_SCANCODE_MASK       // "Keypad (" (the Left Parenthesis key (numeric keypad))
	K_KP_RIGHTPAREN      = SCANCODE_KP_RIGHTPAREN | K_SCANCODE_MASK      // "Keypad )" (the Right Parenthesis key (numeric keypad))
	K_KP_LEFTBRACE       = SCANCODE_KP_LEFTBRACE | K_SCANCODE_MASK       // "Keypad {" (the Left Brace key (numeric keypad))
	K_KP_RIGHTBRACE      = SCANCODE_KP_RIGHTBRACE | K_SCANCODE_MASK      // "Keypad }" (the Right Brace key (numeric keypad))
	K_KP_TAB             = SCANCODE_KP_TAB | K_SCANCODE_MASK             // "Keypad Tab" (the Tab key (numeric keypad))
	K_KP_BACKSPACE       = SCANCODE_KP_BACKSPACE | K_SCANCODE_MASK       // "Keypad Backspace" (the Backspace key (numeric keypad))
	K_KP_A               = SCANCODE_KP_A | K_SCANCODE_MASK               // "Keypad A" (the A key (numeric keypad))
	K_KP_B               = SCANCODE_KP_B | K_SCANCODE_MASK               // "Keypad B" (the B key (numeric keypad))
	K_KP_C               = SCANCODE_KP_C | K_SCANCODE_MASK               // "Keypad C" (the C key (numeric keypad))
	K_KP_D               = SCANCODE_KP_D | K_SCANCODE_MASK               // "Keypad D" (the D key (numeric keypad))
	K_KP_E               = SCANCODE_KP_E | K_SCANCODE_MASK               // "Keypad E" (the E key (numeric keypad))
	K_KP_F               = SCANCODE_KP_F | K_SCANCODE_MASK               // "Keypad F" (the F key (numeric keypad))
	K_KP_XOR             = SCANCODE_KP_XOR | K_SCANCODE_MASK             // "Keypad XOR" (the XOR key (numeric keypad))
	K_KP_POWER           = SCANCODE_KP_POWER | K_SCANCODE_MASK           // "Keypad ^" (the Power key (numeric keypad))
	K_KP_PERCENT         = SCANCODE_KP_PERCENT | K_SCANCODE_MASK         // "Keypad %" (the Percent key (numeric keypad))
	K_KP_LESS            = SCANCODE_KP_LESS | K_SCANCODE_MASK            // "Keypad <" (the Less key (numeric keypad))
	K_KP_GREATER         = SCANCODE_KP_GREATER | K_SCANCODE_MASK         // "Keypad >" (the Greater key (numeric keypad))
	K_KP_AMPERSAND       = SCANCODE_KP_AMPERSAND | K_SCANCODE_MASK       // "Keypad &" (the & key (numeric keypad))
	K_KP_DBLAMPERSAND    = SCANCODE_KP_DBLAMPERSAND | K_SCANCODE_MASK    // "Keypad &&" (the && key (numeric keypad))
	K_KP_VERTICALBAR     = SCANCODE_KP_VERTICALBAR | K_SCANCODE_MASK     // "Keypad |" (the | key (numeric keypad))
	K_KP_DBLVERTICALBAR  = SCANCODE_KP_DBLVERTICALBAR | K_SCANCODE_MASK  // "Keypad ||" (the || key (numeric keypad))
	K_KP_COLON           = SCANCODE_KP_COLON | K_SCANCODE_MASK           // "Keypad :" (the : key (numeric keypad))
	K_KP_HASH            = SCANCODE_KP_HASH | K_SCANCODE_MASK            // "Keypad #" (the # key (numeric keypad))
	K_KP_SPACE           = SCANCODE_KP_SPACE | K_SCANCODE_MASK           // "Keypad Space" (the Space key (numeric keypad))
	K_KP_AT              = SCANCODE_KP_AT | K_SCANCODE_MASK              // "Keypad @" (the @ key (numeric keypad))
	K_KP_EXCLAM          = SCANCODE_KP_EXCLAM | K_SCANCODE_MASK          // "Keypad !" (the ! key (numeric keypad))
	K_KP_MEMSTORE        = SCANCODE_KP_MEMSTORE | K_SCANCODE_MASK        // "Keypad MemStore" (the Mem Store key (numeric keypad))
	K_KP_MEMRECALL       = SCANCODE_KP_MEMRECALL | K_SCANCODE_MASK       // "Keypad MemRecall" (the Mem Recall key (numeric keypad))
	K_KP_MEMCLEAR        = SCANCODE_KP_MEMCLEAR | K_SCANCODE_MASK        // "Keypad MemClear" (the Mem Clear key (numeric keypad))
	K_KP_MEMADD          = SCANCODE_KP_MEMADD | K_SCANCODE_MASK          // "Keypad MemAdd" (the Mem Add key (numeric keypad))
	K_KP_MEMSUBTRACT     = SCANCODE_KP_MEMSUBTRACT | K_SCANCODE_MASK     // "Keypad MemSubtract" (the Mem Subtract key (numeric keypad))
	K_KP_MEMMULTIPLY     = SCANCODE_KP_MEMMULTIPLY | K_SCANCODE_MASK     // "Keypad MemMultiply" (the Mem Multiply key (numeric keypad))
	K_KP_MEMDIVIDE       = SCANCODE_KP_MEMDIVIDE | K_SCANCODE_MASK       // "Keypad MemDivide" (the Mem Divide key (numeric keypad))
	K_KP_PLUSMINUS       = SCANCODE_KP_PLUSMINUS | K_SCANCODE_MASK       // "Keypad +/-" (the +/- key (numeric keypad))
	K_KP_CLEAR           = SCANCODE_KP_CLEAR | K_SCANCODE_MASK           // "Keypad Clear" (the Clear key (numeric keypad))
	K_KP_CLEARENTRY      = SCANCODE_KP_CLEARENTRY | K_SCANCODE_MASK      // "Keypad ClearEntry" (the Clear Entry key (numeric keypad))
	K_KP_BINARY          = SCANCODE_KP_BINARY | K_SCANCODE_MASK          // "Keypad Binary" (the Binary key (numeric keypad))
	K_KP_OCTAL           = SCANCODE_KP_OCTAL | K_SCANCODE_MASK           // "Keypad Octal" (the Octal key (numeric keypad))
	K_KP_DECIMAL         = SCANCODE_KP_DECIMAL | K_SCANCODE_MASK         // "Keypad Decimal" (the Decimal key (numeric keypad))
	K_KP_HEXADECIMAL     = SCANCODE_KP_HEXADECIMAL | K_SCANCODE_MASK     // "Keypad Hexadecimal" (the Hexadecimal key (numeric keypad))

	K_LCTRL  = SCANCODE_LCTRL | K_SCANCODE_MASK  // "Left Ctrl"
	K_LSHIFT = SCANCODE_LSHIFT | K_SCANCODE_MASK // "Left Shift"
	K_LALT   = SCANCODE_LALT | K_SCANCODE_MASK   // "Left Alt" (alt, option)
	K_LGUI   = SCANCODE_LGUI | K_SCANCODE_MASK   // "Left GUI" (windows, command (apple), meta)
	K_RCTRL  = SCANCODE_RCTRL | K_SCANCODE_MASK  // "Right Ctrl"
	K_RSHIFT = SCANCODE_RSHIFT | K_SCANCODE_MASK // "Right Shift"
	K_RALT   = SCANCODE_RALT | K_SCANCODE_MASK   // "Right Alt" (alt, option)
	K_RGUI   = SCANCODE_RGUI | K_SCANCODE_MASK   // "Right GUI" (windows, command (apple), meta)

	K_MODE = SCANCODE_MODE | K_SCANCODE_MASK // "ModeSwitch" (I'm not sure if this is really not covered by any of the above, but since there's a special KMOD_MODE for it I'm adding it here)

	K_AUDIONEXT    = SCANCODE_AUDIONEXT | K_SCANCODE_MASK    // "AudioNext" (the Next Track media key)
	K_AUDIOPREV    = SCANCODE_AUDIOPREV | K_SCANCODE_MASK    // "AudioPrev" (the Previous Track media key)
	K_AUDIOSTOP    = SCANCODE_AUDIOSTOP | K_SCANCODE_MASK    // "AudioStop" (the Stop media key)
	K_AUDIOPLAY    = SCANCODE_AUDIOPLAY | K_SCANCODE_MASK    // "AudioPlay" (the Play media key)
	K_AUDIOMUTE    = SCANCODE_AUDIOMUTE | K_SCANCODE_MASK    // "AudioMute" (the Mute volume key)
	K_MEDIASELECT  = SCANCODE_MEDIASELECT | K_SCANCODE_MASK  // "MediaSelect" (the Media Select key)
	K_WWW          = SCANCODE_WWW | K_SCANCODE_MASK          // "WWW" (the WWW/World Wide Web key)
	K_MAIL         = SCANCODE_MAIL | K_SCANCODE_MASK         // "Mail" (the Mail/eMail key)
	K_CALCULATOR   = SCANCODE_CALCULATOR | K_SCANCODE_MASK   // "Calculator" (the Calculator key)
	K_COMPUTER     = SCANCODE_COMPUTER | K_SCANCODE_MASK     // "Computer" (the My Computer key)
	K_AC_SEARCH    = SCANCODE_AC_SEARCH | K_SCANCODE_MASK    // "AC Search" (the Search key (application control keypad))
	K_AC_HOME      = SCANCODE_AC_HOME | K_SCANCODE_MASK      // "AC Home" (the Home key (application control keypad))
	K_AC_BACK      = SCANCODE_AC_BACK | K_SCANCODE_MASK      // "AC Back" (the Back key (application control keypad))
	K_AC_FORWARD   = SCANCODE_AC_FORWARD | K_SCANCODE_MASK   // "AC Forward" (the Forward key (application control keypad))
	K_AC_STOP      = SCANCODE_AC_STOP | K_SCANCODE_MASK      // "AC Stop" (the Stop key (application control keypad))
	K_AC_REFRESH   = SCANCODE_AC_REFRESH | K_SCANCODE_MASK   // "AC Refresh" (the Refresh key (application control keypad))
	K_AC_BOOKMARKS = SCANCODE_AC_BOOKMARKS | K_SCANCODE_MASK // "AC Bookmarks" (the Bookmarks key (application control keypad))

	K_BRIGHTNESSDOWN = SCANCODE_BRIGHTNESSDOWN | K_SCANCODE_MASK // "BrightnessDown" (the Brightness Down key)
	K_BRIGHTNESSUP   = SCANCODE_BRIGHTNESSUP | K_SCANCODE_MASK   // "BrightnessUp" (the Brightness Up key)
	K_DISPLAYSWITCH  = SCANCODE_DISPLAYSWITCH | K_SCANCODE_MASK  // "DisplaySwitch" (display mirroring/dual display switch, video mode switch)
	K_KBDILLUMTOGGLE = SCANCODE_KBDILLUMTOGGLE | K_SCANCODE_MASK // "KBDIllumToggle" (the Keyboard Illumination Toggle key)
	K_KBDILLUMDOWN   = SCANCODE_KBDILLUMDOWN | K_SCANCODE_MASK   // "KBDIllumDown" (the Keyboard Illumination Down key)
	K_KBDILLUMUP     = SCANCODE_KBDILLUMUP | K_SCANCODE_MASK     // "KBDIllumUp" (the Keyboard Illumination Up key)
	K_EJECT          = SCANCODE_EJECT | K_SCANCODE_MASK          // "Eject" (the Eject key)
	K_SLEEP          = SCANCODE_SLEEP | K_SCANCODE_MASK          // "Sleep" (the Sleep key)
	K_APP1           = SCANCODE_APP1 | K_SCANCODE_MASK
	K_APP2           = SCANCODE_APP2 | K_SCANCODE_MASK

	K_AUDIOREWIND      = SCANCODE_AUDIOREWIND | K_SCANCODE_MASK
	K_AUDIOFASTFORWARD = SCANCODE_AUDIOFASTFORWARD | K_SCANCODE_MASK
)

// An enumeration of key modifier masks.
// (https://wiki.libsdl.org/SDL_Keymod)
const (
	KMOD_NONE     = 0x0000                    // 0 (no modifier is applicable)
	KMOD_LSHIFT   = 0x0001                    // the left Shift key is down
	KMOD_RSHIFT   = 0x0002                    // the right Shift key is down
	KMOD_LCTRL    = 0x0040                    // the left Ctrl (Control) key is down
	KMOD_RCTRL    = 0x0080                    // the right Ctrl (Control) key is down
	KMOD_LALT     = 0x0100                    // the left Alt key is down
	KMOD_RALT     = 0x0200                    // the right Alt key is down
	KMOD_LGUI     = 0x0400                    // the left GUI key (often the Windows key) is down
	KMOD_RGUI     = 0x0800                    // the right GUI key (often the Windows key) is down
	KMOD_NUM      = 0x1000                    // the Num Lock key (may be located on an extended keypad) is down
	KMOD_CAPS     = 0x2000                    // the Caps Lock key is down
	KMOD_MODE     = 0x4000                    // the AltGr key is down
	KMOD_CTRL     = KMOD_LCTRL | KMOD_RCTRL   // (KMOD_LCTRL|KMOD_RCTRL)
	KMOD_SHIFT    = KMOD_LSHIFT | KMOD_RSHIFT // (KMOD_LSHIFT|KMOD_RSHIFT)
	KMOD_ALT      = KMOD_LALT | KMOD_RALT     // (KMOD_LALT|KMOD_RALT)
	KMOD_GUI      = KMOD_LGUI | KMOD_RGUI     // (KMOD_LGUI|KMOD_RGUI)
	KMOD_RESERVED = 0x8000                    // reserved for future use
)

// An enumeration of the predefined log categories.
// (https://wiki.libsdl.org/SDL_LOG_CATEGORY)
const (
	LOG_CATEGORY_APPLICATION = iota // application log
	LOG_CATEGORY_ERROR              // error log
	LOG_CATEGORY_ASSERT             // assert log
	LOG_CATEGORY_SYSTEM             // system log
	LOG_CATEGORY_AUDIO              // audio log
	LOG_CATEGORY_VIDEO              // video log
	LOG_CATEGORY_RENDER             // render log
	LOG_CATEGORY_INPUT              // input log
	LOG_CATEGORY_TEST               // test log
	LOG_CATEGORY_RESERVED1          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED2          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED3          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED4          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED5          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED6          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED7          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED8          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED9          // reserved for future SDL library use
	LOG_CATEGORY_RESERVED10         // reserved for future SDL library use
	LOG_CATEGORY_CUSTOM             // reserved for application use
)

// An enumeration of the predefined log priorities.
// (https://wiki.libsdl.org/SDL_LogPriority)
const (
	LOG_PRIORITY_VERBOSE  = iota + 1 // verbose
	LOG_PRIORITY_DEBUG               // debug
	LOG_PRIORITY_INFO                // info
	LOG_PRIORITY_WARN                // warn
	LOG_PRIORITY_ERROR               // error
	LOG_PRIORITY_CRITICAL            // critical
	NUM_LOG_PRIORITIES               // (internal use)
)

// Cursor types for CreateSystemCursor()
const (
	SYSTEM_CURSOR_ARROW     = iota // arrow
	SYSTEM_CURSOR_IBEAM            // i-beam
	SYSTEM_CURSOR_WAIT             // wait
	SYSTEM_CURSOR_CROSSHAIR        // crosshair
	SYSTEM_CURSOR_WAITARROW        // small wait cursor (or wait if not available)
	SYSTEM_CURSOR_SIZENWSE         // double arrow pointing northwest and southeast
	SYSTEM_CURSOR_SIZENESW         // double arrow pointing northeast and southwest
	SYSTEM_CURSOR_SIZEWE           // double arrow pointing west and east
	SYSTEM_CURSOR_SIZENS           // double arrow pointing north and south
	SYSTEM_CURSOR_SIZEALL          // four pointed arrow pointing north, south, east, and west
	SYSTEM_CURSOR_NO               // slashed circle or crossbones
	SYSTEM_CURSOR_HAND             // hand
	NUM_SYSTEM_CURSORS             // (only for bounding internal arrays)
)

// Scroll direction types for the Scroll event
const (
	MOUSEWHEEL_NORMAL  = iota // the scroll direction is normal
	MOUSEWHEEL_FLIPPED        // the scroll direction is flipped / natural
)

// Used as a mask when testing buttons in buttonstate.
const (
	BUTTON_LEFT   = 1 // left mouse button
	BUTTON_MIDDLE = 2 // middle mouse button
	BUTTON_RIGHT  = 3 // right mouse button
	BUTTON_X1     = 4 // x1 mouse button
	BUTTON_X2     = 5 // x2 mouse button
)

// Pixel types.
const (
	PIXELTYPE_UNKNOWN = iota
	PIXELTYPE_INDEX1
	PIXELTYPE_INDEX4
	PIXELTYPE_INDEX8
	PIXELTYPE_PACKED8
	PIXELTYPE_PACKED16
	PIXELTYPE_PACKED32
	PIXELTYPE_ARRAYU8
	PIXELTYPE_ARRAYU16
	PIXELTYPE_ARRAYU32
	PIXELTYPE_ARRAYF16
	PIXELTYPE_ARRAYF32
)

// Bitmap pixel order high bit -> low bit.
const (
	BITMAPORDER_NONE = iota
	BITMAPORDER_4321
	BITMAPORDER_1234
)

// Packed component order high bit -> low bit.
const (
	PACKEDORDER_NONE = iota
	PACKEDORDER_XRGB
	PACKEDORDER_RGBX
	PACKEDORDER_ARGB
	PACKEDORDER_RGBA
	PACKEDORDER_XBGR
	PACKEDORDER_BGRX
	PACKEDORDER_ABGR
	PACKEDORDER_BGRA
)

// Array component order low byte -> high byte.
const (
	ARRAYORDER_NONE = iota
	ARRAYORDER_RGB
	ARRAYORDER_RGBA
	ARRAYORDER_ARGB
	ARRAYORDER_BGR
	ARRAYORDER_BGRA
	ARRAYORDER_ABGR
)

// Packed component layout.
const (
	PACKEDLAYOUT_NONE = iota
	PACKEDLAYOUT_332
	PACKEDLAYOUT_4444
	PACKEDLAYOUT_1555
	PACKEDLAYOUT_5551
	PACKEDLAYOUT_565
	PACKEDLAYOUT_8888
	PACKEDLAYOUT_2101010
	PACKEDLAYOUT_1010102
)

// Pixel format values.
const (
	PIXELFORMAT_UNKNOWN      = 0
	PIXELFORMAT_INDEX1LSB    = (1 << 28) | ((PIXELTYPE_INDEX1) << 24) | ((BITMAPORDER_4321) << 20) | ((0) << 16) | ((1) << 8) | ((0) << 0)
	PIXELFORMAT_INDEX1MSB    = (1 << 28) | ((PIXELTYPE_INDEX1) << 24) | ((BITMAPORDER_1234) << 20) | ((0) << 16) | ((1) << 8) | ((0) << 0)
	PIXELFORMAT_INDEX4LSB    = (1 << 28) | ((PIXELTYPE_INDEX4) << 24) | ((BITMAPORDER_4321) << 20) | ((0) << 16) | ((4) << 8) | ((0) << 0)
	PIXELFORMAT_INDEX4MSB    = (1 << 28) | ((PIXELTYPE_INDEX4) << 24) | ((BITMAPORDER_1234) << 20) | ((0) << 16) | ((4) << 8) | ((0) << 0)
	PIXELFORMAT_INDEX8       = (1 << 28) | ((PIXELTYPE_INDEX8) << 24) | ((0) << 20) | ((0) << 16) | ((8) << 8) | ((1) << 0)
	PIXELFORMAT_RGB332       = (1 << 28) | ((PIXELTYPE_PACKED8) << 24) | ((PACKEDORDER_XRGB) << 20) | ((PACKEDLAYOUT_332) << 16) | ((8) << 8) | ((1) << 0)
	PIXELFORMAT_RGB444       = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_XRGB) << 20) | ((PACKEDLAYOUT_4444) << 16) | ((12) << 8) | ((2) << 0)
	PIXELFORMAT_RGB555       = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_XRGB) << 20) | ((PACKEDLAYOUT_1555) << 16) | ((15) << 8) | ((2) << 0)
	PIXELFORMAT_BGR555       = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_XBGR) << 20) | ((PACKEDLAYOUT_1555) << 16) | ((15) << 8) | ((2) << 0)
	PIXELFORMAT_ARGB4444     = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_ARGB) << 20) | ((PACKEDLAYOUT_4444) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_RGBA4444     = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_RGBA) << 20) | ((PACKEDLAYOUT_4444) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_ABGR4444     = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_ABGR) << 20) | ((PACKEDLAYOUT_4444) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_BGRA4444     = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_BGRA) << 20) | ((PACKEDLAYOUT_4444) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_ARGB1555     = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_ARGB) << 20) | ((PACKEDLAYOUT_1555) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_RGBA5551     = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_RGBA) << 20) | ((PACKEDLAYOUT_5551) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_ABGR1555     = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_ABGR) << 20) | ((PACKEDLAYOUT_1555) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_BGRA5551     = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_BGRA) << 20) | ((PACKEDLAYOUT_5551) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_RGB565       = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_XRGB) << 20) | ((PACKEDLAYOUT_565) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_BGR565       = (1 << 28) | ((PIXELTYPE_PACKED16) << 24) | ((PACKEDORDER_XBGR) << 20) | ((PACKEDLAYOUT_565) << 16) | ((16) << 8) | ((2) << 0)
	PIXELFORMAT_RGB24        = (1 << 28) | ((PIXELTYPE_ARRAYU8) << 24) | ((ARRAYORDER_RGB) << 20) | ((0) << 16) | ((24) << 8) | ((3) << 0)
	PIXELFORMAT_BGR24        = (1 << 28) | ((PIXELTYPE_ARRAYU8) << 24) | ((ARRAYORDER_BGR) << 20) | ((0) << 16) | ((24) << 8) | ((3) << 0)
	PIXELFORMAT_RGB888       = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_XRGB) << 20) | ((PACKEDLAYOUT_8888) << 16) | ((24) << 8) | ((4) << 0)
	PIXELFORMAT_RGBX8888     = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_RGBX) << 20) | ((PACKEDLAYOUT_8888) << 16) | ((24) << 8) | ((4) << 0)
	PIXELFORMAT_BGR888       = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_XBGR) << 20) | ((PACKEDLAYOUT_8888) << 16) | ((24) << 8) | ((4) << 0)
	PIXELFORMAT_BGRX8888     = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_BGRX) << 20) | ((PACKEDLAYOUT_8888) << 16) | ((24) << 8) | ((4) << 0)
	PIXELFORMAT_ARGB8888     = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_ARGB) << 20) | ((PACKEDLAYOUT_8888) << 16) | ((32) << 8) | ((4) << 0)
	PIXELFORMAT_RGBA8888     = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_RGBA) << 20) | ((PACKEDLAYOUT_8888) << 16) | ((32) << 8) | ((4) << 0)
	PIXELFORMAT_ABGR8888     = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_ABGR) << 20) | ((PACKEDLAYOUT_8888) << 16) | ((32) << 8) | ((4) << 0)
	PIXELFORMAT_BGRA8888     = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_BGRA) << 20) | ((PACKEDLAYOUT_8888) << 16) | ((32) << 8) | ((4) << 0)
	PIXELFORMAT_ARGB2101010  = (1 << 28) | ((PIXELTYPE_PACKED32) << 24) | ((PACKEDORDER_ARGB) << 20) | ((PACKEDLAYOUT_2101010) << 16) | ((32) << 8) | ((4) << 0)
	PIXELFORMAT_YV12         = ('Y' << 0) | ('V' << 8) | ('1' << 16) | ('2' << 24)
	PIXELFORMAT_IYUV         = ('I' << 0) | ('Y' << 8) | ('U' << 16) | ('V' << 24)
	PIXELFORMAT_YUY2         = ('Y' << 0) | ('U' << 8) | ('Y' << 16) | ('2' << 24)
	PIXELFORMAT_UYVY         = ('U' << 0) | ('Y' << 8) | ('V' << 16) | ('Y' << 24)
	PIXELFORMAT_YVYU         = ('Y' << 0) | ('V' << 8) | ('Y' << 16) | ('U' << 24)
	PIXELFORMAT_NV12         = ('N' << 0) | ('V' << 8) | ('1' << 16) | ('2' << 24)
	PIXELFORMAT_NV21         = ('N' << 0) | ('V' << 8) | ('2' << 16) | ('1' << 24)
	PIXELFORMAT_EXTERNAL_OES = ('O' << 0) | ('E' << 8) | ('S' << 16) | (' ' << 24)
)

// These define alpha as the opacity of a surface.
const (
	ALPHA_OPAQUE      = 255
	ALPHA_TRANSPARENT = 0
)

// An enumeration of the basic state of the system's power supply.
// (https://wiki.libsdl.org/SDL_PowerState)
const (
	POWERSTATE_UNKNOWN    = iota // cannot determine power status
	POWERSTATE_ON_BATTERY        // not plugged in, running on the battery
	POWERSTATE_NO_BATTERY        // plugged in, no battery available
	POWERSTATE_CHARGING          // plugged in, charging battery
	POWERSTATE_CHARGED           // plugged in, battery charged
)

// An enumeration of flags used when creating a rendering context.
// (https://wiki.libsdl.org/SDL_RendererFlags)
const (
	RENDERER_SOFTWARE      = 0x00000001 // the renderer is a software fallback
	RENDERER_ACCELERATED   = 0x00000002 // the renderer uses hardware acceleration
	RENDERER_PRESENTVSYNC  = 0x00000004 // present is synchronized with the refresh rate
	RENDERER_TARGETTEXTURE = 0x00000008 // the renderer supports rendering to texture
)

// An enumeration of texture access patterns..
// (https://wiki.libsdl.org/SDL_TextureAccess)
const (
	TEXTUREACCESS_STATIC    = iota // changes rarely, not lockable
	TEXTUREACCESS_STREAMING        // changes frequently, lockable
	TEXTUREACCESS_TARGET           // can be used as a render target
)

// An enumeration of the texture channel modulation used in Renderer.Copy().
// (https://wiki.libsdl.org/SDL_TextureModulate)
const (
	TEXTUREMODULATE_NONE  = 0x00000000 // no modulation
	TEXTUREMODULATE_COLOR = 0x00000001 // srcC = srcC * color
	TEXTUREMODULATE_ALPHA = 0x00000002 // srcA = srcA * alpha
)

// An enumeration of flags that can be used in the flip parameter for Renderer.CopyEx().
// (https://wiki.libsdl.org/SDL_RendererFlip)
const (
	FLIP_NONE       RendererFlip = 0x00000000 // do not flip
	FLIP_HORIZONTAL              = 0x00000001 // flip horizontally
	FLIP_VERTICAL                = 0x00000002 // flip vertically
)

// RWops types
const (
	RWOPS_UNKNOWN   = 0 // unknown stream type
	RWOPS_WINFILE   = 1 // win32 file
	RWOPS_STDFILE   = 2 // stdio file
	RWOPS_JNIFILE   = 3 // android asset
	RWOPS_MEMORY    = 4 // memory stream
	RWOPS_MEMORY_RO = 5 // read-only memory stream
)

// RWops seek from
const (
	RW_SEEK_SET = 0 // seek from the beginning of data
	RW_SEEK_CUR = 1 // seek relative to current read point
	RW_SEEK_END = 2 // seek relative to the end of data
)

// The SDL keyboard scancode representation.
// (https://wiki.libsdl.org/SDL_Scancode)
// (https://wiki.libsdl.org/SDLScancodeLookup)
const (
	SCANCODE_UNKNOWN = 0 // "" (no name, empty string)

	SCANCODE_A = 4  // "A"
	SCANCODE_B = 5  // "B"
	SCANCODE_C = 6  // "C"
	SCANCODE_D = 7  // "D"
	SCANCODE_E = 8  // "E"
	SCANCODE_F = 9  // "F"
	SCANCODE_G = 10 // "G"
	SCANCODE_H = 11 // "H"
	SCANCODE_I = 12 // "I"
	SCANCODE_J = 13 // "J"
	SCANCODE_K = 14 // "K"
	SCANCODE_L = 15 // "L"
	SCANCODE_M = 16 // "M"
	SCANCODE_N = 17 // "N"
	SCANCODE_O = 18 // "O"
	SCANCODE_P = 19 // "P"
	SCANCODE_Q = 20 // "Q"
	SCANCODE_R = 21 // "R"
	SCANCODE_S = 22 // "S"
	SCANCODE_T = 23 // "T"
	SCANCODE_U = 24 // "U"
	SCANCODE_V = 25 // "V"
	SCANCODE_W = 26 // "W"
	SCANCODE_X = 27 // "X"
	SCANCODE_Y = 28 // "Y"
	SCANCODE_Z = 29 // "Z"

	SCANCODE_1 = 30 // "1"
	SCANCODE_2 = 31 // "2"
	SCANCODE_3 = 32 // "3"
	SCANCODE_4 = 33 // "4"
	SCANCODE_5 = 34 // "5"
	SCANCODE_6 = 35 // "6"
	SCANCODE_7 = 36 // "7"
	SCANCODE_8 = 37 // "8"
	SCANCODE_9 = 38 // "9"
	SCANCODE_0 = 39 // "0"

	SCANCODE_RETURN    = 40 // "Return"
	SCANCODE_ESCAPE    = 41 // "Escape" (the Esc key)
	SCANCODE_BACKSPACE = 42 // "Backspace"
	SCANCODE_TAB       = 43 // "Tab" (the Tab key)
	SCANCODE_SPACE     = 44 // "Space" (the Space Bar key(s))

	SCANCODE_MINUS        = 45 // "-"
	SCANCODE_EQUALS       = 46 // "="
	SCANCODE_LEFTBRACKET  = 47 // "["
	SCANCODE_RIGHTBRACKET = 48 // "]"
	SCANCODE_BACKSLASH    = 49 // "\"
	SCANCODE_NONUSHASH    = 50 // "#" (ISO USB keyboards actually use this code instead of 49 for the same key, but all OSes I've seen treat the two codes identically. So, as an implementor, unless your keyboard generates both of those codes and your OS treats them differently, you should generate SDL_SCANCODE_BACKSLASH instead of this code. As a user, you should not rely on this code because SDL will never generate it with most (all?) keyboards.)
	SCANCODE_SEMICOLON    = 51 // ";"
	SCANCODE_APOSTROPHE   = 52 // "'"
	SCANCODE_GRAVE        = 53 // "`"
	SCANCODE_COMMA        = 54 // ","
	SCANCODE_PERIOD       = 55 // "."
	SCANCODE_SLASH        = 56 // "/"
	SCANCODE_CAPSLOCK     = 57 // "CapsLock"
	SCANCODE_F1           = 58 // "F1"
	SCANCODE_F2           = 59 // "F2"
	SCANCODE_F3           = 60 // "F3"
	SCANCODE_F4           = 61 // "F4"
	SCANCODE_F5           = 62 // "F5"
	SCANCODE_F6           = 63 // "F6"
	SCANCODE_F7           = 64 // "F7"
	SCANCODE_F8           = 65 // "F8"
	SCANCODE_F9           = 66 // "F9"
	SCANCODE_F10          = 67 // "F10"
	SCANCODE_F11          = 68 // "F11"
	SCANCODE_F12          = 69 // "F12"
	SCANCODE_PRINTSCREEN  = 70 // "PrintScreen"
	SCANCODE_SCROLLLOCK   = 71 // "ScrollLock"
	SCANCODE_PAUSE        = 72 // "Pause" (the Pause / Break key)
	SCANCODE_INSERT       = 73 // "Insert" (insert on PC, help on some Mac keyboards (but does send code 73, not 117))
	SCANCODE_HOME         = 74 // "Home"
	SCANCODE_PAGEUP       = 75 // "PageUp"
	SCANCODE_DELETE       = 76 // "Delete"
	SCANCODE_END          = 77 // "End"
	SCANCODE_PAGEDOWN     = 78 // "PageDown"
	SCANCODE_RIGHT        = 79 // "Right" (the Right arrow key (navigation keypad))
	SCANCODE_LEFT         = 80 // "Left" (the Left arrow key (navigation keypad))
	SCANCODE_DOWN         = 81 // "Down" (the Down arrow key (navigation keypad))
	SCANCODE_UP           = 82 // "Up" (the Up arrow key (navigation keypad))

	SCANCODE_NUMLOCKCLEAR = 83 // "Numlock" (the Num Lock key (PC) / the Clear key (Mac))
	SCANCODE_KP_DIVIDE    = 84 // "Keypad /" (the / key (numeric keypad))
	SCANCODE_KP_MULTIPLY  = 85 // "Keypad *" (the * key (numeric keypad))
	SCANCODE_KP_MINUS     = 86 // "Keypad -" (the - key (numeric keypad))
	SCANCODE_KP_PLUS      = 87 // "Keypad +" (the + key (numeric keypad))
	SCANCODE_KP_ENTER     = 88 // "Keypad Enter" (the Enter key (numeric keypad))
	SCANCODE_KP_1         = 89 // "Keypad 1" (the 1 key (numeric keypad))
	SCANCODE_KP_2         = 90 // "Keypad 2" (the 2 key (numeric keypad))
	SCANCODE_KP_3         = 91 // "Keypad 3" (the 3 key (numeric keypad))
	SCANCODE_KP_4         = 92 // "Keypad 4" (the 4 key (numeric keypad))
	SCANCODE_KP_5         = 93 // "Keypad 5" (the 5 key (numeric keypad))
	SCANCODE_KP_6         = 94 // "Keypad 6" (the 6 key (numeric keypad))
	SCANCODE_KP_7         = 95 // "Keypad 7" (the 7 key (numeric keypad))
	SCANCODE_KP_8         = 96 // "Keypad 8" (the 8 key (numeric keypad))
	SCANCODE_KP_9         = 97 // "Keypad 9" (the 9 key (numeric keypad))
	SCANCODE_KP_0         = 98 // "Keypad 0" (the 0 key (numeric keypad))
	SCANCODE_KP_PERIOD    = 99 // "Keypad ." (the . key (numeric keypad))

	SCANCODE_NONUSBACKSLASH = 100 // "" (no name, empty string; This is the additional key that ISO keyboards have over ANSI ones, located between left shift and Y. Produces GRAVE ACCENT and TILDE in a US or UK Mac layout, REVERSE SOLIDUS (backslash) and VERTICAL LINE in a US or UK Windows layout, and LESS-THAN SIGN and GREATER-THAN SIGN in a Swiss German, German, or French layout.)
	SCANCODE_APPLICATION    = 101 // "Application" (the Application / Compose / Context Menu (Windows) key)
	SCANCODE_POWER          = 102 // "Power" (The USB document says this is a status flag, not a physical key - but some Mac keyboards do have a power key.)
	SCANCODE_KP_EQUALS      = 103 // "Keypad =" (the = key (numeric keypad))
	SCANCODE_F13            = 104 // "F13"
	SCANCODE_F14            = 105 // "F14"
	SCANCODE_F15            = 106 // "F15"
	SCANCODE_F16            = 107 // "F16"
	SCANCODE_F17            = 108 // "F17"
	SCANCODE_F18            = 109 // "F18"
	SCANCODE_F19            = 110 // "F19"
	SCANCODE_F20            = 111 // "F20"
	SCANCODE_F21            = 112 // "F21"
	SCANCODE_F22            = 113 // "F22"
	SCANCODE_F23            = 114 // "F23"
	SCANCODE_F24            = 115 // "F24"
	SCANCODE_EXECUTE        = 116 // "Execute"
	SCANCODE_HELP           = 117 // "Help"
	SCANCODE_MENU           = 118 // "Menu"
	SCANCODE_SELECT         = 119 // "Select"
	SCANCODE_STOP           = 120 // "Stop"
	SCANCODE_AGAIN          = 121 // "Again" (the Again key (Redo))
	SCANCODE_UNDO           = 122 // "Undo"
	SCANCODE_CUT            = 123 // "Cut"
	SCANCODE_COPY           = 124 // "Copy"
	SCANCODE_PASTE          = 125 // "Paste"
	SCANCODE_FIND           = 126 // "Find"
	SCANCODE_MUTE           = 127 // "Mute"
	SCANCODE_VOLUMEUP       = 128 // "VolumeUp"
	SCANCODE_VOLUMEDOWN     = 129 // "VolumeDown"
	SCANCODE_KP_COMMA       = 133 // "Keypad ," (the Comma key (numeric keypad))
	SCANCODE_KP_EQUALSAS400 = 134 // "Keypad = (AS400)" (the Equals AS400 key (numeric keypad))

	SCANCODE_INTERNATIONAL1 = 135 // "" (no name, empty string; used on Asian keyboards, see footnotes in USB doc)
	SCANCODE_INTERNATIONAL2 = 136 // "" (no name, empty string)
	SCANCODE_INTERNATIONAL3 = 137 // "" (no name, empty string; Yen)
	SCANCODE_INTERNATIONAL4 = 138 // "" (no name, empty string)
	SCANCODE_INTERNATIONAL5 = 139 // "" (no name, empty string)
	SCANCODE_INTERNATIONAL6 = 140 // "" (no name, empty string)
	SCANCODE_INTERNATIONAL7 = 141 // "" (no name, empty string)
	SCANCODE_INTERNATIONAL8 = 142 // "" (no name, empty string)
	SCANCODE_INTERNATIONAL9 = 143 // "" (no name, empty string)
	SCANCODE_LANG1          = 144 // "" (no name, empty string; Hangul/English toggle)
	SCANCODE_LANG2          = 145 // "" (no name, empty string; Hanja conversion)
	SCANCODE_LANG3          = 146 // "" (no name, empty string; Katakana)
	SCANCODE_LANG4          = 147 // "" (no name, empty string; Hiragana)
	SCANCODE_LANG5          = 148 // "" (no name, empty string; Zenkaku/Hankaku)
	SCANCODE_LANG6          = 149 // "" (no name, empty string; reserved)
	SCANCODE_LANG7          = 150 // "" (no name, empty string; reserved)
	SCANCODE_LANG8          = 151 // "" (no name, empty string; reserved)
	SCANCODE_LANG9          = 152 // "" (no name, empty string; reserved)

	SCANCODE_ALTERASE   = 153 // "AltErase" (Erase-Eaze)
	SCANCODE_SYSREQ     = 154 // "SysReq" (the SysReq key)
	SCANCODE_CANCEL     = 155 // "Cancel"
	SCANCODE_CLEAR      = 156 // "Clear"
	SCANCODE_PRIOR      = 157 // "Prior"
	SCANCODE_RETURN2    = 158 // "Return"
	SCANCODE_SEPARATOR  = 159 // "Separator"
	SCANCODE_OUT        = 160 // "Out"
	SCANCODE_OPER       = 161 // "Oper"
	SCANCODE_CLEARAGAIN = 162 // "Clear / Again"
	SCANCODE_CRSEL      = 163 // "CrSel"
	SCANCODE_EXSEL      = 164 // "ExSel"

	SCANCODE_KP_00              = 176 // "Keypad 00" (the 00 key (numeric keypad))
	SCANCODE_KP_000             = 177 // "Keypad 000" (the 000 key (numeric keypad))
	SCANCODE_THOUSANDSSEPARATOR = 178 // "ThousandsSeparator" (the Thousands Separator key)
	SCANCODE_DECIMALSEPARATOR   = 179 // "DecimalSeparator" (the Decimal Separator key)
	SCANCODE_CURRENCYUNIT       = 180 // "CurrencyUnit" (the Currency Unit key)
	SCANCODE_CURRENCYSUBUNIT    = 181 // "CurrencySubUnit" (the Currency Subunit key)
	SCANCODE_KP_LEFTPAREN       = 182 // "Keypad (" (the Left Parenthesis key (numeric keypad))
	SCANCODE_KP_RIGHTPAREN      = 183 // "Keypad )" (the Right Parenthesis key (numeric keypad))
	SCANCODE_KP_LEFTBRACE       = 184 // "Keypad {" (the Left Brace key (numeric keypad))
	SCANCODE_KP_RIGHTBRACE      = 185 // "Keypad }" (the Right Brace key (numeric keypad))
	SCANCODE_KP_TAB             = 186 // "Keypad Tab" (the Tab key (numeric keypad))
	SCANCODE_KP_BACKSPACE       = 187 // "Keypad Backspace" (the Backspace key (numeric keypad))
	SCANCODE_KP_A               = 188 // "Keypad A" (the A key (numeric keypad))
	SCANCODE_KP_B               = 189 // "Keypad B" (the B key (numeric keypad))
	SCANCODE_KP_C               = 190 // "Keypad C" (the C key (numeric keypad))
	SCANCODE_KP_D               = 191 // "Keypad D" (the D key (numeric keypad))
	SCANCODE_KP_E               = 192 // "Keypad E" (the E key (numeric keypad))
	SCANCODE_KP_F               = 193 // "Keypad F" (the F key (numeric keypad))
	SCANCODE_KP_XOR             = 194 // "Keypad XOR" (the XOR key (numeric keypad))
	SCANCODE_KP_POWER           = 195 // "Keypad ^" (the Power key (numeric keypad))
	SCANCODE_KP_PERCENT         = 196 // "Keypad %" (the Percent key (numeric keypad))
	SCANCODE_KP_LESS            = 197 // "Keypad <" (the Less key (numeric keypad))
	SCANCODE_KP_GREATER         = 198 // "Keypad >" (the Greater key (numeric keypad))
	SCANCODE_KP_AMPERSAND       = 199 // "Keypad &" (the & key (numeric keypad))
	SCANCODE_KP_DBLAMPERSAND    = 200 // "Keypad &&" (the && key (numeric keypad))
	SCANCODE_KP_VERTICALBAR     = 201 // "Keypad |" (the | key (numeric keypad))
	SCANCODE_KP_DBLVERTICALBAR  = 202 // "Keypad ||" (the || key (numeric keypad))
	SCANCODE_KP_COLON           = 203 // "Keypad :" (the : key (numeric keypad))
	SCANCODE_KP_HASH            = 204 // "Keypad #" (the # key (numeric keypad))
	SCANCODE_KP_SPACE           = 205 // "Keypad Space" (the Space key (numeric keypad))
	SCANCODE_KP_AT              = 206 // "Keypad @" (the @ key (numeric keypad))
	SCANCODE_KP_EXCLAM          = 207 // "Keypad !" (the ! key (numeric keypad))
	SCANCODE_KP_MEMSTORE        = 208 // "Keypad MemStore" (the Mem Store key (numeric keypad))
	SCANCODE_KP_MEMRECALL       = 209 // "Keypad MemRecall" (the Mem Recall key (numeric keypad))
	SCANCODE_KP_MEMCLEAR        = 210 // "Keypad MemClear" (the Mem Clear key (numeric keypad))
	SCANCODE_KP_MEMADD          = 211 // "Keypad MemAdd" (the Mem Add key (numeric keypad))
	SCANCODE_KP_MEMSUBTRACT     = 212 // "Keypad MemSubtract" (the Mem Subtract key (numeric keypad))
	SCANCODE_KP_MEMMULTIPLY     = 213 // "Keypad MemMultiply" (the Mem Multiply key (numeric keypad))
	SCANCODE_KP_MEMDIVIDE       = 214 // "Keypad MemDivide" (the Mem Divide key (numeric keypad))
	SCANCODE_KP_PLUSMINUS       = 215 // "Keypad +/-" (the +/- key (numeric keypad))
	SCANCODE_KP_CLEAR           = 216 // "Keypad Clear" (the Clear key (numeric keypad))
	SCANCODE_KP_CLEARENTRY      = 217 // "Keypad ClearEntry" (the Clear Entry key (numeric keypad))
	SCANCODE_KP_BINARY          = 218 // "Keypad Binary" (the Binary key (numeric keypad))
	SCANCODE_KP_OCTAL           = 219 // "Keypad Octal" (the Octal key (numeric keypad))
	SCANCODE_KP_DECIMAL         = 220 // "Keypad Decimal" (the Decimal key (numeric keypad))
	SCANCODE_KP_HEXADECIMAL     = 221 // "Keypad Hexadecimal" (the Hexadecimal key (numeric keypad))

	SCANCODE_LCTRL  = 224 // "Left Ctrl"
	SCANCODE_LSHIFT = 225 // "Left Shift"
	SCANCODE_LALT   = 226 // "Left Alt" (alt, option)
	SCANCODE_LGUI   = 227 // "Left GUI" (windows, command (apple), meta)
	SCANCODE_RCTRL  = 228 // "Right Ctrl"
	SCANCODE_RSHIFT = 229 // "Right Shift"
	SCANCODE_RALT   = 230 // "Right Alt" (alt gr, option)
	SCANCODE_RGUI   = 231 // "Right GUI" (windows, command (apple), meta)

	SCANCODE_MODE             = 257 // "ModeSwitch" (I'm not sure if this is really not covered by any of the above, but since there's a special KMOD_MODE for it I'm adding it here)
	SCANCODE_AUDIONEXT        = 258 // "AudioNext" (the Next Track media key)
	SCANCODE_AUDIOPREV        = 259 // "AudioPrev" (the Previous Track media key)
	SCANCODE_AUDIOSTOP        = 260 // "AudioStop" (the Stop media key)
	SCANCODE_AUDIOPLAY        = 261 // "AudioPlay" (the Play media key)
	SCANCODE_AUDIOMUTE        = 262 // "AudioMute" (the Mute volume key)
	SCANCODE_MEDIASELECT      = 263 // "MediaSelect" (the Media Select key)
	SCANCODE_WWW              = 264 // "WWW" (the WWW/World Wide Web key)
	SCANCODE_MAIL             = 265 // "Mail" (the Mail/eMail key)
	SCANCODE_CALCULATOR       = 266 // "Calculator" (the Calculator key)
	SCANCODE_COMPUTER         = 267 // "Computer" (the My Computer key)
	SCANCODE_AC_SEARCH        = 268 // "AC Search" (the Search key (application control keypad))
	SCANCODE_AC_HOME          = 269 // "AC Home" (the Home key (application control keypad))
	SCANCODE_AC_BACK          = 270 // "AC Back" (the Back key (application control keypad))
	SCANCODE_AC_FORWARD       = 271 // "AC Forward" (the Forward key (application control keypad))
	SCANCODE_AC_STOP          = 272 // "AC Stop" (the Stop key (application control keypad))
	SCANCODE_AC_REFRESH       = 273 // "AC Refresh" (the Refresh key (application control keypad))
	SCANCODE_AC_BOOKMARKS     = 274 // "AC Bookmarks" (the Bookmarks key (application control keypad))
	SCANCODE_BRIGHTNESSDOWN   = 275 // "BrightnessDown" (the Brightness Down key)
	SCANCODE_BRIGHTNESSUP     = 276 // "BrightnessUp" (the Brightness Up key)
	SCANCODE_DISPLAYSWITCH    = 277 // "DisplaySwitch" (display mirroring/dual display switch, video mode switch)
	SCANCODE_KBDILLUMTOGGLE   = 278 // "KBDIllumToggle" (the Keyboard Illumination Toggle key)
	SCANCODE_KBDILLUMDOWN     = 279 // "KBDIllumDown" (the Keyboard Illumination Down key)
	SCANCODE_KBDILLUMUP       = 280 // "KBDIllumUp" (the Keyboard Illumination Up key)
	SCANCODE_EJECT            = 281 // "Eject" (the Eject key)
	SCANCODE_SLEEP            = 282 // "Sleep" (the Sleep key)
	SCANCODE_APP1             = 283
	SCANCODE_APP2             = 284
	SCANCODE_AUDIOREWIND      = 285
	SCANCODE_AUDIOFASTFORWARD = 286
	NUM_SCANCODES             = 512
)

// These are the flags which may be passed to SDL_Init().
// (https://wiki.libsdl.org/SDL_Init)
const (
	INIT_TIMER          = 0x00000001 // timer subsystem
	INIT_AUDIO          = 0x00000010 // audio subsystem
	INIT_VIDEO          = 0x00000020 // video subsystem; automatically initializes the events subsystem
	INIT_JOYSTICK       = 0x00000200 // joystick subsystem; automatically initializes the events subsystem
	INIT_HAPTIC         = 0x00001000 // haptic (force feedback) subsystem
	INIT_GAMECONTROLLER = 0x00002000 // controller subsystem; automatically initializes the joystick subsystem
	INIT_EVENTS         = 0x00004000 // events subsystem
	INIT_NOPARACHUTE    = 0x00100000 // compatibility; this flag is ignored
	INIT_SENSOR         = 0x00008000 // sensor subsystem
	INIT_EVERYTHING     = INIT_TIMER | INIT_AUDIO | INIT_VIDEO | INIT_EVENTS |
		INIT_JOYSTICK | INIT_HAPTIC | INIT_GAMECONTROLLER |
		INIT_SENSOR // all of the above subsystems
)

const (
	RELEASED = 0
	PRESSED  = 1
)

// Surface flags (internal use)
const (
	SWSURFACE = 0          // just here for compatibility
	PREALLOC  = 0x00000001 // surface uses preallocated memory
	RLEACCEL  = 0x00000002 // surface is RLE encoded
	DONTFREE  = 0x00000004 // surface is referenced internally
)

// YUV Conversion Modes
const (
	YUV_CONVERSION_JPEG      YUV_CONVERSION_MODE = iota // Full range JPEG
	YUV_CONVERSION_BT601                                // BT.601 (the default)
	YUV_CONVERSION_BT709                                // BT.709
	YUV_CONVERSION_AUTOMATIC                            // BT.601 for SD content, BT.709 for HD content
)

// Various supported windowing subsystems.
const (
	SYSWM_UNKNOWN  = iota
	SYSWM_WINDOWS  // Microsoft Windows
	SYSWM_X11      // X Window System
	SYSWM_DIRECTFB // DirectFB
	SYSWM_COCOA    // Apple Mac OS X
	SYSWM_UIKIT    // Apple iOS
	SYSWM_WAYLAND  // Wayland (>= SDL 2.0.2)
	SYSWM_MIR      // Mir (>= SDL 2.0.2)
	SYSWM_WINRT    // WinRT (>= SDL 2.0.3)
	SYSWM_ANDROID  // Android (>= SDL 2.0.4)
	SYSWM_VIVANTE  // Vivante (>= SDL 2.0.5)
	SYSWM_OS2
)

// The version of SDL in use.
// NOTE that this is currently the version that this Go wrapper was created
// with. You should check your DLL's version using GetVersion.
const (
	MAJOR_VERSION = 2 // major version
	MINOR_VERSION = 0 // minor version
	PATCHLEVEL    = 9 // update version (patchlevel)
)

// An enumeration of window states.
// (https://wiki.libsdl.org/SDL_WindowFlags)
const (
	WINDOW_FULLSCREEN         = 0x00000001                     // fullscreen window
	WINDOW_OPENGL             = 0x00000002                     // window usable with OpenGL context
	WINDOW_SHOWN              = 0x00000004                     // window is visible
	WINDOW_HIDDEN             = 0x00000008                     // window is not visible
	WINDOW_BORDERLESS         = 0x00000010                     // no window decoration
	WINDOW_RESIZABLE          = 0x00000020                     // window can be resized
	WINDOW_MINIMIZED          = 0x00000040                     // window is minimized
	WINDOW_MAXIMIZED          = 0x00000080                     // window is maximized
	WINDOW_INPUT_GRABBED      = 0x00000100                     // window has grabbed input focus
	WINDOW_INPUT_FOCUS        = 0x00000200                     // window has input focus
	WINDOW_MOUSE_FOCUS        = 0x00000400                     // window has mouse focus
	WINDOW_FULLSCREEN_DESKTOP = WINDOW_FULLSCREEN | 0x00001000 // fullscreen window at the current desktop resolution
	WINDOW_FOREIGN            = 0x00000800                     // window not created by SDL
	WINDOW_ALLOW_HIGHDPI      = 0x00002000                     // window should be created in high-DPI mode if supported (>= SDL 2.0.1)
	WINDOW_MOUSE_CAPTURE      = 0x00004000                     // window has mouse captured (unrelated to INPUT_GRABBED, >= SDL 2.0.4)
	WINDOW_ALWAYS_ON_TOP      = 0x00008000                     // window should always be above others (X11 only, >= SDL 2.0.5)
	WINDOW_SKIP_TASKBAR       = 0x00010000                     // window should not be added to the taskbar (X11 only, >= SDL 2.0.5)
	WINDOW_UTILITY            = 0x00020000                     // window should be treated as a utility window (X11 only, >= SDL 2.0.5)
	WINDOW_TOOLTIP            = 0x00040000                     // window should be treated as a tooltip (X11 only, >= SDL 2.0.5)
	WINDOW_POPUP_MENU         = 0x00080000                     // window should be treated as a popup menu (X11 only, >= SDL 2.0.5)
	WINDOW_VULKAN             = 0x10000000                     // window usable for Vulkan surface (>= SDL 2.0.6)
)

// An enumeration of window events.
// (https://wiki.libsdl.org/SDL_WindowEventID)
const (
	WINDOWEVENT_NONE         = iota // (never used)
	WINDOWEVENT_SHOWN               // window has been shown
	WINDOWEVENT_HIDDEN              // window has been hidden
	WINDOWEVENT_EXPOSED             // window has been exposed and should be redrawn
	WINDOWEVENT_MOVED               // window has been moved to data1, data2
	WINDOWEVENT_RESIZED             // window has been resized to data1xdata2; this event is always preceded by WINDOWEVENT_SIZE_CHANGED
	WINDOWEVENT_SIZE_CHANGED        // window size has changed, either as a result of an API call or through the system or user changing the window size; this event is followed by WINDOWEVENT_RESIZED if the size was changed by an external event, i.e. the user or the window manager
	WINDOWEVENT_MINIMIZED           // window has been minimized
	WINDOWEVENT_MAXIMIZED           // window has been maximized
	WINDOWEVENT_RESTORED            // window has been restored to normal size and position
	WINDOWEVENT_ENTER               // window has gained mouse focus
	WINDOWEVENT_LEAVE               // window has lost mouse focus
	WINDOWEVENT_FOCUS_GAINED        // window has gained keyboard focus
	WINDOWEVENT_FOCUS_LOST          // window has lost keyboard focus
	WINDOWEVENT_CLOSE               // the window manager requests that the window be closed
	WINDOWEVENT_TAKE_FOCUS          // window is being offered a focus (should SDL_SetWindowInputFocus() on itself or a subwindow, or ignore) (>= SDL 2.0.5)
	WINDOWEVENT_HIT_TEST            // window had a hit test that wasn't SDL_HITTEST_NORMAL (>= SDL 2.0.5)
)

// Window position flags.
// (https://wiki.libsdl.org/SDL_CreateWindow)
const (
	WINDOWPOS_UNDEFINED_MASK = 0x1FFF0000 // used to indicate that you don't care what the window position is
	WINDOWPOS_UNDEFINED      = 0x1FFF0000 // used to indicate that you don't care what the window position is
	WINDOWPOS_CENTERED_MASK  = 0x2FFF0000 // used to indicate that the window position should be centered
	WINDOWPOS_CENTERED       = 0x2FFF0000 // used to indicate that the window position should be centered
)

// An enumeration of message box flags (e.g. if supported message box will display warning icon).
// (https://wiki.libsdl.org/SDL_MessageBoxFlags)
const (
	MESSAGEBOX_ERROR       = 0x00000010 // error dialog
	MESSAGEBOX_WARNING     = 0x00000020 // warning dialog
	MESSAGEBOX_INFORMATION = 0x00000040 // informational dialog
)

// Flags for MessageBoxButtonData.
const (
	MESSAGEBOX_BUTTON_RETURNKEY_DEFAULT = 0x00000001 // marks the default button when return is hit
	MESSAGEBOX_BUTTON_ESCAPEKEY_DEFAULT = 0x00000002 // marks the default button when escape is hit
)

// OpenGL configuration attributes.
// (https://wiki.libsdl.org/SDL_GL_SetAttribute)
const (
	GL_RED_SIZE                   = iota // the minimum number of bits for the red channel of the color buffer; defaults to 3
	GL_GREEN_SIZE                        // the minimum number of bits for the green channel of the color buffer; defaults to 3
	GL_BLUE_SIZE                         // the minimum number of bits for the blue channel of the color buffer; defaults to 2
	GL_ALPHA_SIZE                        // the minimum number of bits for the alpha channel of the color buffer; defaults to 0
	GL_BUFFER_SIZE                       // the minimum number of bits for frame buffer size; defaults to 0
	GL_DOUBLEBUFFER                      // whether the output is single or double buffered; defaults to double buffering on
	GL_DEPTH_SIZE                        // the minimum number of bits in the depth buffer; defaults to 16
	GL_STENCIL_SIZE                      // the minimum number of bits in the stencil buffer; defaults to 0
	GL_ACCUM_RED_SIZE                    // the minimum number of bits for the red channel of the accumulation buffer; defaults to 0
	GL_ACCUM_GREEN_SIZE                  // the minimum number of bits for the green channel of the accumulation buffer; defaults to 0
	GL_ACCUM_BLUE_SIZE                   // the minimum number of bits for the blue channel of the accumulation buffer; defaults to 0
	GL_ACCUM_ALPHA_SIZE                  // the minimum number of bits for the alpha channel of the accumulation buffer; defaults to 0
	GL_STEREO                            // whether the output is stereo 3D; defaults to off
	GL_MULTISAMPLEBUFFERS                // the number of buffers used for multisample anti-aliasing; defaults to 0; see Remarks for details
	GL_MULTISAMPLESAMPLES                // the number of samples used around the current pixel used for multisample anti-aliasing; defaults to 0; see Remarks for details
	GL_ACCELERATED_VISUAL                // set to 1 to require hardware acceleration, set to 0 to force software rendering; defaults to allow either
	GL_RETAINED_BACKING                  // not used (deprecated)
	GL_CONTEXT_MAJOR_VERSION             // OpenGL context major version
	GL_CONTEXT_MINOR_VERSION             // OpenGL context minor version
	GL_CONTEXT_EGL                       // not used (deprecated)
	GL_CONTEXT_FLAGS                     // some combination of 0 or more of elements of the GLcontextFlag enumeration; defaults to 0 (https://wiki.libsdl.org/SDL_GLcontextFlag)
	GL_CONTEXT_PROFILE_MASK              // type of GL context (Core, Compatibility, ES); default value depends on platform (https://wiki.libsdl.org/SDL_GLprofile)
	GL_SHARE_WITH_CURRENT_CONTEXT        // OpenGL context sharing; defaults to 0
	GL_FRAMEBUFFER_SRGB_CAPABLE          // requests sRGB capable visual; defaults to 0 (>= SDL 2.0.1)
	GL_CONTEXT_RELEASE_BEHAVIOR          // sets context the release behavior; defaults to 1 (>= SDL 2.0.4)
	GL_CONTEXT_RESET_NOTIFICATION        // (>= SDL 2.0.6)
	GL_CONTEXT_NO_ERROR                  // (>= SDL 2.0.6)
)

// An enumeration of OpenGL profiles.
// (https://wiki.libsdl.org/SDL_GLprofile)
const (
	GL_CONTEXT_PROFILE_CORE          = 0x0001 // OpenGL core profile - deprecated functions are disabled
	GL_CONTEXT_PROFILE_COMPATIBILITY = 0x0002 // OpenGL compatibility profile - deprecated functions are allowed
	GL_CONTEXT_PROFILE_ES            = 0x0004 // OpenGL ES profile - only a subset of the base OpenGL functionality is available
)

// An enumeration of OpenGL context configuration flags.
// (https://wiki.libsdl.org/SDL_GLcontextFlag)
const (
	GL_CONTEXT_DEBUG_FLAG              = 0x0001 // intended to put the GL into a "debug" mode which might offer better developer insights, possibly at a loss of performance
	GL_CONTEXT_FORWARD_COMPATIBLE_FLAG = 0x0002 // intended to put the GL into a "forward compatible" mode, which means that no deprecated functionality will be supported, possibly at a gain in performance, and only applies to GL 3.0 and later contexts
	GL_CONTEXT_ROBUST_ACCESS_FLAG      = 0x0004 // intended to require a GL context that supports the GL_ARB_robustness extension--a mode that offers a few APIs that are safer than the usual defaults (think snprintf() vs sprintf())
	GL_CONTEXT_RESET_ISOLATION_FLAG    = 0x0008 // intended to require the GL to make promises about what to do in the face of driver or hardware failure
)

// CACHELINE_SIZE is a cacheline size used for padding.
const CACHELINE_SIZE = 128

const K_SCANCODE_MASK = 1 << 30

// MIX_MAXVOLUME is the full audio volume value used in MixAudioFormat() and AudioFormat().
// (https://wiki.libsdl.org/SDL_MixAudioFormat)
const MIX_MAXVOLUME = 128 // full audio volume

// TOUCH_MOUSEID is the device ID for mouse events simulated with touch input
const TOUCH_MOUSEID = 0xFFFFFFFF // uint32(-1)

const (
	PIXELFORMAT_RGBA32 = PIXELFORMAT_ABGR8888
	PIXELFORMAT_ARGB32 = PIXELFORMAT_BGRA8888
	PIXELFORMAT_BGRA32 = PIXELFORMAT_ARGB8888
	PIXELFORMAT_ABGR32 = PIXELFORMAT_RGBA8888
)

const SDL_STANDARD_GRAVITY = 9.80665

const TEXTINPUTEVENT_TEXT_SIZE = 32

var ErrInvalidParameters = errors.New("Invalid Parameters")

var (
	dll = syscall.NewLazyDLL("SDL2.dll")

	addHintCallback                   = dll.NewProc("SDL_AddHintCallback")
	audioInit                         = dll.NewProc("SDL_AudioInit")
	audioQuit                         = dll.NewProc("SDL_AudioQuit")
	buildAudioCVT                     = dll.NewProc("SDL_BuildAudioCVT")
	calculateGammaRamp                = dll.NewProc("SDL_CalculateGammaRamp")
	captureMouse                      = dll.NewProc("SDL_CaptureMouse")
	clearError                        = dll.NewProc("SDL_ClearError")
	clearHints                        = dll.NewProc("SDL_ClearHints")
	clearQueuedAudio                  = dll.NewProc("SDL_ClearQueuedAudio")
	getError                          = dll.NewProc("SDL_GetError")
	closeAudio                        = dll.NewProc("SDL_CloseAudio")
	closeAudioDevice                  = dll.NewProc("SDL_CloseAudioDevice")
	convertAudio                      = dll.NewProc("SDL_ConvertAudio")
	convertPixels                     = dll.NewProc("SDL_ConvertPixels")
	createWindowAndRenderer           = dll.NewProc("SDL_CreateWindowAndRenderer")
	delEventWatch                     = dll.NewProc("SDL_DelEventWatch")
	delay                             = dll.NewProc("SDL_Delay")
	dequeueAudio                      = dll.NewProc("SDL_DequeueAudio")
	disableScreenSaver                = dll.NewProc("SDL_DisableScreenSaver")
	enableScreenSaver                 = dll.NewProc("SDL_EnableScreenSaver")
	sdlError                          = dll.NewProc("SDL_Error")
	flushEvent                        = dll.NewProc("SDL_FlushEvent")
	flushEvents                       = dll.NewProc("SDL_FlushEvents")
	freeCursor                        = dll.NewProc("SDL_FreeCursor")
	freeWAV                           = dll.NewProc("SDL_FreeWAV")
	gl_DeleteContext                  = dll.NewProc("SDL_GL_DeleteContext")
	gl_ExtensionSupported             = dll.NewProc("SDL_GL_ExtensionSupported")
	gl_GetAttribute                   = dll.NewProc("SDL_GL_GetAttribute")
	gl_GetProcAddress                 = dll.NewProc("SDL_GL_GetProcAddress")
	gl_GetSwapInterval                = dll.NewProc("SDL_GL_GetSwapInterval")
	gl_LoadLibrary                    = dll.NewProc("SDL_GL_LoadLibrary")
	gl_SetAttribute                   = dll.NewProc("SDL_GL_SetAttribute")
	gl_SetSwapInterval                = dll.NewProc("SDL_GL_SetSwapInterval")
	gl_UnloadLibrary                  = dll.NewProc("SDL_GL_UnloadLibrary")
	gameControllerAddMapping          = dll.NewProc("SDL_GameControllerAddMapping")
	gameControllerEventState          = dll.NewProc("SDL_GameControllerEventState")
	gameControllerGetStringForAxis    = dll.NewProc("SDL_GameControllerGetStringForAxis")
	gameControllerGetStringForButton  = dll.NewProc("SDL_GameControllerGetStringForButton")
	gameControllerMappingForGUID      = dll.NewProc("SDL_GameControllerMappingForGUID")
	gameControllerMappingForIndex     = dll.NewProc("SDL_GameControllerMappingForIndex")
	gameControllerNameForIndex        = dll.NewProc("SDL_GameControllerNameForIndex")
	gameControllerNumMappings         = dll.NewProc("SDL_GameControllerNumMappings")
	gameControllerUpdate              = dll.NewProc("SDL_GameControllerUpdate")
	getAudioDeviceName                = dll.NewProc("SDL_GetAudioDeviceName")
	getAudioDriver                    = dll.NewProc("SDL_GetAudioDriver")
	getBasePath                       = dll.NewProc("SDL_GetBasePath")
	getCPUCacheLineSize               = dll.NewProc("SDL_GetCPUCacheLineSize")
	getCPUCount                       = dll.NewProc("SDL_GetCPUCount")
	getClipboardText                  = dll.NewProc("SDL_GetClipboardText")
	getCurrentAudioDriver             = dll.NewProc("SDL_GetCurrentAudioDriver")
	getCurrentVideoDriver             = dll.NewProc("SDL_GetCurrentVideoDriver")
	getDisplayDPI                     = dll.NewProc("SDL_GetDisplayDPI")
	getDisplayName                    = dll.NewProc("SDL_GetDisplayName")
	eventState                        = dll.NewProc("SDL_EventState")
	filterEvents                      = dll.NewProc("SDL_FilterEvents")
	getHint                           = dll.NewProc("SDL_GetHint")
	getKeyName                        = dll.NewProc("SDL_GetKeyName")
	getKeyboardState                  = dll.NewProc("SDL_GetKeyboardState")
	getMouseState                     = dll.NewProc("SDL_GetMouseState")
	getNumAudioDevices                = dll.NewProc("SDL_GetNumAudioDevices")
	getNumAudioDrivers                = dll.NewProc("SDL_GetNumAudioDrivers")
	getNumDisplayModes                = dll.NewProc("SDL_GetNumDisplayModes")
	getNumRenderDrivers               = dll.NewProc("SDL_GetNumRenderDrivers")
	getNumTouchDevices                = dll.NewProc("SDL_GetNumTouchDevices")
	getNumTouchFingers                = dll.NewProc("SDL_GetNumTouchFingers")
	getNumVideoDisplays               = dll.NewProc("SDL_GetNumVideoDisplays")
	getNumVideoDrivers                = dll.NewProc("SDL_GetNumVideoDrivers")
	getPerformanceCounter             = dll.NewProc("SDL_GetPerformanceCounter")
	getPerformanceFrequency           = dll.NewProc("SDL_GetPerformanceFrequency")
	getPixelFormatName                = dll.NewProc("SDL_GetPixelFormatName")
	getPlatform                       = dll.NewProc("SDL_GetPlatform")
	getPowerInfo                      = dll.NewProc("SDL_GetPowerInfo")
	getPrefPath                       = dll.NewProc("SDL_GetPrefPath")
	getQueuedAudioSize                = dll.NewProc("SDL_GetQueuedAudioSize")
	getRGB                            = dll.NewProc("SDL_GetRGB")
	getRGBA                           = dll.NewProc("SDL_GetRGBA")
	getRelativeMouseMode              = dll.NewProc("SDL_GetRelativeMouseMode")
	getRelativeMouseState             = dll.NewProc("SDL_GetRelativeMouseState")
	getRenderDriverInfo               = dll.NewProc("SDL_GetRenderDriverInfo")
	getRevision                       = dll.NewProc("SDL_GetRevision")
	getRevisionNumber                 = dll.NewProc("SDL_GetRevisionNumber")
	getScancodeName                   = dll.NewProc("SDL_GetScancodeName")
	getSystemRAM                      = dll.NewProc("SDL_GetSystemRAM")
	getTicks                          = dll.NewProc("SDL_GetTicks")
	getVersion                        = dll.NewProc("SDL_GetVersion")
	getVideoDriver                    = dll.NewProc("SDL_GetVideoDriver")
	hapticIndex                       = dll.NewProc("SDL_HapticIndex")
	hapticName                        = dll.NewProc("SDL_HapticName")
	hapticOpened                      = dll.NewProc("SDL_HapticOpened")
	has3DNow                          = dll.NewProc("SDL_Has3DNow")
	hasAVX                            = dll.NewProc("SDL_HasAVX")
	hasAVX2                           = dll.NewProc("SDL_HasAVX2")
	hasAltiVec                        = dll.NewProc("SDL_HasAltiVec")
	hasClipboardText                  = dll.NewProc("SDL_HasClipboardText")
	hasEvent                          = dll.NewProc("SDL_HasEvent")
	hasEvents                         = dll.NewProc("SDL_HasEvents")
	hasMMX                            = dll.NewProc("SDL_HasMMX")
	hasNEON                           = dll.NewProc("SDL_HasNEON")
	hasRDTSC                          = dll.NewProc("SDL_HasRDTSC")
	hasSSE                            = dll.NewProc("SDL_HasSSE")
	hasSSE2                           = dll.NewProc("SDL_HasSSE2")
	hasSSE3                           = dll.NewProc("SDL_HasSSE3")
	hasSSE41                          = dll.NewProc("SDL_HasSSE41")
	hasSSE42                          = dll.NewProc("SDL_HasSSE42")
	hasScreenKeyboardSupport          = dll.NewProc("SDL_HasScreenKeyboardSupport")
	sdlInit                           = dll.NewProc("SDL_Init")
	initSubSystem                     = dll.NewProc("SDL_InitSubSystem")
	isGameController                  = dll.NewProc("SDL_IsGameController")
	isScreenKeyboardShown             = dll.NewProc("SDL_IsScreenKeyboardShown")
	isScreenSaverEnabled              = dll.NewProc("SDL_IsScreenSaverEnabled")
	isTextInputActive                 = dll.NewProc("SDL_IsTextInputActive")
	joystickEventState                = dll.NewProc("SDL_JoystickEventState")
	joystickGetDeviceProduct          = dll.NewProc("SDL_JoystickGetDeviceProduct")
	joystickGetDeviceProductVersion   = dll.NewProc("SDL_JoystickGetDeviceProductVersion")
	joystickGetDeviceVendor           = dll.NewProc("SDL_JoystickGetDeviceVendor")
	joystickIsHaptic                  = dll.NewProc("SDL_JoystickIsHaptic")
	joystickNameForIndex              = dll.NewProc("SDL_JoystickNameForIndex")
	joystickUpdate                    = dll.NewProc("SDL_JoystickUpdate")
	loadDollarTemplates               = dll.NewProc("SDL_LoadDollarTemplates")
	loadFile                          = dll.NewProc("SDL_LoadFile")
	lockAudio                         = dll.NewProc("SDL_LockAudio")
	lockAudioDevice                   = dll.NewProc("SDL_LockAudioDevice")
	lockJoysticks                     = dll.NewProc("SDL_LockJoysticks")
	log                               = dll.NewProc("SDL_Log")
	logCritical                       = dll.NewProc("SDL_LogCritical")
	logDebug                          = dll.NewProc("SDL_LogDebug")
	logError                          = dll.NewProc("SDL_LogError")
	logInfo                           = dll.NewProc("SDL_LogInfo")
	logMessage                        = dll.NewProc("SDL_LogMessage")
	logResetPriorities                = dll.NewProc("SDL_LogResetPriorities")
	logSetAllPriority                 = dll.NewProc("SDL_LogSetAllPriority")
	logSetPriority                    = dll.NewProc("SDL_LogSetPriority")
	logVerbose                        = dll.NewProc("SDL_LogVerbose")
	logWarn                           = dll.NewProc("SDL_LogWarn")
	mapRGB                            = dll.NewProc("SDL_MapRGB")
	mapRGBA                           = dll.NewProc("SDL_MapRGBA")
	masksToPixelFormatEnum            = dll.NewProc("SDL_MasksToPixelFormatEnum")
	mixAudio                          = dll.NewProc("SDL_MixAudio")
	mixAudioFormat                    = dll.NewProc("SDL_MixAudioFormat")
	mouseIsHaptic                     = dll.NewProc("SDL_MouseIsHaptic")
	numHaptics                        = dll.NewProc("SDL_NumHaptics")
	numJoysticks                      = dll.NewProc("SDL_NumJoysticks")
	numSensors                        = dll.NewProc("SDL_NumSensors")
	openAudio                         = dll.NewProc("SDL_OpenAudio")
	pauseAudio                        = dll.NewProc("SDL_PauseAudio")
	pauseAudioDevice                  = dll.NewProc("SDL_PauseAudioDevice")
	peepEvents                        = dll.NewProc("SDL_PeepEvents")
	pixelFormatEnumToMasks            = dll.NewProc("SDL_PixelFormatEnumToMasks")
	pumpEvents                        = dll.NewProc("SDL_PumpEvents")
	pushEvent                         = dll.NewProc("SDL_PushEvent")
	queueAudio                        = dll.NewProc("SDL_QueueAudio")
	quit                              = dll.NewProc("SDL_Quit")
	quitSubSystem                     = dll.NewProc("SDL_QuitSubSystem")
	recordGesture                     = dll.NewProc("SDL_RecordGesture")
	registerEvents                    = dll.NewProc("SDL_RegisterEvents")
	saveAllDollarTemplates            = dll.NewProc("SDL_SaveAllDollarTemplates")
	saveDollarTemplate                = dll.NewProc("SDL_SaveDollarTemplate")
	sensorGetDeviceName               = dll.NewProc("SDL_SensorGetDeviceName")
	sensorGetDeviceNonPortableType    = dll.NewProc("SDL_SensorGetDeviceNonPortableType")
	sensorUpdate                      = dll.NewProc("SDL_SensorUpdate")
	setClipboardText                  = dll.NewProc("SDL_SetClipboardText")
	setCursor                         = dll.NewProc("SDL_SetCursor")
	setError                          = dll.NewProc("SDL_SetError")
	setEventFilter                    = dll.NewProc("SDL_SetEventFilter")
	setHint                           = dll.NewProc("SDL_SetHint")
	setHintWithPriority               = dll.NewProc("SDL_SetHintWithPriority")
	setModState                       = dll.NewProc("SDL_SetModState")
	setRelativeMouseMode              = dll.NewProc("SDL_SetRelativeMouseMode")
	setTextInputRect                  = dll.NewProc("SDL_SetTextInputRect")
	setYUVConversionMode              = dll.NewProc("SDL_SetYUVConversionMode")
	showCursor                        = dll.NewProc("SDL_ShowCursor")
	showMessageBox                    = dll.NewProc("SDL_ShowMessageBox")
	showSimpleMessageBox              = dll.NewProc("SDL_ShowSimpleMessageBox")
	startTextInput                    = dll.NewProc("SDL_StartTextInput")
	stopTextInput                     = dll.NewProc("SDL_StopTextInput")
	unlockAudio                       = dll.NewProc("SDL_UnlockAudio")
	unlockAudioDevice                 = dll.NewProc("SDL_UnlockAudioDevice")
	unlockJoysticks                   = dll.NewProc("SDL_UnlockJoysticks")
	videoInit                         = dll.NewProc("SDL_VideoInit")
	videoQuit                         = dll.NewProc("SDL_VideoQuit")
	vulkan_GetVkGetInstanceProcAddr   = dll.NewProc("SDL_Vulkan_GetVkGetInstanceProcAddr")
	vulkan_LoadLibrary                = dll.NewProc("SDL_Vulkan_LoadLibrary")
	vulkan_UnloadLibrary              = dll.NewProc("SDL_Vulkan_UnloadLibrary")
	warpMouseGlobal                   = dll.NewProc("SDL_WarpMouseGlobal")
	wasInit                           = dll.NewProc("SDL_WasInit")
	openAudioDevice                   = dll.NewProc("SDL_OpenAudioDevice")
	getAudioDeviceStatus              = dll.NewProc("SDL_GetAudioDeviceStatus")
	getAudioStatus                    = dll.NewProc("SDL_GetAudioStatus")
	newAudioStream                    = dll.NewProc("SDL_NewAudioStream")
	audioStreamAvailable              = dll.NewProc("SDL_AudioStreamAvailable")
	audioStreamClear                  = dll.NewProc("SDL_AudioStreamClear")
	audioStreamFlush                  = dll.NewProc("SDL_AudioStreamFlush")
	freeAudioStream                   = dll.NewProc("SDL_FreeAudioStream")
	audioStreamGet                    = dll.NewProc("SDL_AudioStreamGet")
	audioStreamPut                    = dll.NewProc("SDL_AudioStreamPut")
	composeCustomBlendMode            = dll.NewProc("SDL_ComposeCustomBlendMode")
	createCond                        = dll.NewProc("SDL_CreateCond")
	condBroadcast                     = dll.NewProc("SDL_CondBroadcast")
	destroyCond                       = dll.NewProc("SDL_DestroyCond")
	condSignal                        = dll.NewProc("SDL_CondSignal")
	condWait                          = dll.NewProc("SDL_CondWait")
	condWaitTimeout                   = dll.NewProc("SDL_CondWaitTimeout")
	createColorCursor                 = dll.NewProc("SDL_CreateColorCursor")
	createCursor                      = dll.NewProc("SDL_CreateCursor")
	createSystemCursor                = dll.NewProc("SDL_CreateSystemCursor")
	delHintCallback                   = dll.NewProc("SDL_DelHintCallback")
	getCursor                         = dll.NewProc("SDL_GetCursor")
	getDefaultCursor                  = dll.NewProc("SDL_GetDefaultCursor")
	getClosestDisplayMode             = dll.NewProc("SDL_GetClosestDisplayMode")
	getCurrentDisplayMode             = dll.NewProc("SDL_GetCurrentDisplayMode")
	getDesktopDisplayMode             = dll.NewProc("SDL_GetDesktopDisplayMode")
	getDisplayMode                    = dll.NewProc("SDL_GetDisplayMode")
	pollEvent                         = dll.NewProc("SDL_PollEvent")
	waitEvent                         = dll.NewProc("SDL_WaitEvent")
	waitEventTimeout                  = dll.NewProc("SDL_WaitEventTimeout")
	addEventWatch                     = dll.NewProc("SDL_AddEventWatch")
	getTouchFinger                    = dll.NewProc("SDL_GetTouchFinger")
	gameControllerFromInstanceID      = dll.NewProc("SDL_GameControllerFromInstanceID")
	gameControllerOpen                = dll.NewProc("SDL_GameControllerOpen")
	gameControllerGetAttached         = dll.NewProc("SDL_GameControllerGetAttached")
	gameControllerGetAxis             = dll.NewProc("SDL_GameControllerGetAxis")
	gameControllerGetBindForAxis      = dll.NewProc("SDL_GameControllerGetBindForAxis")
	gameControllerGetBindForButton    = dll.NewProc("SDL_GameControllerGetBindForButton")
	gameControllerGetButton           = dll.NewProc("SDL_GameControllerGetButton")
	gameControllerClose               = dll.NewProc("SDL_GameControllerClose")
	gameControllerGetJoystick         = dll.NewProc("SDL_GameControllerGetJoystick")
	gameControllerMapping             = dll.NewProc("SDL_GameControllerMapping")
	gameControllerName                = dll.NewProc("SDL_GameControllerName")
	gameControllerGetProduct          = dll.NewProc("SDL_GameControllerGetProduct")
	gameControllerGetProductVersion   = dll.NewProc("SDL_GameControllerGetProductVersion")
	gameControllerGetVendor           = dll.NewProc("SDL_GameControllerGetVendor")
	gameControllerGetAxisFromString   = dll.NewProc("SDL_GameControllerGetAxisFromString")
	gameControllerGetButtonFromString = dll.NewProc("SDL_GameControllerGetButtonFromString")
	hapticOpen                        = dll.NewProc("SDL_HapticOpen")
	hapticOpenFromJoystick            = dll.NewProc("SDL_HapticOpenFromJoystick")
	hapticOpenFromMouse               = dll.NewProc("SDL_HapticOpenFromMouse")
	hapticClose                       = dll.NewProc("SDL_HapticClose")
	hapticDestroyEffect               = dll.NewProc("SDL_HapticDestroyEffect")
	hapticEffectSupported             = dll.NewProc("SDL_HapticEffectSupported")
	hapticGetEffectStatus             = dll.NewProc("SDL_HapticGetEffectStatus")
	hapticNewEffect                   = dll.NewProc("SDL_HapticNewEffect")
	hapticNumAxes                     = dll.NewProc("SDL_HapticNumAxes")
	hapticNumEffects                  = dll.NewProc("SDL_HapticNumEffects")
	hapticNumEffectsPlaying           = dll.NewProc("SDL_HapticNumEffectsPlaying")
	hapticPause                       = dll.NewProc("SDL_HapticPause")
	hapticQuery                       = dll.NewProc("SDL_HapticQuery")
	hapticRumbleInit                  = dll.NewProc("SDL_HapticRumbleInit")
	hapticRumblePlay                  = dll.NewProc("SDL_HapticRumblePlay")
	hapticRumbleStop                  = dll.NewProc("SDL_HapticRumbleStop")
	hapticRumbleSupported             = dll.NewProc("SDL_HapticRumbleSupported")
	hapticRunEffect                   = dll.NewProc("SDL_HapticRunEffect")
	hapticSetAutocenter               = dll.NewProc("SDL_HapticSetAutocenter")
	hapticSetGain                     = dll.NewProc("SDL_HapticSetGain")
	hapticStopAll                     = dll.NewProc("SDL_HapticStopAll")
	hapticStopEffect                  = dll.NewProc("SDL_HapticStopEffect")
	hapticUnpause                     = dll.NewProc("SDL_HapticUnpause")
	hapticUpdateEffect                = dll.NewProc("SDL_HapticUpdateEffect")
	joystickFromInstanceID            = dll.NewProc("SDL_JoystickFromInstanceID")
	joystickOpen                      = dll.NewProc("SDL_JoystickOpen")
	joystickGetAttached               = dll.NewProc("SDL_JoystickGetAttached")
	joystickGetAxis                   = dll.NewProc("SDL_JoystickGetAxis")
	joystickGetAxisInitialState       = dll.NewProc("SDL_JoystickGetAxisInitialState")
	joystickGetBall                   = dll.NewProc("SDL_JoystickGetBall")
	joystickGetButton                 = dll.NewProc("SDL_JoystickGetButton")
	joystickClose                     = dll.NewProc("SDL_JoystickClose")
	joystickCurrentPowerLevel         = dll.NewProc("SDL_JoystickCurrentPowerLevel")
	joystickGetGUID                   = dll.NewProc("SDL_JoystickGetGUID")
	joystickGetHat                    = dll.NewProc("SDL_JoystickGetHat")
	joystickInstanceID                = dll.NewProc("SDL_JoystickInstanceID")
	joystickName                      = dll.NewProc("SDL_JoystickName")
	joystickNumAxes                   = dll.NewProc("SDL_JoystickNumAxes")
	joystickNumBalls                  = dll.NewProc("SDL_JoystickNumBalls")
	joystickNumButtons                = dll.NewProc("SDL_JoystickNumButtons")
	joystickNumHats                   = dll.NewProc("SDL_JoystickNumHats")
	joystickGetProduct                = dll.NewProc("SDL_JoystickGetProduct")
	joystickGetProductVersion         = dll.NewProc("SDL_JoystickGetProductVersion")
	joystickGetType                   = dll.NewProc("SDL_JoystickGetType")
	joystickGetVendor                 = dll.NewProc("SDL_JoystickGetVendor")
	joystickGetDeviceGUID             = dll.NewProc("SDL_JoystickGetDeviceGUID")
	joystickGetGUIDFromString         = dll.NewProc("SDL_JoystickGetGUIDFromString")
	joystickGetDeviceInstanceID       = dll.NewProc("SDL_JoystickGetDeviceInstanceID")
	joystickGetDeviceType             = dll.NewProc("SDL_JoystickGetDeviceType")
	getKeyFromName                    = dll.NewProc("SDL_GetKeyFromName")
	getKeyFromScancode                = dll.NewProc("SDL_GetKeyFromScancode")
	getModState                       = dll.NewProc("SDL_GetModState")
	logGetPriority                    = dll.NewProc("SDL_LogGetPriority")
	createMutex                       = dll.NewProc("SDL_CreateMutex")
	destroyMutex                      = dll.NewProc("SDL_DestroyMutex")
	lockMutex                         = dll.NewProc("SDL_LockMutex")
	tryLockMutex                      = dll.NewProc("SDL_TryLockMutex")
	unlockMutex                       = dll.NewProc("SDL_UnlockMutex")
	allocPalette                      = dll.NewProc("SDL_AllocPalette")
	freePalette                       = dll.NewProc("SDL_FreePalette")
	setPaletteColors                  = dll.NewProc("SDL_SetPaletteColors")
	allocFormat                       = dll.NewProc("SDL_AllocFormat")
	freeFormat                        = dll.NewProc("SDL_FreeFormat")
	setPixelFormatPalette             = dll.NewProc("SDL_SetPixelFormatPalette")
	allocRW                           = dll.NewProc("SDL_AllocRW")
	rwFromFile                        = dll.NewProc("SDL_RWFromFile")
	rwFromMem                         = dll.NewProc("SDL_RWFromMem")
	rwClose                           = dll.NewProc("RWclose")
	freeRW                            = dll.NewProc("SDL_FreeRW")
	loadFile_RW                       = dll.NewProc("SDL_LoadFile_RW")
	readBE16                          = dll.NewProc("SDL_ReadBE16")
	readBE32                          = dll.NewProc("SDL_ReadBE32")
	readBE64                          = dll.NewProc("SDL_ReadBE64")
	readLE16                          = dll.NewProc("SDL_ReadLE16")
	readLE32                          = dll.NewProc("SDL_ReadLE32")
	readLE64                          = dll.NewProc("SDL_ReadLE64")
	readU8                            = dll.NewProc("SDL_ReadU8")
	writeBE16                         = dll.NewProc("SDL_WriteBE16")
	writeBE32                         = dll.NewProc("SDL_WriteBE32")
	writeBE64                         = dll.NewProc("SDL_WriteBE64")
	writeLE16                         = dll.NewProc("SDL_WriteLE16")
	writeLE32                         = dll.NewProc("SDL_WriteLE32")
	writeLE64                         = dll.NewProc("SDL_WriteLE64")
	writeU8                           = dll.NewProc("SDL_WriteU8")
	getDisplayBounds                  = dll.NewProc("SDL_GetDisplayBounds")
	getDisplayUsableBounds            = dll.NewProc("SDL_GetDisplayUsableBounds")
	createRenderer                    = dll.NewProc("SDL_CreateRenderer")
	createSoftwareRenderer            = dll.NewProc("SDL_CreateSoftwareRenderer")
	renderClear                       = dll.NewProc("SDL_RenderClear")
	renderCopy                        = dll.NewProc("SDL_RenderCopy")
	renderCopyEx                      = dll.NewProc("SDL_RenderCopyEx")
	createTexture                     = dll.NewProc("SDL_CreateTexture")
	createTextureFromSurface          = dll.NewProc("SDL_CreateTextureFromSurface")
	destroyRenderer                   = dll.NewProc("SDL_DestroyRenderer")
	renderDrawLine                    = dll.NewProc("SDL_RenderDrawLine")
	renderDrawLines                   = dll.NewProc("SDL_RenderDrawLines")
	renderDrawPoint                   = dll.NewProc("SDL_RenderDrawPoint")
	renderDrawPoints                  = dll.NewProc("SDL_RenderDrawPoints")
	renderDrawRect                    = dll.NewProc("SDL_RenderDrawRect")
	renderDrawRects                   = dll.NewProc("SDL_RenderDrawRects")
	renderFillRect                    = dll.NewProc("SDL_RenderFillRect")
	renderFillRects                   = dll.NewProc("SDL_RenderFillRects")
	renderGetClipRect                 = dll.NewProc("SDL_RenderGetClipRect")
	getRenderDrawBlendMode            = dll.NewProc("SDL_GetRenderDrawBlendMode")
	getRenderDrawColor                = dll.NewProc("SDL_GetRenderDrawColor")
	getRendererInfo                   = dll.NewProc("SDL_GetRendererInfo")
	renderGetLogicalSize              = dll.NewProc("SDL_RenderGetLogicalSize")
	renderGetMetalCommandEncoder      = dll.NewProc("SDL_RenderGetMetalCommandEncoder")
	renderGetMetalLayer               = dll.NewProc("SDL_RenderGetMetalLayer")
	getRendererOutputSize             = dll.NewProc("SDL_GetRendererOutputSize")
	getRenderTarget                   = dll.NewProc("SDL_GetRenderTarget")
	renderGetScale                    = dll.NewProc("SDL_RenderGetScale")
	renderGetViewport                 = dll.NewProc("SDL_RenderGetViewport")
	renderPresent                     = dll.NewProc("SDL_RenderPresent")
	renderReadPixels                  = dll.NewProc("SDL_RenderReadPixels")
	renderTargetSupported             = dll.NewProc("SDL_RenderTargetSupported")
	renderSetClipRect                 = dll.NewProc("SDL_RenderSetClipRect")
	setRenderDrawBlendMode            = dll.NewProc("SDL_SetRenderDrawBlendMode")
	setRenderDrawColor                = dll.NewProc("SDL_SetRenderDrawColor")
	renderSetLogicalSize              = dll.NewProc("SDL_RenderSetLogicalSize")
	setRenderTarget                   = dll.NewProc("SDL_SetRenderTarget")
	renderSetScale                    = dll.NewProc("SDL_RenderSetScale")
	renderSetViewport                 = dll.NewProc("SDL_RenderSetViewport")
	getScancodeFromKey                = dll.NewProc("SDL_GetScancodeFromKey")
	getScancodeFromName               = dll.NewProc("SDL_GetScancodeFromName")
	createSemaphore                   = dll.NewProc("SDL_CreateSemaphore")
	destroySemaphore                  = dll.NewProc("SDL_DestroySemaphore")
	semPost                           = dll.NewProc("SDL_SemPost")
	semTryWait                        = dll.NewProc("SDL_SemTryWait")
	semValue                          = dll.NewProc("SDL_SemValue")
	semWait                           = dll.NewProc("SDL_SemWait")
	semWaitTimeout                    = dll.NewProc("SDL_SemWaitTimeout")
	sensorFromInstanceID              = dll.NewProc("SDL_SensorFromInstanceID")
	sensorOpen                        = dll.NewProc("SDL_SensorOpen")
	sensorClose                       = dll.NewProc("SDL_SensorClose")
	sensorGetData                     = dll.NewProc("SDL_SensorGetData")
	sensorGetInstanceID               = dll.NewProc("SDL_SensorGetInstanceID")
	sensorGetName                     = dll.NewProc("SDL_SensorGetName")
	sensorGetNonPortableType          = dll.NewProc("SDL_SensorGetNonPortableType")
	sensorGetType                     = dll.NewProc("SDL_SensorGetType")
	sensorGetDeviceInstanceID         = dll.NewProc("SDL_SensorGetDeviceInstanceID")
	sensorGetDeviceType               = dll.NewProc("SDL_SensorGetDeviceType")
	loadObject                        = dll.NewProc("SDL_LoadObject")
	loadFunction                      = dll.NewProc("SDL_LoadFunction")
	unloadObject                      = dll.NewProc("SDL_UnloadObject")
	createRGBSurface                  = dll.NewProc("SDL_CreateRGBSurface")
	createRGBSurfaceFrom              = dll.NewProc("SDL_CreateRGBSurfaceFrom")
	createRGBSurfaceWithFormat        = dll.NewProc("SDL_CreateRGBSurfaceWithFormat")
	createRGBSurfaceWithFormatFrom    = dll.NewProc("SDL_CreateRGBSurfaceWithFormatFrom")
	loadBMP_RW                        = dll.NewProc("SDL_LoadBMP_RW")
	blitSurface                       = dll.NewProc("SDL_BlitSurface")
	blitScaled                        = dll.NewProc("SDL_BlitScaled")
	convertSurface                    = dll.NewProc("SDL_ConvertSurface")
	convertSurfaceFormat              = dll.NewProc("SDL_ConvertSurfaceFormat")
	duplicateSurface                  = dll.NewProc("SDL_DuplicateSurface")
	fillRect                          = dll.NewProc("SDL_FillRect")
	fillRects                         = dll.NewProc("SDL_FillRects")
	freeSurface                       = dll.NewProc("SDL_FreeSurface")
	getSurfaceAlphaMod                = dll.NewProc("SDL_GetSurfaceAlphaMod")
	getSurfaceBlendMode               = dll.NewProc("SDL_GetSurfaceBlendMode")
	getClipRect                       = dll.NewProc("SDL_GetClipRect")
	getColorKey                       = dll.NewProc("SDL_GetColorKey")
	getSurfaceColorMod                = dll.NewProc("SDL_GetSurfaceColorMod")
	lockSurface                       = dll.NewProc("SDL_LockSurface")
	lowerBlit                         = dll.NewProc("SDL_LowerBlit")
	lowerBlitScaled                   = dll.NewProc("SDL_LowerBlitScaled")
	saveBMP_RW                        = dll.NewProc("SDL_SaveBMP_RW")
	setSurfaceAlphaMod                = dll.NewProc("SDL_SetSurfaceAlphaMod")
	setSurfaceBlendMode               = dll.NewProc("SDL_SetSurfaceBlendMode")
	setClipRect                       = dll.NewProc("SDL_SetClipRect")
	setColorKey                       = dll.NewProc("SDL_SetColorKey")
	setSurfaceColorMod                = dll.NewProc("SDL_SetSurfaceColorMod")
	setSurfacePalette                 = dll.NewProc("SDL_SetSurfacePalette")
	setSurfaceRLE                     = dll.NewProc("SDL_SetSurfaceRLE")
	softStretch                       = dll.NewProc("SDL_SoftStretch")
	unlockSurface                     = dll.NewProc("SDL_UnlockSurface")
	upperBlit                         = dll.NewProc("SDL_UpperBlit")
	upperBlitScaled                   = dll.NewProc("SDL_UpperBlitScaled")
	destroyTexture                    = dll.NewProc("SDL_DestroyTexture")
	gl_BindTexture                    = dll.NewProc("SDL_GL_BindTexture")
	gl_UnbindTexture                  = dll.NewProc("SDL_GL_UnbindTexture")
	getTextureAlphaMod                = dll.NewProc("SDL_GetTextureAlphaMod")
	getTextureBlendMode               = dll.NewProc("SDL_GetTextureBlendMode")
	lockTexture                       = dll.NewProc("SDL_LockTexture")
	queryTexture                      = dll.NewProc("SDL_QueryTexture")
	setTextureAlphaMod                = dll.NewProc("SDL_SetTextureAlphaMod")
	setTextureBlendMode               = dll.NewProc("SDL_SetTextureBlendMode")
	setTextureColorMod                = dll.NewProc("SDL_SetTextureColorMod")
	unlockTexture                     = dll.NewProc("SDL_UnlockTexture")
	updateTexture                     = dll.NewProc("SDL_UpdateTexture")
	updateYUVTexture                  = dll.NewProc("SDL_UpdateYUVTexture")
	getTouchDevice                    = dll.NewProc("SDL_GetTouchDevice")
	createWindow                      = dll.NewProc("SDL_CreateWindow")
	createWindowFrom                  = dll.NewProc("SDL_CreateWindowFrom")
	getKeyboardFocus                  = dll.NewProc("SDL_GetKeyboardFocus")
	getMouseFocus                     = dll.NewProc("SDL_GetMouseFocus")
	getWindowFromID                   = dll.NewProc("SDL_GetWindowFromID")
	destroyWindow                     = dll.NewProc("SDL_DestroyWindow")
	gl_CreateContext                  = dll.NewProc("SDL_GL_CreateContext")
	gl_GetDrawableSize                = dll.NewProc("SDL_GL_GetDrawableSize")
	gl_MakeCurrent                    = dll.NewProc("SDL_GL_MakeCurrent")
	gl_SwapWindow                     = dll.NewProc("SDL_GL_SwapWindow")
	getWindowBrightness               = dll.NewProc("SDL_GetWindowBrightness")
	getWindowData                     = dll.NewProc("SDL_GetWindowData")
	getWindowDisplayIndex             = dll.NewProc("SDL_GetWindowDisplayIndex")
	getWindowDisplayMode              = dll.NewProc("SDL_GetWindowDisplayMode")
	getWindowFlags                    = dll.NewProc("SDL_GetWindowFlags")
	getWindowGammaRamp                = dll.NewProc("SDL_GetWindowGammaRamp")
	getWindowGrab                     = dll.NewProc("SDL_GetWindowGrab")
	getWindowID                       = dll.NewProc("SDL_GetWindowID")
	getWindowMaximumSize              = dll.NewProc("SDL_GetWindowMaximumSize")
	getWindowMinimumSize              = dll.NewProc("SDL_GetWindowMinimumSize")
	getWindowPixelFormat              = dll.NewProc("SDL_GetWindowPixelFormat")
	getWindowPosition                 = dll.NewProc("SDL_GetWindowPosition")
	getRenderer                       = dll.NewProc("SDL_GetRenderer")
	getWindowSize                     = dll.NewProc("SDL_GetWindowSize")
	getWindowSurface                  = dll.NewProc("SDL_GetWindowSurface")
	getWindowTitle                    = dll.NewProc("SDL_GetWindowTitle")
	getWindowWMInfo                   = dll.NewProc("SDL_GetWindowWMInfo")
	getWindowOpacity                  = dll.NewProc("SDL_GetWindowOpacity")
	hideWindow                        = dll.NewProc("SDL_HideWindow")
	maximizeWindow                    = dll.NewProc("SDL_MaximizeWindow")
	minimizeWindow                    = dll.NewProc("SDL_MinimizeWindow")
	raiseWindow                       = dll.NewProc("SDL_RaiseWindow")
	restoreWindow                     = dll.NewProc("SDL_RestoreWindow")
	setWindowBordered                 = dll.NewProc("SDL_SetWindowBordered")
	setWindowBrightness               = dll.NewProc("SDL_SetWindowBrightness")
	setWindowData                     = dll.NewProc("SDL_SetWindowData")
	setWindowDisplayMode              = dll.NewProc("SDL_SetWindowDisplayMode")
	setWindowFullscreen               = dll.NewProc("SDL_SetWindowFullscreen")
	setWindowGammaRamp                = dll.NewProc("SDL_SetWindowGammaRamp")
	setWindowGrab                     = dll.NewProc("SDL_SetWindowGrab")
	setWindowIcon                     = dll.NewProc("SDL_SetWindowIcon")
	setWindowMaximumSize              = dll.NewProc("SDL_SetWindowMaximumSize")
	setWindowMinimumSize              = dll.NewProc("SDL_SetWindowMinimumSize")
	setWindowPosition                 = dll.NewProc("SDL_SetWindowPosition")
	setWindowResizable                = dll.NewProc("SDL_SetWindowResizable")
	setWindowSize                     = dll.NewProc("SDL_SetWindowSize")
	setWindowTitle                    = dll.NewProc("SDL_SetWindowTitle")
	setWindowOpacity                  = dll.NewProc("SDL_SetWindowOpacity")
	showWindow                        = dll.NewProc("SDL_ShowWindow")
	updateWindowSurface               = dll.NewProc("SDL_UpdateWindowSurface")
	updateWindowSurfaceRects          = dll.NewProc("SDL_UpdateWindowSurfaceRects")
	vulkan_GetDrawableSize            = dll.NewProc("SDL_Vulkan_GetDrawableSize")
	vulkan_GetInstanceExtensions      = dll.NewProc("SDL_Vulkan_GetInstanceExtensions")
	warpMouseInWindow                 = dll.NewProc("SDL_WarpMouseInWindow")
	getYUVConversionMode              = dll.NewProc("SDL_GetYUVConversionMode")
	getYUVConversionModeForResolution = dll.NewProc("SDL_GetYUVConversionModeForResolution")
)

var hintCallbacks = make(map[string]HintCallbackAndData)

// hintCallback returns uintptr because we use it as an argument to
// syscall.NewCallback, which expects the function to return it.
func theHintCallback(userdata, name, oldValue, newValue uintptr) uintptr {
	n := sdlToGoString(name)
	if c, ok := hintCallbacks[n]; ok && c.callback != nil {
		c.callback(c.data, n, sdlToGoString(oldValue), sdlToGoString(newValue))
	}
	return 0
}

var hintCallbackPtr = syscall.NewCallback(theHintCallback)

// AddHintCallback adds a function to watch a particular hint.
// (https://wiki.libsdl.org/SDL_AddHintCallback)
func AddHintCallback(name string, fn HintCallback, data interface{}) {
	hintCallbacks[name] = HintCallbackAndData{
		callback: fn,
		data:     data,
	}
	n := append([]byte(name), 0)
	addHintCallback.Call(
		uintptr(unsafe.Pointer(&n[0])),
		hintCallbackPtr,
		0,
	)
}

// AudioInit initializes a particular audio driver.
// (https://wiki.libsdl.org/SDL_AudioInit)
func AudioInit(driverName string) error {
	d := append([]byte(driverName), 0)
	ret, _, _ := audioInit.Call(uintptr(unsafe.Pointer(&d[0])))
	if ret != 0 {
		return GetError()
	}
	return nil
}

// AudioQuit shuts down audio if you initialized it with AudioInit().
// (https://wiki.libsdl.org/SDL_AudioQuit)
func AudioQuit() {
	audioQuit.Call()
}

// Btoi returns 0 or 1 according to the value of b.
func Btoi(b bool) int {
	if b == true {
		return 1
	}
	return 0
}

// BuildAudioCVT initializes an AudioCVT structure for conversion.
// (https://wiki.libsdl.org/SDL_BuildAudioCVT)
func BuildAudioCVT(
	cvt *AudioCVT,
	srcFormat AudioFormat,
	srcChannels uint8,
	srcRate int,
	dstFormat AudioFormat,
	dstChannels uint8,
	dstRate int,
) (converted bool, err error) {
	ret, _, _ := buildAudioCVT.Call(
		uintptr(unsafe.Pointer(cvt)),
		uintptr(srcFormat),
		uintptr(srcChannels),
		uintptr(srcRate),
		uintptr(dstFormat),
		uintptr(dstChannels),
		uintptr(dstRate),
	)
	if ret == 0 {
		return false, nil
	}
	if ret == 1 {
		return true, nil
	}
	return false, GetError()
}

// Button is used as a mask when testing buttons in buttonstate.
func Button(flag uint32) uint32 {
	return 1 << (flag - 1)
}

// ButtonLMask is used as a mask when testing buttons in buttonstate.
func ButtonLMask() uint32 {
	return Button(BUTTON_LEFT)
}

// ButtonMMask is used as a mask when testing buttons in buttonstate.
func ButtonMMask() uint32 {
	return Button(BUTTON_MIDDLE)
}

// ButtonRMask is used as a mask when testing buttons in buttonstate.
func ButtonRMask() uint32 {
	return Button(BUTTON_RIGHT)
}

// ButtonX1Mask is used as a mask when testing buttons in buttonstate.
func ButtonX1Mask() uint32 {
	return Button(BUTTON_X1)
}

// ButtonX2Mask is used as a mask when testing buttons in buttonstate.
func ButtonX2Mask() uint32 {
	return Button(BUTTON_X2)
}

// COMPILEDVERSION returns the SDL version number that you compiled against.
// (https://wiki.libsdl.org/SDL_COMPILEDVERSION)
func COMPILEDVERSION() int {
	return VERSIONNUM(MAJOR_VERSION, MINOR_VERSION, PATCHLEVEL)
}

// CalculateGammaRamp calculates a 256 entry gamma ramp for a gamma value.
// (https://wiki.libsdl.org/SDL_CalculateGammaRamp)
func CalculateGammaRamp(gamma float32, ramp *[256]uint16) {
	calculateGammaRamp.Call(
		uintptr(gamma),
		uintptr(unsafe.Pointer(ramp)),
	)
}

// CaptureMouse captures the mouse and tracks input outside an SDL window.
// (https://wiki.libsdl.org/SDL_CaptureMouse)
func CaptureMouse(toggle bool) error {
	ret, _, _ := captureMouse.Call(uintptr(Btoi(toggle)))
	if ret != 0 {
		return GetError()
	}
	return nil
}

// ClearError clears any previous error message.
// (https://wiki.libsdl.org/SDL_ClearError)
func ClearError() {
	clearError.Call()
}

// ClearHints clears all hints.
// (https://wiki.libsdl.org/SDL_ClearHints)
func ClearHints() {
	clearHints.Call()
}

// ClearQueuedAudio drops any queued audio data waiting to be sent to the hardware.
// (https://wiki.libsdl.org/SDL_ClearQueuedAudio)
func ClearQueuedAudio(dev AudioDeviceID) {
	clearQueuedAudio.Call(uintptr(dev))
}

// CloseAudio closes the audio device. New programs might want to use CloseAudioDevice() instead.
// (https://wiki.libsdl.org/SDL_CloseAudio)
func CloseAudio() {
	closeAudio.Call()
}

// CloseAudioDevice shuts down audio processing and closes the audio device.
// (https://wiki.libsdl.org/SDL_CloseAudioDevice)
func CloseAudioDevice(dev AudioDeviceID) {
	closeAudioDevice.Call(uintptr(dev))
}

// ConvertAudio converts audio data to a desired audio format.
// (https://wiki.libsdl.org/SDL_ConvertAudio)
func ConvertAudio(cvt *AudioCVT) error {
	ret, _, _ := convertAudio.Call(uintptr(unsafe.Pointer(cvt)))
	if ret != 0 {
		return GetError()
	}
	return nil
}

// ConvertPixels copies a block of pixels of one format to another format.
// (https://wiki.libsdl.org/SDL_ConvertPixels)
func ConvertPixels(
	width, height int32,
	srcFormat uint32,
	src unsafe.Pointer,
	srcPitch int,
	dstFormat uint32,
	dst unsafe.Pointer,
	dstPitch int,
) error {
	ret, _, _ := convertPixels.Call(
		uintptr(width),
		uintptr(height),
		uintptr(srcFormat),
		uintptr(src),
		uintptr(srcPitch),
		uintptr(dstFormat),
		uintptr(dst),
		uintptr(dstPitch),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// CreateWindowAndRenderer returns a new window and default renderer.
// (https://wiki.libsdl.org/SDL_CreateWindowAndRenderer)
func CreateWindowAndRenderer(w, h int32, flags uint32) (*Window, *Renderer, error) {
	var window Window
	var renderer Renderer
	ret, _, _ := createWindowAndRenderer.Call(
		uintptr(w),
		uintptr(h),
		uintptr(flags),
		uintptr(unsafe.Pointer(&window)),
		uintptr(unsafe.Pointer(&renderer)),
	)
	if ret == ^uintptr(0) {
		return nil, nil, GetError()
	}
	return &window, &renderer, nil
}

// DelEventWatch removes an event watch callback added with AddEventWatch().
// (https://wiki.libsdl.org/SDL_DelEventWatch)
func DelEventWatch(handle EventWatchHandle) {
	context, ok := eventWatches[handle]
	if !ok {
		return
	}
	delete(eventWatches, context.handle)
	delEventWatch.Call(
		eventFilterCallbackPtr,
		uintptr(context.handle),
	)
}

// DelHintCallback removes a function watching a particular hint.
// (https://wiki.libsdl.org/SDL_DelHintCallback)
func DelHintCallback(name string) {
	delete(hintCallbacks, name)
	n := append([]byte(name), 0)
	delHintCallback.Call(
		uintptr(unsafe.Pointer(&n[0])),
		hintCallbackPtr,
		0,
	)
}

// Delay waits a specified number of milliseconds before returning.
// (https://wiki.libsdl.org/SDL_Delay)
func Delay(ms uint32) {
	delay.Call(uintptr(ms))
}

// DequeueAudio dequeues more audio on non-callback devices.
// (https://wiki.libsdl.org/SDL_DequeueAudio)
func DequeueAudio(dev AudioDeviceID, data []byte) error {
	ret, _, _ := dequeueAudio.Call(
		uintptr(dev),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// DisableScreenSaver prevents the screen from being blanked by a screen saver.
// (https://wiki.libsdl.org/SDL_DisableScreenSaver)
func DisableScreenSaver() {
	disableScreenSaver.Call()
}

// Do the specified function in the main thread.
// For this function to work, you must have correctly used sdl.Main(..) in your
// main() function. Calling this function before/without sdl.Main(..) will cause
// a panic.
func Do(f func()) {
	callInMain(f)
}

// Calls a function in the main thread. It is only properly initialized inside
// sdl.Main(..). As a default, it panics. It is used by sdl.Do(..) below.
var callInMain = func(f func()) {
	panic("sdl.Main(main func()) must be called before sdl.Do(f func())")
}

// EnableScreenSaver allows the screen to be blanked by a screen saver.
// (https://wiki.libsdl.org/SDL_EnableScreenSaver)
func EnableScreenSaver() {
	enableScreenSaver.Call()
}

// Error sets the SDL error message to the specified error code.
func Error(code ErrorCode) {
	sdlError.Call(uintptr(code))
}

// EventState sets the state of processing events by type.
// (https://wiki.libsdl.org/SDL_EventState)
func EventState(typ uint32, state int) uint8 {
	ret, _, _ := eventState.Call(uintptr(typ), uintptr(state))
	return uint8(ret)
}

// FilterEvents run a specific filter function on the current event queue, removing any events for which the filter returns 0.
// (https://wiki.libsdl.org/SDL_FilterEvents)
func FilterEvents(filter EventFilter, userdata interface{}) {
	context := newEventFilterCallbackContext(filter, userdata)
	filterEvents.Call(
		eventFilterCallbackPtr,
		uintptr(context.handle),
	)
}

// FilterEventsFunc run a specific function on the current event queue, removing any events for which the filter returns 0.
// (https://wiki.libsdl.org/SDL_FilterEvents)
func FilterEventsFunc(filter eventFilterFunc, userdata interface{}) {
	FilterEvents(filter, userdata)
}

type eventFilterFunc func(Event, interface{}) bool

func (ef eventFilterFunc) FilterEvent(e Event, userdata interface{}) bool {
	return ef(e, userdata)
}

// FlushEvent clears events from the event queue.
// (https://wiki.libsdl.org/SDL_FlushEvent)
func FlushEvent(typ uint32) {
	flushEvent.Call(uintptr(typ))
}

// FlushEvents clears events from the event queue.
// (https://wiki.libsdl.org/SDL_FlushEvents)
func FlushEvents(minType, maxType uint32) {
	flushEvents.Call(uintptr(minType), uintptr(maxType))
}

// FreeCursor frees a cursor created with CreateCursor(), CreateColorCursor() or CreateSystemCursor().
// (https://wiki.libsdl.org/SDL_FreeCursor)
func FreeCursor(cursor *Cursor) {
	freeCursor.Call(uintptr(unsafe.Pointer(cursor)))
}

// FreeWAV frees data previously allocated with LoadWAV() or LoadWAVRW().
// (https://wiki.libsdl.org/SDL_FreeWAV)
func FreeWAV(audioBuf []uint8) {
	freeWAV.Call(uintptr(unsafe.Pointer(&audioBuf[0])))
}

// GLDeleteContext deletes an OpenGL context.
// (https://wiki.libsdl.org/SDL_GL_DeleteContext)
func GLDeleteContext(context GLContext) {
	gl_DeleteContext.Call(uintptr(context))
}

// GLExtensionSupported reports whether an OpenGL extension is supported for the current context.
// (https://wiki.libsdl.org/SDL_GL_ExtensionSupported)
func GLExtensionSupported(extension string) bool {
	e := append([]byte(extension), 0)
	ret, _, _ := gl_ExtensionSupported.Call(uintptr(unsafe.Pointer(&e[0])))
	return ret != 0
}

// GLGetAttribute returns the actual value for an attribute from the current context.
// (https://wiki.libsdl.org/SDL_GL_GetAttribute)
func GLGetAttribute(attr GLattr) (int, error) {
	var value int
	ret, _, _ := gl_GetAttribute.Call(uintptr(attr), uintptr(unsafe.Pointer(&value)))
	if ret != 0 {
		return value, GetError()
	}
	return value, nil
}

// GLGetProcAddress returns an OpenGL function by name.
// (https://wiki.libsdl.org/SDL_GL_GetProcAddress)
func GLGetProcAddress(proc string) unsafe.Pointer {
	p := append([]byte(proc), 0)
	ret, _, _ := gl_GetProcAddress.Call(uintptr(unsafe.Pointer(&p[0])))
	return unsafe.Pointer(ret)
}

// GLGetSwapInterval returns the swap interval for the current OpenGL context.
// (https://wiki.libsdl.org/SDL_GL_GetSwapInterval)
func GLGetSwapInterval() (int, error) {
	ret, _, _ := gl_GetSwapInterval.Call()
	return int(ret), errorFromInt(int(ret))
}

// errorFromInt returns GetError() if passed negative value, otherwise it returns nil.
func errorFromInt(code int) error {
	if code < 0 {
		return GetError()
	}
	return nil
}

// GLLoadLibrary dynamically loads an OpenGL library.
// (https://wiki.libsdl.org/SDL_GL_LoadLibrary)
func GLLoadLibrary(path string) error {
	p := append([]byte(path), 0)
	ret, _, _ := gl_LoadLibrary.Call(uintptr(unsafe.Pointer(&p[0])))
	return errorFromInt(int(ret))
}

// GLSetAttribute sets an OpenGL window attribute before window creation.
// (https://wiki.libsdl.org/SDL_GL_SetAttribute)
func GLSetAttribute(attr GLattr, value int) error {
	ret, _, _ := gl_SetAttribute.Call(uintptr(attr), uintptr(value))
	return errorFromInt(int(ret))
}

// GLSetSwapInterval sets the swap interval for the current OpenGL context.
// (https://wiki.libsdl.org/SDL_GL_SetSwapInterval)
func GLSetSwapInterval(interval int) error {
	ret, _, _ := gl_SetSwapInterval.Call(uintptr(interval))
	return errorFromInt(int(ret))
}

// GLUnloadLibrary unloads the OpenGL library previously loaded by GLLoadLibrary().
// (https://wiki.libsdl.org/SDL_GL_UnloadLibrary)
func GLUnloadLibrary() {
	gl_UnloadLibrary.Call()
}

// GameControllerAddMapping adds support for controllers that SDL is unaware of or to cause an existing controller to have a different binding.
// (https://wiki.libsdl.org/SDL_GameControllerAddMapping)
func GameControllerAddMapping(mappingString string) int {
	m := append([]byte(mappingString), 0)
	ret, _, _ := gameControllerAddMapping.Call(uintptr(unsafe.Pointer(&m[0])))
	return int(ret)
}

// GameControllerEventState returns the current state of, enable, or disable events dealing with Game Controllers. This will not disable Joystick events, which can also be fired by a controller (see https://wiki.libsdl.org/SDL_JoystickEventState).
// (https://wiki.libsdl.org/SDL_GameControllerEventState)
func GameControllerEventState(state int) int {
	ret, _, _ := gameControllerEventState.Call(uintptr(state))
	return int(ret)
}

// GameControllerGetStringForAxis converts from an axis enum to a string.
// (https://wiki.libsdl.org/SDL_GameControllerGetStringForAxis)
func GameControllerGetStringForAxis(axis GameControllerAxis) string {
	ret, _, _ := gameControllerGetStringForAxis.Call(uintptr(axis))
	return sdlToGoString(ret)
}

// GameControllerGetStringForButton turns a button enum into a string mapping.
// (https://wiki.libsdl.org/SDL_GameControllerGetStringForButton)
func GameControllerGetStringForButton(btn GameControllerButton) string {
	ret, _, _ := gameControllerGetStringForButton.Call(uintptr(btn))
	return sdlToGoString(ret)
}

// GameControllerMappingForGUID returns the game controller mapping string for a
// given GUID.
//(https://wiki.libsdl.org/SDL_GameControllerMappingForGUID)
func GameControllerMappingForGUID(guid JoystickGUID) string {
	//	mappingString := C.SDL_GameControllerMappingForGUID(guid.c())
	//defer C.free(unsafe.Pointer(mappingString))
	//return C.GoString(mappingString)
	return "" // TODO
}

// GameControllerMappingForIndex returns the game controller mapping string at a
// particular index.
func GameControllerMappingForIndex(index int) string {
	ret, _, _ := gameControllerMappingForIndex.Call(uintptr(index))
	return sdlToGoString(ret)
}

// GameControllerNameForIndex returns the implementation dependent name for the game controller.
// (https://wiki.libsdl.org/SDL_GameControllerNameForIndex)
func GameControllerNameForIndex(index int) string {
	ret, _, _ := gameControllerNameForIndex.Call(uintptr(index))
	return sdlToGoString(ret)
}

// GameControllerNumMappings returns the number of mappings installed.
func GameControllerNumMappings() int {
	ret, _, _ := gameControllerNumMappings.Call()
	return int(ret)
}

// GameControllerUpdate manually pumps game controller updates if not using the loop.
// (https://wiki.libsdl.org/SDL_GameControllerUpdate)
func GameControllerUpdate() {
	gameControllerUpdate.Call()
}

// GetAudioDeviceName returns the name of a specific audio device.
// (https://wiki.libsdl.org/SDL_GetAudioDeviceName)
func GetAudioDeviceName(index int, isCapture bool) string {
	ret, _, _ := getAudioDeviceName.Call(uintptr(index), uintptr(Btoi(isCapture)))
	return sdlToGoString(ret)
}

// GetAudioDriver returns the name of a built in audio driver.
// (https://wiki.libsdl.org/SDL_GetAudioDriver)
func GetAudioDriver(index int) string {
	ret, _, _ := getAudioDriver.Call(uintptr(index))
	return sdlToGoString(ret)
}

// GetBasePath returns the directory where the application was run from. This is where the application data directory is.
// (https://wiki.libsdl.org/SDL_GetBasePath)
func GetBasePath() string {
	ret, _, _ := getBasePath.Call()
	return sdlToGoString(ret)
}

// GetCPUCacheLineSize returns the L1 cache line size of the CPU.
// (https://wiki.libsdl.org/SDL_GetCPUCacheLineSize)
func GetCPUCacheLineSize() int {
	ret, _, _ := getCPUCacheLineSize.Call()
	return int(ret)
}

// GetCPUCount returns the number of CPU cores available.
// (https://wiki.libsdl.org/SDL_GetCPUCount)
func GetCPUCount() int {
	ret, _, _ := getCPUCount.Call()
	return int(ret)
}

// GetClipboardText returns UTF-8 text from the clipboard.
// (https://wiki.libsdl.org/SDL_GetClipboardText)
func GetClipboardText() (string, error) {
	ret, _, _ := getClipboardText.Call()
	if ret == 0 {
		return "", GetError()
	}
	return sdlToGoString(ret), nil
}

// GetCurrentAudioDriver returns the name of the current audio driver.
// (https://wiki.libsdl.org/SDL_GetCurrentAudioDriver)
func GetCurrentAudioDriver() string {
	ret, _, _ := getCurrentAudioDriver.Call()
	return sdlToGoString(ret)
}

// GetCurrentVideoDriver returns the name of the currently initialized video driver.
// (https://wiki.libsdl.org/SDL_GetCurrentVideoDriver)
func GetCurrentVideoDriver() (string, error) {
	ret, _, _ := getCurrentVideoDriver.Call()
	if ret == 0 {
		return "", GetError()
	}
	return sdlToGoString(ret), nil
}

// GetDisplayDPI returns the dots/pixels-per-inch for a display.
// (https://wiki.libsdl.org/SDL_GetDisplayDPI)
func GetDisplayDPI(displayIndex int) (ddpi, hdpi, vdpi float32, err error) {
	ret, _, _ := getDisplayDPI.Call(
		uintptr(displayIndex),
		uintptr(unsafe.Pointer(&ddpi)),
		uintptr(unsafe.Pointer(&hdpi)),
		uintptr(unsafe.Pointer(&vdpi)),
	)
	err = errorFromInt(int(ret))
	return
}

// GetDisplayName returns the name of a display in UTF-8 encoding.
// (https://wiki.libsdl.org/SDL_GetDisplayName)
func GetDisplayName(displayIndex int) (string, error) {
	ret, _, _ := getDisplayName.Call(uintptr(displayIndex))
	if ret == 0 {
		return "", GetError()
	}
	return sdlToGoString(ret), nil
}

// GetError returns the last error that occurred, or an empty string if there hasn't been an error message set since the last call to ClearError().
// (https://wiki.libsdl.org/SDL_GetError)
func GetError() error {
	ret, _, _ := getError.Call()
	if ret != 0 {
		s := sdlToGoString(ret)
		// SDL_GetError returns "an empty string if there hasn't been an error message"
		if s != "" {
			return errors.New(s)
		}
	}
	return nil
}

// GetEventState returns the current processing state of the specified event
// (https://wiki.libsdl.org/SDL_EventState)
func GetEventState(typ uint32) uint8 {
	ret, _, _ := eventState.Call(uintptr(typ), ^uintptr(0) /* == QUERY */)
	return uint8(ret)
}

// GetHint returns the value of a hint.
// (https://wiki.libsdl.org/SDL_GetHint)
func GetHint(name string) string {
	n := append([]byte(name), 0)
	ret, _, _ := getHint.Call(uintptr(unsafe.Pointer(&n[0])))
	return sdlToGoString(ret)
}

// GetKeyName returns a human-readable name for a key.
// (https://wiki.libsdl.org/SDL_GetKeyName)
func GetKeyName(code Keycode) string {
	ret, _, _ := getKeyName.Call(uintptr(code))
	return sdlToGoString(ret)
}

// GetKeyboardState returns a snapshot of the current state of the keyboard.
// (https://wiki.libsdl.org/SDL_GetKeyboardState)
func GetKeyboardState() []uint8 {
	var numkeys int
	start, _, _ := getKeyboardState.Call(uintptr(unsafe.Pointer(&numkeys)))
	keys := reflect.SliceHeader{}
	keys.Len = int(numkeys)
	keys.Cap = int(numkeys)
	keys.Data = uintptr(unsafe.Pointer(start))
	return *(*[]uint8)(unsafe.Pointer(&keys))
}

// GetMouseState returns the current state of the mouse.
// (https://wiki.libsdl.org/SDL_GetMouseState)
func GetMouseState() (x, y int32, state uint32) {
	ret, _, _ := getMouseState.Call(
		uintptr(unsafe.Pointer(&x)),
		uintptr(unsafe.Pointer(&y)),
	)
	state = uint32(ret)
	return
}

// GetNumAudioDevices returns the number of built-in audio devices.
// (https://wiki.libsdl.org/SDL_GetNumAudioDevices)
func GetNumAudioDevices(isCapture bool) int {
	ret, _, _ := getNumAudioDevices.Call(uintptr(Btoi(isCapture)))
	return int(ret)
}

// GetNumAudioDrivers returns the number of built-in audio drivers.
// (https://wiki.libsdl.org/SDL_GetNumAudioDrivers)
func GetNumAudioDrivers() int {
	ret, _, _ := getNumAudioDrivers.Call()
	return int(ret)
}

// GetNumDisplayModes returns the number of available display modes.
// (https://wiki.libsdl.org/SDL_GetNumDisplayModes)
func GetNumDisplayModes(displayIndex int) (int, error) {
	ret, _, _ := getNumDisplayModes.Call(uintptr(displayIndex))
	return int(ret), errorFromInt(int(ret))
}

// GetNumRenderDrivers returns the number of 2D rendering drivers available for the current display.
// (https://wiki.libsdl.org/SDL_GetNumRenderDrivers)
func GetNumRenderDrivers() (int, error) {
	ret, _, _ := getNumRenderDrivers.Call()
	return int(ret), errorFromInt(int(ret))
}

// GetNumTouchDevices returns the number of registered touch devices.
// (https://wiki.libsdl.org/SDL_GetNumTouchDevices)
func GetNumTouchDevices() int {
	ret, _, _ := getNumTouchDevices.Call()
	return int(ret)
}

// GetNumTouchFingers returns the number of active fingers for a given touch device.
// (https://wiki.libsdl.org/SDL_GetNumTouchFingers)
func GetNumTouchFingers(t TouchID) int {
	// TODO TouchID is an int64, this will probably not work on a 32 bit OS
	ret, _, _ := getNumTouchFingers.Call(uintptr(t))
	return int(ret)
}

// GetNumVideoDisplays returns the number of available video displays.
// (https://wiki.libsdl.org/SDL_GetNumVideoDisplays)
func GetNumVideoDisplays() (int, error) {
	ret, _, _ := getNumVideoDisplays.Call()
	return int(ret), errorFromInt(int(ret))
}

// GetNumVideoDrivers returns the number of video drivers compiled into SDL.
// (https://wiki.libsdl.org/SDL_GetNumVideoDrivers)
func GetNumVideoDrivers() (int, error) {
	ret, _, _ := getNumVideoDrivers.Call()
	return int(ret), errorFromInt(int(ret))
}

// GetPerformanceCounter returns the current value of the high resolution counter.
// (https://wiki.libsdl.org/SDL_GetPerformanceCounter)
func GetPerformanceCounter() uint64 {
	// TODO what about 32 bit OS?
	ret, _, _ := getPerformanceCounter.Call()
	return uint64(ret)
}

// GetPerformanceFrequency returns the count per second of the high resolution counter.
// (https://wiki.libsdl.org/SDL_GetPerformanceFrequency)
func GetPerformanceFrequency() uint64 {
	// TODO what about 32 bit OS?
	ret, _, _ := getPerformanceFrequency.Call()
	return uint64(ret)
}

// GetPixelFormatName returns the human readable name of a pixel format.
// (https://wiki.libsdl.org/SDL_GetPixelFormatName)
func GetPixelFormatName(format uint) string {
	ret, _, _ := getPixelFormatName.Call(uintptr(format))
	return sdlToGoString(ret)
}

// GetPlatform returns the name of the platform.
// (https://wiki.libsdl.org/SDL_GetPlatform)
func GetPlatform() string {
	ret, _, _ := getPlatform.Call()
	return sdlToGoString(ret)
}

// GetPowerInfo returns the current power supply details.
// (https://wiki.libsdl.org/SDL_GetPowerInfo)
func GetPowerInfo() (state, secs, percent int) {
	ret, _, _ := getPowerInfo.Call(
		uintptr(unsafe.Pointer(&secs)),
		uintptr(unsafe.Pointer(&percent)),
	)
	state = int(ret)
	return
}

// GetPrefPath returns the "pref dir". This is meant to be where the application can write personal files (Preferences and save games, etc.) that are specific to the application. This directory is unique per user and per application.
// (https://wiki.libsdl.org/SDL_GetPrefPath)
func GetPrefPath(org, app string) string {
	o := append([]byte(org), 0)
	a := append([]byte(app), 0)
	ret, _, _ := getPrefPath.Call(
		uintptr(unsafe.Pointer(&o[0])),
		uintptr(unsafe.Pointer(&a[0])),
	)
	return sdlToGoString(ret)
}

// GetQueuedAudioSize returns the number of bytes of still-queued audio.
// (https://wiki.libsdl.org/SDL_GetQueuedAudioSize)
func GetQueuedAudioSize(dev AudioDeviceID) uint32 {
	ret, _, _ := getQueuedAudioSize.Call(uintptr(dev))
	return uint32(ret)
}

// GetRGB returns RGB values from a pixel in the specified format.
// (https://wiki.libsdl.org/SDL_GetRGB)
func GetRGB(pixel uint32, format *PixelFormat) (r, g, b uint8) {
	getRGB.Call(
		uintptr(pixel),
		uintptr(unsafe.Pointer(&format)),
		uintptr(unsafe.Pointer(&r)),
		uintptr(unsafe.Pointer(&g)),
		uintptr(unsafe.Pointer(&b)),
	)
	return
}

// GetRGBA returns RGBA values from a pixel in the specified format.
// (https://wiki.libsdl.org/SDL_GetRGBA)
func GetRGBA(pixel uint32, format *PixelFormat) (r, g, b, a uint8) {
	getRGBA.Call(
		uintptr(pixel),
		uintptr(unsafe.Pointer(&format)),
		uintptr(unsafe.Pointer(&r)),
		uintptr(unsafe.Pointer(&g)),
		uintptr(unsafe.Pointer(&b)),
		uintptr(unsafe.Pointer(&a)),
	)
	return
}

// GetRelativeMouseMode reports where relative mouse mode is enabled.
// (https://wiki.libsdl.org/SDL_GetRelativeMouseMode)
func GetRelativeMouseMode() bool {
	ret, _, _ := getRelativeMouseMode.Call()
	return ret != 0
}

// GetRelativeMouseState returns the relative state of the mouse.
// (https://wiki.libsdl.org/SDL_GetRelativeMouseState)
func GetRelativeMouseState() (x, y int32, state uint32) {
	ret, _, _ := getRelativeMouseState.Call(
		uintptr(unsafe.Pointer(&x)),
		uintptr(unsafe.Pointer(&y)),
	)
	state = uint32(ret)
	return
}

// GetRenderDriverInfo returns information about a specific 2D rendering driver for the current display.
// (https://wiki.libsdl.org/SDL_GetRenderDriverInfo)
func GetRenderDriverInfo(index int, info *RendererInfo) (int, error) {
	var cInfo struct {
		name uintptr
		RendererInfoData
	}
	ret, _, _ := getRenderDriverInfo.Call(
		uintptr(index),
		uintptr(unsafe.Pointer(&cInfo)),
	)
	if ret != 0 {
		return int(ret), GetError()
	}
	info.Name = sdlToGoString(cInfo.name)
	info.RendererInfoData = cInfo.RendererInfoData
	return 0, nil
}

// GetRevision returns the code revision of SDL that is linked against your program.
// (https://wiki.libsdl.org/SDL_GetRevision)
func GetRevision() string {
	ret, _, _ := getRevision.Call()
	return sdlToGoString(ret)
}

// GetRevisionNumber returns the revision number of SDL that is linked against your program.
// (https://wiki.libsdl.org/SDL_GetRevisionNumber)
func GetRevisionNumber() int {
	ret, _, _ := getRevisionNumber.Call()
	return int(ret)
}

// GetScancodeName returns a human-readable name for a scancode
// (https://wiki.libsdl.org/SDL_GetScancodeName)
func GetScancodeName(code Scancode) string {
	ret, _, _ := getScancodeName.Call(uintptr(code))
	return sdlToGoString(ret)
}

// GetSystemRAM returns the amount of RAM configured in the system.
// (https://wiki.libsdl.org/SDL_GetSystemRAM)
func GetSystemRAM() int {
	ret, _, _ := getSystemRAM.Call()
	return int(ret)
}

// GetTicks returns the number of milliseconds since the SDL library initialization.
// (https://wiki.libsdl.org/SDL_GetTicks)
func GetTicks() uint32 {
	ret, _, _ := getTicks.Call()
	return uint32(ret)
}

// GetVersion returns the version of SDL that is linked against your program.
// (https://wiki.libsdl.org/SDL_GetVersion)
func GetVersion(v *Version) {
	getVersion.Call(uintptr(unsafe.Pointer(v)))
}

// GetVideoDriver returns the name of a built in video driver.
// (https://wiki.libsdl.org/SDL_GetVideoDriver)
func GetVideoDriver(index int) string {
	ret, _, _ := getVideoDriver.Call(uintptr(index))
	return sdlToGoString(ret)
}

// HapticIndex returns the index of a haptic device.
// (https://wiki.libsdl.org/SDL_HapticIndex)
func HapticIndex(h *Haptic) (int, error) {
	ret, _, _ := hapticIndex.Call(uintptr(unsafe.Pointer(h)))
	return int(ret), errorFromInt(int(ret))
}

// HapticName returns the implementation dependent name of a haptic device.
// (https://wiki.libsdl.org/SDL_HapticName)
func HapticName(index int) (string, error) {
	ret, _, _ := hapticName.Call(uintptr(index))
	if ret == 0 {
		return "", GetError()
	}
	return sdlToGoString(ret), nil
}

// HapticOpened reports whether the haptic device at the designated index has been opened.
// (https://wiki.libsdl.org/SDL_HapticOpened)
func HapticOpened(index int) (bool, error) {
	ret, _, _ := hapticOpened.Call(uintptr(index))
	if ret == 0 {
		return false, GetError()
	}
	return ret == 1, nil
}

// Has3DNow reports whether the CPU has 3DNow! features.
// (https://wiki.libsdl.org/SDL_Has3DNow)
func Has3DNow() bool {
	ret, _, _ := has3DNow.Call()
	return ret > 0
}

// HasAVX reports whether the CPU has AVX features.
// (https://wiki.libsdl.org/SDL_HasAVX)
func HasAVX() bool {
	ret, _, _ := hasAVX.Call()
	return ret > 0
}

// HasAVX2 reports whether the CPU has AVX2 features.
// (https://wiki.libsdl.org/SDL_HasAVX2)
func HasAVX2() bool {
	ret, _, _ := hasAVX2.Call()
	return ret > 0
}

// HasAltiVec reports whether the CPU has AltiVec features.
// (https://wiki.libsdl.org/SDL_HasAltiVec)
func HasAltiVec() bool {
	ret, _, _ := hasAltiVec.Call()
	return ret > 0
}

// HasClipboardText reports whether the clipboard exists and contains a text string that is non-empty.
// (https://wiki.libsdl.org/SDL_HasClipboardText)
func HasClipboardText() bool {
	ret, _, _ := hasClipboardText.Call()
	return ret > 0
}

// HasEvent checks for the existence of certain event types in the event queue.
// (https://wiki.libsdl.org/SDL_HasEvent)
func HasEvent(type_ uint32) bool {
	ret, _, _ := hasEvent.Call()
	return ret > 0
}

// HasEvents checks for the existence of a range of event types in the event queue.
// (https://wiki.libsdl.org/SDL_HasEvents)
func HasEvents(minType, maxType uint32) bool {
	ret, _, _ := hasEvents.Call(uintptr(minType), uintptr(maxType))
	return ret != 0
}

// HasMMX reports whether the CPU has MMX features.
// (https://wiki.libsdl.org/SDL_HasMMX)
func HasMMX() bool {
	ret, _, _ := hasMMX.Call()
	return ret > 0
}

// HasNEON reports whether the CPU has NEON features.
// (https://wiki.libsdl.org/SDL_HasNEON)
func HasNEON() bool {
	ret, _, _ := hasNEON.Call()
	return ret > 0
}

// HasRDTSC reports whether the CPU has the RDTSC instruction.
// (https://wiki.libsdl.org/SDL_HasRDTSC)
func HasRDTSC() bool {
	ret, _, _ := hasRDTSC.Call()
	return ret > 0
}

// HasSSE reports whether the CPU has SSE features.
// (https://wiki.libsdl.org/SDL_HasSSE)
func HasSSE() bool {
	ret, _, _ := hasSSE.Call()
	return ret > 0
}

// HasSSE2 reports whether the CPU has SSE2 features.
// (https://wiki.libsdl.org/SDL_HasSSE2)
func HasSSE2() bool {
	ret, _, _ := hasSSE2.Call()
	return ret > 0
}

// HasSSE3 reports whether the CPU has SSE3 features.
// (https://wiki.libsdl.org/SDL_HasSSE3)
func HasSSE3() bool {
	ret, _, _ := hasSSE3.Call()
	return ret > 0
}

// HasSSE41 reports whether the CPU has SSE4.1 features.
// (https://wiki.libsdl.org/SDL_HasSSE41)
func HasSSE41() bool {
	ret, _, _ := hasSSE41.Call()
	return ret > 0
}

// HasSSE42 reports whether the CPU has SSE4.2 features.
// (https://wiki.libsdl.org/SDL_HasSSE42)
func HasSSE42() bool {
	ret, _, _ := hasSSE42.Call()
	return ret > 0
}

// HasScreenKeyboardSupport reports whether the platform has some screen keyboard support.
// (https://wiki.libsdl.org/SDL_HasScreenKeyboardSupport)
func HasScreenKeyboardSupport() bool {
	ret, _, _ := hasScreenKeyboardSupport.Call()
	return ret > 0
}

// Init initialize the SDL library. This must be called before using most other SDL functions.
// (https://wiki.libsdl.org/SDL_Init)
func Init(flags uint32) error {
	ret, _, _ := sdlInit.Call(uintptr(flags))
	if ret != 0 {
		return GetError()
	}
	return nil
}

// InitSubSystem initializes specific SDL subsystems.
// (https://wiki.libsdl.org/SDL_InitSubSystem)
func InitSubSystem(flags uint32) error {
	ret, _, _ := initSubSystem.Call(uintptr(flags))
	if ret != 0 {
		return GetError()
	}
	return nil
}

// IsGameController reports whether the given joystick is supported by the game controller interface.
// (https://wiki.libsdl.org/SDL_IsGameController)
func IsGameController(index int) bool {
	ret, _, _ := isGameController.Call(uintptr(index))
	return ret != 0
}

// IsScreenKeyboardShown reports whether the screen keyboard is shown for given window.
// (https://wiki.libsdl.org/SDL_IsScreenKeyboardShown)
func IsScreenKeyboardShown(window *Window) bool {
	ret, _, _ := isScreenKeyboardShown.Call(uintptr(unsafe.Pointer(window)))
	return ret > 0
}

// IsScreenSaverEnabled reports whether the screensaver is currently enabled.
// (https://wiki.libsdl.org/SDL_IsScreenSaverEnabled)
func IsScreenSaverEnabled() bool {
	ret, _, _ := isScreenSaverEnabled.Call()
	return ret != 0
}

// IsTextInputActive checks whether or not Unicode text input events are enabled.
// (https://wiki.libsdl.org/SDL_IsTextInputActive)
func IsTextInputActive() bool {
	ret, _, _ := isTextInputActive.Call()
	return ret > 0
}

// JoystickEventState enables or disables joystick event polling.
// (https://wiki.libsdl.org/SDL_JoystickEventState)
func JoystickEventState(state int) int {
	ret, _, _ := joystickEventState.Call(uintptr(state))
	return int(ret)
}

// JoystickGetDeviceProduct returns the USB product ID of a joystick, if
// available, 0 otherwise.
func JoystickGetDeviceProduct(index int) int {
	ret, _, _ := joystickGetDeviceProduct.Call(uintptr(index))
	return int(ret)
}

// JoystickGetDeviceProductVersion returns the product version of a joystick, if
// available, 0 otherwise.
func JoystickGetDeviceProductVersion(index int) int {
	ret, _, _ := joystickGetDeviceProductVersion.Call(uintptr(index))
	return int(ret)
}

// JoystickGetDeviceVendor returns the USB vendor ID of a joystick, if
// available, 0 otherwise.
func JoystickGetDeviceVendor(index int) int {
	ret, _, _ := joystickGetDeviceVendor.Call(uintptr(index))
	return int(ret)
}

// JoystickGetGUIDString returns an ASCII string representation for a given JoystickGUID.
// (https://wiki.libsdl.org/SDL_JoystickGetGUIDString)
func JoystickGetGUIDString(guid JoystickGUID) string {
	return "" // TODO
	//_pszGUID := make([]rune, 1024)
	//pszGUID := C.CString(string(_pszGUID[:]))
	//defer C.free(unsafe.Pointer(pszGUID))
	//C.SDL_JoystickGetGUIDString(guid.c(), pszGUID, C.int(unsafe.Sizeof(_pszGUID)))
	//return C.GoString(pszGUID)
}

// JoystickIsHaptic reports whether a joystick has haptic features.
// (https://wiki.libsdl.org/SDL_JoystickIsHaptic)
func JoystickIsHaptic(joy *Joystick) (bool, error) {
	ret, _, _ := joystickIsHaptic.Call(uintptr(unsafe.Pointer(joy)))
	return ret != 0, errorFromInt(int(ret))
}

// JoystickNameForIndex returns the implementation dependent name of a joystick.
// (https://wiki.libsdl.org/SDL_JoystickNameForIndex)
func JoystickNameForIndex(index int) string {
	ret, _, _ := joystickNameForIndex.Call(uintptr(index))
	return sdlToGoString(ret)
}

// JoystickUpdate updates the current state of the open joysticks.
// (https://wiki.libsdl.org/SDL_JoystickUpdate)
func JoystickUpdate() {
	joystickUpdate.Call()
}

// LoadDollarTemplates loads Dollar Gesture templates from a file.
// (https://wiki.libsdl.org/SDL_LoadDollarTemplates)
func LoadDollarTemplates(t TouchID, src *RWops) int {
	// TODO passing int64 as uintptr, does it work on 32 bit OS?
	ret, _, _ := loadDollarTemplates.Call(
		uintptr(t),
		uintptr(unsafe.Pointer(src)),
	)
	return int(ret)
}

// LoadFile loads an entire file
// (https://wiki.libsdl.org/SDL_LoadFile)
func LoadFile(file string) (data []byte, size int) {
	return RWFromFile(file, "rb").LoadFileRW(true)
}

// LockAudio locks the audio device. New programs might want to use LockAudioDevice() instead.
// (https://wiki.libsdl.org/SDL_LockAudio)
func LockAudio() {
	lockAudio.Call()
}

// LockAudioDevice locks out the audio callback function for a specified device.
// (https://wiki.libsdl.org/SDL_LockAudioDevice)
func LockAudioDevice(dev AudioDeviceID) {
	lockAudioDevice.Call(uintptr(dev))
}

// LockJoysticks locks joysticks for multi-threaded access to the joystick API
// TODO: (https://wiki.libsdl.org/SDL_LockJoysticks)
func LockJoysticks() {
	lockJoysticks.Call()
}

// Log logs a message with LOG_CATEGORY_APPLICATION and LOG_PRIORITY_INFO.
// (https://wiki.libsdl.org/SDL_Log)
func Log(str string, args ...interface{}) {
	s := append([]byte(fmt.Sprintf(str, args...)), 0)
	log.Call(uintptr(unsafe.Pointer(&s[0])))
}

// LogCritical logs a message with LOG_PRIORITY_CRITICAL.
// (https://wiki.libsdl.org/SDL_LogCritical)
func LogCritical(category int, str string, args ...interface{}) {
	s := append([]byte(fmt.Sprintf(str, args...)), 0)
	logCritical.Call(uintptr(category), uintptr(unsafe.Pointer(&s[0])))
}

// LogDebug logs a message with LOG_PRIORITY_DEBUG.
// (https://wiki.libsdl.org/SDL_LogDebug)
func LogDebug(category int, str string, args ...interface{}) {
	s := append([]byte(fmt.Sprintf(str, args...)), 0)
	logDebug.Call(uintptr(category), uintptr(unsafe.Pointer(&s[0])))
}

// LogError logs a message with LOG_PRIORITY_ERROR.
// (https://wiki.libsdl.org/SDL_LogError)
func LogError(category int, str string, args ...interface{}) {
	s := append([]byte(fmt.Sprintf(str, args...)), 0)
	logError.Call(uintptr(category), uintptr(unsafe.Pointer(&s[0])))
}

// LogInfo logs a message with LOG_PRIORITY_INFO.
// (https://wiki.libsdl.org/SDL_LogInfo)
func LogInfo(category int, str string, args ...interface{}) {
	s := append([]byte(fmt.Sprintf(str, args...)), 0)
	logInfo.Call(uintptr(category), uintptr(unsafe.Pointer(&s[0])))
}

// LogMessage logs a message with the specified category and priority.
// (https://wiki.libsdl.org/SDL_LogMessage)
func LogMessage(category int, pri LogPriority, str string, args ...interface{}) {
	s := append([]byte(fmt.Sprintf(str, args...)), 0)
	logMessage.Call(uintptr(category), uintptr(pri), uintptr(unsafe.Pointer(&s[0])))
}

// LogResetPriorities resets all priorities to default.
// (https://wiki.libsdl.org/SDL_LogResetPriorities)
func LogResetPriorities() {
	logResetPriorities.Call()
}

// LogSetAllPriority sets the priority of all log categories.
// (https://wiki.libsdl.org/SDL_LogSetAllPriority)
func LogSetAllPriority(p LogPriority) {
	logSetAllPriority.Call(uintptr(p))
}

// LogSetOutputFunction replaces the default log output function with one of your own.
// (https://wiki.libsdl.org/SDL_LogSetOutputFunction)
func LogSetOutputFunction(f LogOutputFunction, data interface{}) {
	// TODO
	//ctx := &logOutputFunctionCtx{
	//	f: f,
	//	d: data,
	//}
	//C.LogSetOutputFunction(unsafe.Pointer(ctx))
	//logOutputFunctionCache = f
	//logOutputDataCache = data
}

// LogSetPriority sets the priority of a particular log category.
// (https://wiki.libsdl.org/SDL_LogSetPriority)
func LogSetPriority(category int, p LogPriority) {
	logSetPriority.Call(uintptr(category), uintptr(p))
}

// LogVerbose logs a message with LOG_PRIORITY_VERBOSE.
// (https://wiki.libsdl.org/SDL_LogVerbose)
func LogVerbose(category int, str string, args ...interface{}) {
	s := append([]byte(fmt.Sprintf(str, args...)), 0)
	logVerbose.Call(uintptr(category), uintptr(unsafe.Pointer(&s[0])))
}

// LogWarn logs a message with LOG_PRIORITY_WARN.
// (https://wiki.libsdl.org/SDL_LogWarn)
func LogWarn(category int, str string, args ...interface{}) {
	s := append([]byte(fmt.Sprintf(str, args...)), 0)
	logWarn.Call(uintptr(category), uintptr(unsafe.Pointer(&s[0])))
}

// Main entry point. Run this function at the beginning of main(), and pass your
// own main body to it as a function. E.g.:
//
// 	func main() {
// 		sdl.Main(func() {
// 			// Your code here....
// 			// [....]
//
// 			// Calls to SDL can be made by any goroutine, but always guarded by sdl.Do()
// 			sdl.Do(func() {
// 				sdl.Init(0)
// 			})
// 		})
// 	}
//
// Avoid calling functions like os.Exit(..) within your passed-in function since
// they don't respect deferred calls. Instead, do this:
//
// 	func main() {
// 		var exitcode int
// 		sdl.Main(func() {
// 			exitcode = run()) // assuming run has signature func() int
// 		})
// 		os.Exit(exitcode)
// 	}
func Main(main func()) {
	// TODO
}

// MapRGB maps an RGB triple to an opaque pixel value for a given pixel format.
// (https://wiki.libsdl.org/SDL_MapRGB)
func MapRGB(format *PixelFormat, r, g, b uint8) uint32 {
	ret, _, _ := mapRGB.Call(
		uintptr(unsafe.Pointer(format)),
		uintptr(r),
		uintptr(g),
		uintptr(b),
	)
	return uint32(ret)
}

// MapRGBA maps an RGBA quadruple to a pixel value for a given pixel format.
// (https://wiki.libsdl.org/SDL_MapRGBA)
func MapRGBA(format *PixelFormat, r, g, b, a uint8) uint32 {
	ret, _, _ := mapRGBA.Call(
		uintptr(unsafe.Pointer(format)),
		uintptr(r),
		uintptr(g),
		uintptr(b),
		uintptr(a),
	)
	return uint32(ret)
}

// MasksToPixelFormatEnum converts a bpp value and RGBA masks to an enumerated pixel format.
// (https://wiki.libsdl.org/SDL_MasksToPixelFormatEnum)
func MasksToPixelFormatEnum(bpp int, rmask, gmask, bmask, amask uint32) uint {
	ret, _, _ := masksToPixelFormatEnum.Call(
		uintptr(bpp),
		uintptr(rmask),
		uintptr(gmask),
		uintptr(bmask),
		uintptr(amask),
	)
	return uint(ret)
}

// MixAudio mixes audio data. New programs might want to use MixAudioFormat() instead.
// (https://wiki.libsdl.org/SDL_MixAudio)
func MixAudio(dst, src *uint8, len uint32, volume int) {
	mixAudio.Call(
		uintptr(unsafe.Pointer(dst)),
		uintptr(unsafe.Pointer(src)),
		uintptr(len),
		uintptr(volume),
	)
}

// MixAudioFormat mixes audio data in a specified format.
// (https://wiki.libsdl.org/SDL_MixAudioFormat)
func MixAudioFormat(dst, src *uint8, format AudioFormat, len uint32, volume int) {
	mixAudioFormat.Call(
		uintptr(unsafe.Pointer(dst)),
		uintptr(unsafe.Pointer(src)),
		uintptr(format),
		uintptr(len),
		uintptr(volume),
	)
}

// MouseIsHaptic reports whether or not the current mouse has haptic capabilities.
// (https://wiki.libsdl.org/SDL_MouseIsHaptic)
func MouseIsHaptic() (bool, error) {
	ret, _, _ := mouseIsHaptic.Call()
	return ret != 0, errorFromInt(int(ret))
}

// NumHaptics returns the number of haptic devices attached to the system.
// (https://wiki.libsdl.org/SDL_NumHaptics)
func NumHaptics() (int, error) {
	ret, _, _ := numHaptics.Call()
	return int(ret), errorFromInt(int(ret))
}

// NumJoysticks returns the number of joysticks attached to the system.
// (https://wiki.libsdl.org/SDL_NumJoysticks)
func NumJoysticks() int {
	ret, _, _ := numJoysticks.Call()
	return int(ret)
}

// NumSensors counts the number of sensors attached to the system right now
// (https://wiki.libsdl.org/SDL_NumSensors)
func NumSensors() int {
	ret, _, _ := numSensors.Call()
	return int(ret)
}

// OpenAudio opens the audio device. New programs might want to use OpenAudioDevice() instead.
// (https://wiki.libsdl.org/SDL_OpenAudio)
func OpenAudio(desired, obtained *AudioSpec) error {
	ret, _, _ := openAudio.Call(
		uintptr(unsafe.Pointer(desired)),
		uintptr(unsafe.Pointer(obtained)),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// OutOfMemory sets SDL error message to ENOMEM (out of memory).
func OutOfMemory() {
	Error(ENOMEM)
}

// PauseAudio pauses and unpauses the audio device. New programs might want to use SDL_PauseAudioDevice() instead.
// (https://wiki.libsdl.org/SDL_PauseAudio)
func PauseAudio(pauseOn bool) {
	pauseAudio.Call(uintptr(Btoi(pauseOn)))
}

// PauseAudioDevice pauses and unpauses audio playback on a specified device.
// (https://wiki.libsdl.org/SDL_PauseAudioDevice)
func PauseAudioDevice(dev AudioDeviceID, pauseOn bool) {
	pauseAudioDevice.Call(
		uintptr(dev),
		uintptr(Btoi(pauseOn)),
	)
}

// PeepEvents checks the event queue for messages and optionally returns them.
// (https://wiki.libsdl.org/SDL_PeepEvents)
func PeepEvents(events []Event, action EventAction, minType, maxType uint32) (storedEvents int, err error) {
	// TODO look at what the original version does and figure out why
	ret, _, _ := peepEvents.Call(
		uintptr(unsafe.Pointer(&events[0])),
		uintptr(len(events)),
		uintptr(action),
		uintptr(minType),
		uintptr(maxType),
	)
	storedEvents = int(ret)
	if ret > uintptr(len(events)) {
		err = GetError()
		storedEvents = -1
	}
	return
}

// PixelFormatEnumToMasks converts one of the enumerated pixel formats to a bpp value and RGBA masks.
// (https://wiki.libsdl.org/SDL_PixelFormatEnumToMasks)
func PixelFormatEnumToMasks(format uint) (bpp int, rmask, gmask, bmask, amask uint32, err error) {
	ret, _, _ := pixelFormatEnumToMasks.Call(
		uintptr(format),
		uintptr(unsafe.Pointer(&bpp)),
		uintptr(unsafe.Pointer(&rmask)),
		uintptr(unsafe.Pointer(&gmask)),
		uintptr(unsafe.Pointer(&bmask)),
		uintptr(unsafe.Pointer(&amask)),
	)
	if ret == 0 {
		err = GetError()
	}
	return
}

// PumpEvents pumps the event loop, gathering events from the input devices.
// (https://wiki.libsdl.org/SDL_PumpEvents)
func PumpEvents() {
	pumpEvents.Call()
}

// PushEvent adds an event to the event queue.
// (https://wiki.libsdl.org/SDL_PushEvent)
func PushEvent(event Event) (filtered bool, err error) {
	// TODO
	//_event := (*C.SDL_Event)(unsafe.Pointer(cEvent(event)))
	//if ok := int(C.SDL_PushEvent(_event)); ok < 0 {
	//	filtered, err = false, GetError()
	//} else if ok == 0 {
	//	filtered, err = true, nil
	//}
	return
}

// QueueAudio queues more audio on non-callback devices.
// (https://wiki.libsdl.org/SDL_QueueAudio)
func QueueAudio(dev AudioDeviceID, data []byte) error {
	ret, _, _ := queueAudio.Call(
		uintptr(dev),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// Quit cleans up all initialized subsystems. You should call it upon all exit conditions.
// (https://wiki.libsdl.org/SDL_Quit)
func Quit() {
	quit.Call()
	// TODO
	//eventFilterCache = nil
	//for k := range eventWatches {
	//	delete(eventWatches, k)
	//}
}

// QuitSubSystem shuts down specific SDL subsystems.
// (https://wiki.libsdl.org/SDL_QuitSubSystem)
func QuitSubSystem(flags uint32) {
	quitSubSystem.Call(uintptr(flags))
}

// RecordGesture begins recording a gesture on a specified touch device or all touch devices.
// (https://wiki.libsdl.org/SDL_RecordGesture)
func RecordGesture(t TouchID) int {
	ret, _, _ := recordGesture.Call(uintptr(t))
	return int(ret)
}

// RegisterEvents allocates a set of user-defined events, and return the beginning event number for that set of events.
// (https://wiki.libsdl.org/SDL_RegisterEvents)
func RegisterEvents(numEvents int) uint32 {
	ret, _, _ := registerEvents.Call(uintptr(numEvents))
	return uint32(ret)
}

// SaveAllDollarTemplates saves all currently loaded Dollar Gesture templates.
// (https://wiki.libsdl.org/SDL_SaveAllDollarTemplates)
func SaveAllDollarTemplates(src *RWops) int {
	ret, _, _ := saveAllDollarTemplates.Call(uintptr(unsafe.Pointer(src)))
	return int(ret)
}

// SaveDollarTemplate saves a currently loaded Dollar Gesture template.
// (https://wiki.libsdl.org/SDL_SaveDollarTemplate)
func SaveDollarTemplate(g GestureID, src *RWops) int {
	ret, _, _ := saveDollarTemplate.Call(uintptr(g), uintptr(unsafe.Pointer(src)))
	return int(ret)
}

// SensorGetDeviceName gets the implementation dependent name of a sensor.
//
// This can be called before any sensors are opened.
//
// Returns the sensor name, or empty string if deviceIndex is out of range.
// (https://wiki.libsdl.org/SDL_SensorGetDeviceName)
func SensorGetDeviceName(deviceIndex int) (name string) {
	ret, _, _ := sensorGetDeviceName.Call(uintptr(deviceIndex))
	return sdlToGoString(ret)
}

// SensorGetDeviceNonPortableType gets the platform dependent type of a sensor.
//
// This can be called before any sensors are opened.
//
// Returns the sensor platform dependent type, or -1 if deviceIndex is out of range.
// (https://wiki.libsdl.org/SDL_SensorGetDeviceNonPortableType)
func SensorGetDeviceNonPortableType(deviceIndex int) (typ int) {
	ret, _, _ := sensorGetDeviceNonPortableType.Call(uintptr(deviceIndex))
	return int(ret)
}

// SensorUpdate updates the current state of the open sensors.
//
// This is called automatically by the event loop if sensor events are enabled.
//
// This needs to be called from the thread that initialized the sensor subsystem.
// (https://wiki.libsdl.org/SDL_SensorUpdate)
func SensorUpdate() {
	sensorUpdate.Call()
}

// SetClipboardText puts UTF-8 text into the clipboard.
// (https://wiki.libsdl.org/SDL_SetClipboardText)
func SetClipboardText(text string) error {
	t := append([]byte(text), 0)
	ret, _, _ := setClipboardText.Call(uintptr(unsafe.Pointer(&t[0])))
	if ret != 0 {
		return GetError()
	}
	return nil
}

// SetCursor sets the active cursor.
// (https://wiki.libsdl.org/SDL_SetCursor)
func SetCursor(cursor *Cursor) {
	setCursor.Call(uintptr(unsafe.Pointer(cursor)))
}

// SetError set the SDL error message.
// (https://wiki.libsdl.org/SDL_SetError)
func SetError(err error) {
	var msg string
	if err != nil {
		msg = err.Error()
	}
	m := append([]byte(msg), 0)
	setError.Call(uintptr(unsafe.Pointer(&m[0])))
}

// SetEventFilter sets up a filter to process all events before they change internal state and are posted to the internal event queue.
// (https://wiki.libsdl.org/SDL_SetEventFilter)
func SetEventFilter(filter EventFilter, userdata interface{}) {
	if eventFilterCache == nil && filter == nil {
		// nothing to do...
		return
	}

	if eventFilterCache == nil && filter != nil {
		// We had no event filter before and do now; lets set
		// goSetEventFilterCallback() as the event filter.
		setEventFilter.Call(setEventFilterCallbackPtr, 0)
	} else if eventFilterCache != nil && filter == nil {
		// We had an event filter before, but no longer do, lets clear the
		// event filter
		setEventFilter.Call(0, 0)
	}

	eventFilterCache = filter
}

func theSetEventFilterCallback(data, event uintptr) uintptr {
	// No check for eventFilterCache != nil. Why? because it should never be
	// nil since the callback is set/unset based on the last filter being nil
	// /non-nil. If there is an issue, then it should panic here so we can
	// figure out why that is.

	return wrapEventFilterCallback(eventFilterCache, event, nil)
}

var setEventFilterCallbackPtr = syscall.NewCallback(theSetEventFilterCallback)

// SetEventFilterFunc sets up a function to process all events before they change internal state and are posted to the internal event queue.
// (https://wiki.libsdl.org/SDL_SetEventFilter)
func SetEventFilterFunc(filterFunc eventFilterFunc, userdata interface{}) {
	SetEventFilter(filterFunc, userdata)
}

// SetHint sets a hint with normal priority.
// (https://wiki.libsdl.org/SDL_SetHint)
func SetHint(name, value string) bool {
	n := append([]byte(name), 0)
	v := append([]byte(value), 0)
	ret, _, _ := setHint.Call(
		uintptr(unsafe.Pointer(&n[0])),
		uintptr(unsafe.Pointer(&v[0])),
	)
	return ret != 0
}

// SetHintWithPriority sets a hint with a specific priority.
// (https://wiki.libsdl.org/SDL_SetHintWithPriority)
func SetHintWithPriority(name, value string, hp HintPriority) bool {
	n := append([]byte(name), 0)
	v := append([]byte(value), 0)
	ret, _, _ := setHintWithPriority.Call(
		uintptr(unsafe.Pointer(&n[0])),
		uintptr(unsafe.Pointer(&v[0])),
		uintptr(hp),
	)
	return ret != 0
}

// SetModState sets the current key modifier state for the keyboard.
// (https://wiki.libsdl.org/SDL_SetModState)
func SetModState(mod Keymod) {
	setModState.Call(uintptr(mod))
}

// SetRelativeMouseMode sets relative mouse mode.
// (https://wiki.libsdl.org/SDL_SetRelativeMouseMode)
func SetRelativeMouseMode(enabled bool) int {
	ret, _, _ := setRelativeMouseMode.Call(uintptr(Btoi(enabled)))
	return int(ret)
}

// SetTextInputRect sets the rectangle used to type Unicode text inputs.
// (https://wiki.libsdl.org/SDL_SetTextInputRect)
func SetTextInputRect(rect *Rect) {
	setTextInputRect.Call(uintptr(unsafe.Pointer(rect)))
}

// SetYUVConversionMode sets the YUV conversion mode
// TODO: (https://wiki.libsdl.org/SDL_SetYUVConversionMode)
func SetYUVConversionMode(mode YUV_CONVERSION_MODE) {
	setYUVConversionMode.Call(uintptr(mode))
}

// ShowCursor toggles whether or not the cursor is shown.
// (https://wiki.libsdl.org/SDL_ShowCursor)
func ShowCursor(toggle int) (int, error) {
	ret, _, _ := showCursor.Call(uintptr(toggle))
	return int(ret), errorFromInt(int(ret))
}

// ShowMessageBox creates a modal message box.
// (https://wiki.libsdl.org/SDL_ShowMessageBox)
func ShowMessageBox(data *MessageBoxData) (buttonid int32, err error) {
	// TODO
	//_title := C.CString(data.Title)
	//defer C.free(unsafe.Pointer(_title))
	//_message := C.CString(data.Message)
	//defer C.free(unsafe.Pointer(_message))
	//
	//var cbuttons []C.SDL_MessageBoxButtonData
	//var cbtntexts []*C.char
	//defer func(texts []*C.char) {
	//	for _, t := range texts {
	//		C.free(unsafe.Pointer(t))
	//	}
	//}(cbtntexts)
	//
	//for _, btn := range data.Buttons {
	//	ctext := C.CString(btn.Text)
	//	cbtn := C.SDL_MessageBoxButtonData{
	//		flags:    C.Uint32(btn.Flags),
	//		buttonid: C.int(btn.ButtonID),
	//		text:     ctext,
	//	}
	//
	//	cbuttons = append(cbuttons, cbtn)
	//	cbtntexts = append(cbtntexts, ctext)
	//}
	//
	//cdata := C.SDL_MessageBoxData{
	//	flags:       C.Uint32(data.Flags),
	//	window:      data.Window.cptr(),
	//	title:       _title,
	//	message:     _message,
	//	numbuttons:  C.int(data.NumButtons),
	//	buttons:     &cbuttons[0],
	//	colorScheme: data.ColorScheme.cptr(),
	//}
	//
	//buttonid = int32(C.ShowMessageBox(cdata))
	//return buttonid, errorFromInt(int(buttonid))
	return
}

// ShowSimpleMessageBox displays a simple modal message box.
// (https://wiki.libsdl.org/SDL_ShowSimpleMessageBox)
func ShowSimpleMessageBox(flags uint32, title, message string, window *Window) error {
	t := append([]byte(title), 0)
	m := append([]byte(message), 0)
	ret, _, _ := showSimpleMessageBox.Call(
		uintptr(flags),
		uintptr(unsafe.Pointer(&t[0])),
		uintptr(unsafe.Pointer(&m[0])),
		uintptr(unsafe.Pointer(window)),
	)
	return errorFromInt(int(ret))
}

// StartTextInput starts accepting Unicode text input events.
// (https://wiki.libsdl.org/SDL_StartTextInput)
func StartTextInput() {
	startTextInput.Call()
}

// StopTextInput stops receiving any text input events.
// (https://wiki.libsdl.org/SDL_StopTextInput)
func StopTextInput() {
	stopTextInput.Call()
}

// UnlockAudio unlocks the audio device. New programs might want to use UnlockAudioDevice() instead.
// (https://wiki.libsdl.org/SDL_UnlockAudio)
func UnlockAudio() {
	unlockAudio.Call()
}

// UnlockAudioDevice unlocks the audio callback function for a specified device.
// (https://wiki.libsdl.org/SDL_UnlockAudioDevice)
func UnlockAudioDevice(dev AudioDeviceID) {
	unlockAudioDevice.Call(uintptr(dev))
}

// UnlockJoysticks unlocks joysticks for multi-threaded access to the joystick API
// TODO: (https://wiki.libsdl.org/SDL_UnlockJoysticks)
func UnlockJoysticks() {
	unlockJoysticks.Call()
}

// Unsupported sets SDL error message to UNSUPPORTED (that operation is not supported).
func Unsupported() {
	Error(UNSUPPORTED)
}

// VERSION fills the selected struct with the version of SDL in use.
// (https://wiki.libsdl.org/SDL_VERSION)
func VERSION(v *Version) {
	v.Major = MAJOR_VERSION
	v.Minor = MINOR_VERSION
	v.Patch = PATCHLEVEL
}

// VERSIONNUM converts separate version components into a single numeric value.
// (https://wiki.libsdl.org/SDL_VERSIONNUM)
func VERSIONNUM(x, y, z int) int {
	return (x*1000 + y*100 + z)
}

// VERSION_ATLEAST reports whether the SDL version compiled against is at least as new as the specified version.
// (https://wiki.libsdl.org/SDL_VERSION_ATLEAST)
func VERSION_ATLEAST(x, y, z int) bool {
	return COMPILEDVERSION() >= VERSIONNUM(x, y, z)
}

// VideoInit initializes the video subsystem, optionally specifying a video driver.
// (https://wiki.libsdl.org/SDL_VideoInit)
func VideoInit(driverName string) error {
	d := append([]byte(driverName), 0)
	ret, _, _ := videoInit.Call(uintptr(unsafe.Pointer(&d[0])))
	return errorFromInt(int(ret))
}

// VideoQuit shuts down the video subsystem, if initialized with VideoInit().
// (https://wiki.libsdl.org/SDL_VideoQuit)
func VideoQuit() {
	videoQuit.Call()
}

// VulkanGetVkGetInstanceProcAddr gets the address of the vkGetInstanceProcAddr function.
// (https://wiki.libsdl.org/SDL_Vulkan_GetVkInstanceProcAddr)
func VulkanGetVkGetInstanceProcAddr() unsafe.Pointer {
	ret, _, _ := vulkan_GetVkGetInstanceProcAddr.Call()
	return unsafe.Pointer(ret)
}

// VulkanLoadLibrary dynamically loads a Vulkan loader library.
// (https://wiki.libsdl.org/SDL_Vulkan_LoadLibrary)
func VulkanLoadLibrary(path string) error {
	var ret uintptr
	if path == "" {
		ret, _, _ = vulkan_LoadLibrary.Call(0)
	} else {
		p := append([]byte(path), 0)
		ret, _, _ = vulkan_LoadLibrary.Call(uintptr(unsafe.Pointer(&p[0])))
	}
	if ret != 0 {
		return GetError()
	}
	return nil
}

// VulkanUnloadLibrary unloads the Vulkan loader library previously loaded by VulkanLoadLibrary().
// (https://wiki.libsdl.org/SDL_Vulkan_UnloadLibrary)
func VulkanUnloadLibrary() {
	vulkan_UnloadLibrary.Call()
}

// WarpMouseGlobal moves the mouse to the given position in global screen space.
// (https://wiki.libsdl.org/SDL_WarpMouseGlobal)
func WarpMouseGlobal(x, y int32) error {
	ret, _, _ := warpMouseGlobal.Call(uintptr(x), uintptr(y))
	return errorFromInt(int(ret))
}

// WasInit returns a mask of the specified subsystems which have previously been initialized.
// (https://wiki.libsdl.org/SDL_WasInit)
func WasInit(flags uint32) uint32 {
	ret, _, _ := wasInit.Call(uintptr(flags))
	return uint32(ret)
}

// AudioCVT contains audio data conversion information.
// (https://wiki.libsdl.org/SDL_AudioCVT)
type AudioCVT struct {
	Needed      int32           // set to 1 if conversion possible
	SrcFormat   AudioFormat     // source audio format
	DstFormat   AudioFormat     // target audio format
	RateIncr    float64         // rate conversion increment
	Buf         unsafe.Pointer  // the buffer to hold entire audio data. Use AudioCVT.BufAsSlice() for access via a Go slice
	Len         int32           // length of original audio buffer
	LenCVT      int32           // length of converted audio buffer
	LenMult     int32           // buf must be len*len_mult big
	LenRatio    float64         // given len, final size is len*len_ratio
	filters     [10]AudioFilter // filter list (internal use)
	filterIndex int32           // current audio conversion function (internal use)
}

// AllocBuf allocates the requested memory for AudioCVT buffer.
func (cvt *AudioCVT) AllocBuf(size uintptr) {
	// TODO
	//cvt.Buf = C.malloc(C.size_t(size))
}

// BufAsSlice returns AudioCVT.buf as byte slice.
// NOTE: Must be used after ConvertAudio() because it uses LenCVT as slice length.
func (cvt AudioCVT) BufAsSlice() []byte {
	var b []byte
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sliceHeader.Len = int(cvt.LenCVT)
	sliceHeader.Cap = int(cvt.Len * cvt.LenMult)
	sliceHeader.Data = uintptr(unsafe.Pointer(cvt.Buf))
	return b
}

// FreeBuf deallocates the memory previously allocated from AudioCVT buffer.
func (cvt *AudioCVT) FreeBuf() {
	// TODO
	//C.free(cvt.Buf)
}

// AudioCallback is a function to call when the audio device needs more data.
// (https://wiki.libsdl.org/SDL_AudioSpec)
type AudioCallback uintptr // TODO this is a function pointer in C

// AudioDeviceEvent contains audio device event information.
// (https://wiki.libsdl.org/SDL_AudioDeviceEvent)
type AudioDeviceEvent struct {
	Type      uint32 // AUDIODEVICEADDED, AUDIODEVICEREMOVED
	Timestamp uint32 // the timestamp of the event
	Which     uint32 // the audio device index for the AUDIODEVICEADDED event (valid until next GetNumAudioDevices() call), AudioDeviceID for the AUDIODEVICEREMOVED event
	IsCapture uint8  // zero if an audio output device, non-zero if an audio capture device
	_         uint8  // padding
	_         uint8  // padding
	_         uint8  // padding
}

// GetTimestamp returns the timestamp of the event.
func (e *AudioDeviceEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *AudioDeviceEvent) GetType() uint32 {
	return e.Type
}

// AudioDeviceID is ID of an audio device previously opened with OpenAudioDevice().
// (https://wiki.libsdl.org/SDL_OpenAudioDevice)
type AudioDeviceID uint32

// OpenAudioDevice opens a specific audio device.
// (https://wiki.libsdl.org/SDL_OpenAudioDevice)
func OpenAudioDevice(device string, isCapture bool, desired, obtained *AudioSpec, allowedChanges int) (AudioDeviceID, error) {
	d := append([]byte(device), 0)
	var devicePtr uintptr
	if device != "" {
		devicePtr = uintptr(unsafe.Pointer(&d[0]))
	}
	ret, _, _ := openAudioDevice.Call(
		devicePtr,
		uintptr(Btoi(isCapture)),
		uintptr(unsafe.Pointer(&desired)),
		uintptr(unsafe.Pointer(&obtained)),
		uintptr(allowedChanges),
	)
	if ret == 0 {
		return 0, GetError()
	}
	return AudioDeviceID(ret), nil
}

// AudioFilter is the filter list used in AudioCVT() (internal use)
// (https://wiki.libsdl.org/SDL_AudioCVT)
type AudioFilter uintptr // TODO this is a function pointer in C

// AudioFormat is an enumeration of audio formats.
// (https://wiki.libsdl.org/SDL_AudioFormat)
type AudioFormat uint16

// BitSize returns audio formats bit size.
// (https://wiki.libsdl.org/SDL_AudioFormat)
func (fmt AudioFormat) BitSize() uint8 {
	return uint8(fmt & AUDIO_MASK_BITSIZE)
}

// IsBigEndian reports whether audio format is big-endian.
// (https://wiki.libsdl.org/SDL_AudioFormat)
func (fmt AudioFormat) IsBigEndian() bool {
	return (fmt & AUDIO_MASK_ENDIAN) > 0
}

// IsFloat reports whether audio format is float.
// (https://wiki.libsdl.org/SDL_AudioFormat)
func (fmt AudioFormat) IsFloat() bool {
	return (fmt & AUDIO_MASK_DATATYPE) > 0
}

// IsInt reports whether audio format is integer.
// (https://wiki.libsdl.org/SDL_AudioFormat)
func (fmt AudioFormat) IsInt() bool {
	return !fmt.IsFloat()
}

// IsLittleEndian reports whether audio format is little-endian.
// (https://wiki.libsdl.org/SDL_AudioFormat)
func (fmt AudioFormat) IsLittleEndian() bool {
	return !fmt.IsBigEndian()
}

// IsSigned reports whether audio format is signed.
// (https://wiki.libsdl.org/SDL_AudioFormat)
func (fmt AudioFormat) IsSigned() bool {
	return (fmt & AUDIO_MASK_SIGNED) > 0
}

// IsUnsigned reports whether audio format is unsigned.
// (https://wiki.libsdl.org/SDL_AudioFormat)
func (fmt AudioFormat) IsUnsigned() bool {
	return !fmt.IsSigned()
}

// AudioSpec contains the audio output format. It also contains a callback that is called when the audio device needs more data.
// (https://wiki.libsdl.org/SDL_AudioSpec)
type AudioSpec struct {
	Freq     int32          // DSP frequency (samples per second)
	Format   AudioFormat    // audio data format
	Channels uint8          // number of separate sound channels
	Silence  uint8          // audio buffer silence value (calculated)
	Samples  uint16         // audio buffer size in samples (power of 2)
	_        uint16         // padding
	Size     uint32         // audio buffer size in bytes (calculated)
	Callback AudioCallback  // the function to call when the audio device needs more data
	UserData unsafe.Pointer // a pointer that is passed to callback (otherwise ignored by SDL)
}

// LoadWAV loads a WAVE from a file.
// (https://wiki.libsdl.org/SDL_LoadWAV)
func LoadWAV(file string) ([]byte, *AudioSpec) {
	// TODO
	return nil, nil
	//_file := C.CString(file)
	//_rb := C.CString("rb")
	//defer C.free(unsafe.Pointer(_file))
	//defer C.free(unsafe.Pointer(_rb))
	//
	//var _audioBuf *C.Uint8
	//var _audioLen C.Uint32
	//audioSpec := (*AudioSpec)(unsafe.Pointer(C.SDL_LoadWAV_RW(C.SDL_RWFromFile(_file, _rb), 1, (&AudioSpec{}).cptr(), &_audioBuf, &_audioLen)))
	//
	//var b []byte
	//sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	//sliceHeader.Len = (int)(_audioLen)
	//sliceHeader.Cap = (int)(_audioLen)
	//sliceHeader.Data = uintptr(unsafe.Pointer(_audioBuf))
	//return b, audioSpec
}

// LoadWAVRW loads a WAVE from the data source, automatically freeing that source if freeSrc is true.
// (https://wiki.libsdl.org/SDL_LoadWAV_RW)
func LoadWAVRW(src *RWops, freeSrc bool) ([]byte, *AudioSpec) {
	// TODO
	return nil, nil
	//var _audioBuf *C.Uint8
	//var _audioLen C.Uint32
	//audioSpec := (*AudioSpec)(unsafe.Pointer(C.SDL_LoadWAV_RW(src.cptr(), C.int(Btoi(freeSrc)), (&AudioSpec{}).cptr(), &_audioBuf, &_audioLen)))
	//
	//var b []byte
	//sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	//sliceHeader.Len = (int)(_audioLen)
	//sliceHeader.Cap = (int)(_audioLen)
	//sliceHeader.Data = uintptr(unsafe.Pointer(_audioBuf))
	//return b, audioSpec
}

// AudioStatus is an enumeration of audio device states.
// (https://wiki.libsdl.org/SDL_AudioStatus)
type AudioStatus uint32

// GetAudioDeviceStatus returns the current audio state of an audio device.
// (https://wiki.libsdl.org/SDL_GetAudioDeviceStatus)
func GetAudioDeviceStatus(dev AudioDeviceID) AudioStatus {
	ret, _, _ := getAudioDeviceStatus.Call(uintptr(dev))
	return AudioStatus(ret)
}

// GetAudioStatus returns the current audio state of the audio device. New programs might want to use GetAudioDeviceStatus() instead.
// (https://wiki.libsdl.org/SDL_GetAudioStatus)
func GetAudioStatus() AudioStatus {
	ret, _, _ := getAudioStatus.Call()
	return AudioStatus(ret)
}

// AudioStream is a new audio conversion interface.
// (https://wiki.libsdl.org/SDL_AudioStream)
type AudioStream uintptr

// NewAudioStream creates a new audio stream
// TODO: (https://wiki.libsdl.org/SDL_NewAudioStream)
func NewAudioStream(srcFormat AudioFormat, srcChannels uint8, srcRate int, dstFormat AudioFormat, dstChannels uint8, dstRate int) (stream *AudioStream, err error) {
	ret, _, _ := newAudioStream.Call(
		uintptr(srcFormat),
		uintptr(srcChannels),
		uintptr(srcRate),
		uintptr(dstFormat),
		uintptr(dstChannels),
		uintptr(dstRate),
	)
	if ret == 0 {
		return nil, GetError()
	}
	return (*AudioStream)(unsafe.Pointer(ret)), nil
}

// Available gets the number of converted/resampled bytes available
// TODO: (https://wiki.libsdl.org/SDL_AudioStreamAvailable)
func (stream *AudioStream) Available() (err error) {
	ret, _, _ := audioStreamAvailable.Call(uintptr(unsafe.Pointer(stream)))
	return errorFromInt(int(ret))
}

// Clear clears any pending data in the stream without converting it
// TODO: (https://wiki.libsdl.org/SDL_AudioStreamClear)
func (stream *AudioStream) Clear() {
	audioStreamClear.Call(uintptr(unsafe.Pointer(stream)))
}

// Flush tells the stream that you're done sending data, and anything being buffered
// should be converted/resampled and made available immediately.
// TODO: (https://wiki.libsdl.org/SDL_AudioStreamFlush)
func (stream *AudioStream) Flush() (err error) {
	ret, _, _ := audioStreamFlush.Call(uintptr(unsafe.Pointer(stream)))
	return errorFromInt(int(ret))
}

// Free frees the audio stream
// TODO: (https://wiki.libsdl.org/SDL_AudoiStreamFree)
func (stream *AudioStream) Free() {
	freeAudioStream.Call(uintptr(unsafe.Pointer(stream)))
}

// Get gets converted/resampled data from the stream
// TODO: (https://wiki.libsdl.org/SDL_AudioStreamGet)
func (stream *AudioStream) Get(buf []byte) (err error) {
	if len(buf) == 0 {
		return nil
	}
	ret, _, _ := audioStreamGet.Call(
		uintptr(unsafe.Pointer(stream)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	return errorFromInt(int(ret))
}

// Put adds data to be converted/resampled to the stream
// TODO: (https://wiki.libsdl.org/SDL_AudioStreamPut)
func (stream *AudioStream) Put(buf []byte) (err error) {
	if len(buf) == 0 {
		return nil
	}
	ret, _, _ := audioStreamPut.Call(
		uintptr(unsafe.Pointer(stream)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	return errorFromInt(int(ret))
}

// BlendFactor is an enumeration of blend factors used when creating a custom blend mode with ComposeCustomBlendMode().
// (https://wiki.libsdl.org/SDL_BlendFactor)
type BlendFactor uint32

// BlendMode is an enumeration of blend modes used in Render.Copy() and drawing operations.
// (https://wiki.libsdl.org/SDL_BlendMode)
type BlendMode uint32

// ComposeCustomBlendMode creates a custom blend mode, which may or may not be supported by a given renderer
// The result of the blend mode operation will be:
//     dstRGB = dstRGB * dstColorFactor colorOperation srcRGB * srcColorFactor
// and
//     dstA = dstA * dstAlphaFactor alphaOperation srcA * srcAlphaFactor
// (https://wiki.libsdl.org/SDL_ComposeCustomBlendMode)
func ComposeCustomBlendMode(srcColorFactor, dstColorFactor BlendFactor, colorOperation BlendOperation, srcAlphaFactor, dstAlphaFactor BlendFactor, alphaOperation BlendOperation) BlendMode {
	ret, _, _ := composeCustomBlendMode.Call(
		uintptr(srcColorFactor),
		uintptr(dstColorFactor),
		uintptr(colorOperation),
		uintptr(srcAlphaFactor),
		uintptr(dstAlphaFactor),
		uintptr(alphaOperation),
	)
	return BlendMode(ret)
}

// BlendOperation is an enumeration of blend operations used when creating a custom blend mode with ComposeCustomBlendMode().
// (https://wiki.libsdl.org/SDL_BlendOperation)
type BlendOperation uint32

// CEvent is a union of all event structures used in SDL.
// (https://wiki.libsdl.org/SDL_Event)
type CEvent struct {
	Type uint32
	_    [52]byte // padding
}

// ClipboardEvent contains clipboard event information.
// (https://wiki.libsdl.org/SDL_EventType)
type ClipboardEvent struct {
	Type      uint32 // CLIPBOARDUPDATE
	Timestamp uint32 // timestamp of the event
}

// GetTimestamp returns the timestamp of the event.
func (e *ClipboardEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *ClipboardEvent) GetType() uint32 {
	return e.Type
}

// CocoaInfo contains Apple Mac OS X window information.
type CocoaInfo struct {
	Window unsafe.Pointer // the Cocoa window
}

// Color represents a color. This implements image/color.Color interface.
// (https://wiki.libsdl.org/SDL_Color)
type Color color.RGBA

// Uint32 return uint32 representation of RGBA color.
func (c Color) Uint32() uint32 {
	var v uint32
	v |= uint32(c.A) << 24
	v |= uint32(c.R) << 16
	v |= uint32(c.G) << 8
	v |= uint32(c.B)
	return v
}

// CommonEvent contains common event data.
// (https://wiki.libsdl.org/SDL_Event)
type CommonEvent struct {
	Type      uint32 // the event type
	Timestamp uint32 // timestamp of the event
}

// GetTimestamp returns the timestamp of the event.
func (e *CommonEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *CommonEvent) GetType() uint32 {
	return e.Type
}

// Cond is the SDL condition variable structure.
type Cond struct {
	Lock     *Mutex
	Waiting  int
	Signals  int
	WaitSem  *Sem
	WaitDone *Sem
}

// CreateCond (https://wiki.libsdl.org/SDL_CreateCond)
func CreateCond() *Cond {
	ret, _, _ := createCond.Call()
	return (*Cond)(unsafe.Pointer(ret))
}

// Broadcast restarts all threads that are waiting on the condition variable.
// (https://wiki.libsdl.org/SDL_CondBroadcast)
func (cond *Cond) Broadcast() error {
	ret, _, _ := condBroadcast.Call(uintptr(unsafe.Pointer(cond)))
	if ret != 0 {
		return GetError()
	}
	return nil
}

// Destroy creates a condition variable.
// (https://wiki.libsdl.org/SDL_DestroyCond)
func (cond *Cond) Destroy() {
	destroyCond.Call(uintptr(unsafe.Pointer(cond)))
}

// Signal restarts one of the threads that are waiting on the condition variable.
// (https://wiki.libsdl.org/SDL_CondSignal)
func (cond *Cond) Signal() error {
	ret, _, _ := condSignal.Call(uintptr(unsafe.Pointer(cond)))
	if ret != 0 {
		return GetError()
	}
	return nil
}

// Wait waits until a condition variable is signaled.
// (https://wiki.libsdl.org/SDL_CondWait)
func (cond *Cond) Wait(mutex *Mutex) error {
	ret, _, _ := condWait.Call(
		uintptr(unsafe.Pointer(cond)),
		uintptr(unsafe.Pointer(mutex)),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// WaitTimeout waits until a condition variable is signaled or a specified amount of time has passed.
// (https://wiki.libsdl.org/SDL_CondWaitTimeout)
func (cond *Cond) WaitTimeout(mutex *Mutex, ms uint32) error {
	ret, _, _ := condWaitTimeout.Call(
		uintptr(unsafe.Pointer(cond)),
		uintptr(unsafe.Pointer(mutex)),
		uintptr(ms),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// ControllerAxisEvent contains game controller axis motion event information.
// (https://wiki.libsdl.org/SDL_ControllerAxisEvent)
type ControllerAxisEvent struct {
	Type      uint32     // CONTROLLERAXISMOTION
	Timestamp uint32     // the timestamp of the event
	Which     JoystickID // the joystick instance id
	Axis      uint8      // the controller axis (https://wiki.libsdl.org/SDL_GameControllerAxis)
	_         uint8      // padding
	_         uint8      // padding
	_         uint8      // padding
	Value     int16      // the axis value (range: -32768 to 32767)
	_         uint16     // padding
}

// GetTimestamp returns the timestamp of the event.
func (e *ControllerAxisEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *ControllerAxisEvent) GetType() uint32 {
	return e.Type
}

// ControllerButtonEvent contains game controller button event information.
// (https://wiki.libsdl.org/SDL_ControllerButtonEvent)
type ControllerButtonEvent struct {
	Type      uint32     // CONTROLLERBUTTONDOWN, CONTROLLERBUTTONUP
	Timestamp uint32     // the timestamp of the event
	Which     JoystickID // the joystick instance id
	Button    uint8      // the controller button (https://wiki.libsdl.org/SDL_GameControllerButton)
	State     uint8      // PRESSED, RELEASED
	_         uint8      // padding
	_         uint8      // padding
}

// GetTimestamp returns the timestamp of the event.
func (e *ControllerButtonEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *ControllerButtonEvent) GetType() uint32 {
	return e.Type
}

// ControllerDeviceEvent contains controller device event information.
// (https://wiki.libsdl.org/SDL_ControllerDeviceEvent)
type ControllerDeviceEvent struct {
	Type      uint32     // CONTROLLERDEVICEADDED, CONTROLLERDEVICEREMOVED, SDL_CONTROLLERDEVICEREMAPPED
	Timestamp uint32     // the timestamp of the event
	Which     JoystickID // the joystick device index for the CONTROLLERDEVICEADDED event or instance id for the CONTROLLERDEVICEREMOVED or CONTROLLERDEVICEREMAPPED event
}

// GetTimestamp returns the timestamp of the event.
func (e *ControllerDeviceEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *ControllerDeviceEvent) GetType() uint32 {
	return e.Type
}

// Cursor is a custom cursor created by CreateCursor() or CreateColorCursor().
type Cursor struct{}

// CreateColorCursor creates a color cursor.
// (https://wiki.libsdl.org/SDL_CreateColorCursor)
func CreateColorCursor(surface *Surface, hotX, hotY int32) *Cursor {
	ret, _, _ := createColorCursor.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(hotX),
		uintptr(hotY),
	)
	return (*Cursor)(unsafe.Pointer(ret))
}

// CreateCursor creates a cursor using the specified bitmap data and mask (in MSB format).
// (https://wiki.libsdl.org/SDL_CreateCursor)
func CreateCursor(data, mask *uint8, w, h, hotX, hotY int32) *Cursor {
	ret, _, _ := createCursor.Call(
		uintptr(unsafe.Pointer(data)),
		uintptr(unsafe.Pointer(mask)),
		uintptr(w),
		uintptr(h),
		uintptr(hotX),
		uintptr(hotY),
	)
	return (*Cursor)(unsafe.Pointer(ret))
}

// CreateSystemCursor creates a system cursor.
// (https://wiki.libsdl.org/SDL_CreateSystemCursor)
func CreateSystemCursor(id SystemCursor) *Cursor {
	ret, _, _ := createSystemCursor.Call(uintptr(id))
	return (*Cursor)(unsafe.Pointer(ret))
}

// GetCursor returns the active cursor.
// (https://wiki.libsdl.org/SDL_GetCursor)
func GetCursor() *Cursor {
	ret, _, _ := getCursor.Call()
	return (*Cursor)(unsafe.Pointer(ret))
}

// GetDefaultCursor returns the default cursor.
// (https://wiki.libsdl.org/SDL_GetDefaultCursor)
func GetDefaultCursor() *Cursor {
	ret, _, _ := getDefaultCursor.Call()
	return (*Cursor)(unsafe.Pointer(ret))
}

// DFBInfo contains DirectFB window information.
type DFBInfo struct {
	Dfb     unsafe.Pointer // the DirectFB main interface
	Window  unsafe.Pointer // the DirectFB window handle
	Surface unsafe.Pointer // the DirectFB client surface
}

// DisplayMode contains the description of a display mode.
// (https://wiki.libsdl.org/SDL_DisplayMode)
type DisplayMode struct {
	Format      uint32         // one of the PixelFormatEnum values (https://wiki.libsdl.org/SDL_PixelFormatEnum)
	W           int32          // width, in screen coordinates
	H           int32          // height, in screen coordinates
	RefreshRate int32          // refresh rate (in Hz), or 0 for unspecified
	DriverData  unsafe.Pointer // driver-specific data, initialize to 0
}

// GetClosestDisplayMode returns the closest match to the requested display mode.
// (https://wiki.libsdl.org/SDL_GetClosestDisplayMode)
func GetClosestDisplayMode(displayIndex int, mode *DisplayMode, closest *DisplayMode) (*DisplayMode, error) {
	ret, _, _ := getClosestDisplayMode.Call(
		uintptr(displayIndex),
		uintptr(unsafe.Pointer(mode)),
		uintptr(unsafe.Pointer(closest)),
	)
	if ret == 0 {
		return nil, GetError()
	}
	return (*DisplayMode)(unsafe.Pointer(ret)), nil
}

// GetCurrentDisplayMode returns information about the current display mode.
// (https://wiki.libsdl.org/SDL_GetCurrentDisplayMode)
func GetCurrentDisplayMode(displayIndex int) (mode DisplayMode, err error) {
	ret, _, _ := getCurrentDisplayMode.Call(
		uintptr(displayIndex),
		uintptr(unsafe.Pointer(&mode)),
	)
	err = errorFromInt(int(ret))
	return
}

// GetDesktopDisplayMode returns information about the desktop display mode.
// (https://wiki.libsdl.org/SDL_GetDesktopDisplayMode)
func GetDesktopDisplayMode(displayIndex int) (mode DisplayMode, err error) {
	ret, _, _ := getDesktopDisplayMode.Call(
		uintptr(displayIndex),
		uintptr(unsafe.Pointer(&mode)),
	)
	err = errorFromInt(int(ret))
	return
}

// GetDisplayMode retruns information about a specific display mode.
// (https://wiki.libsdl.org/SDL_GetDisplayMode)
func GetDisplayMode(displayIndex int, modeIndex int) (mode DisplayMode, err error) {
	ret, _, _ := getDisplayMode.Call(
		uintptr(displayIndex),
		uintptr(modeIndex),
		uintptr(unsafe.Pointer(&mode)),
	)
	err = errorFromInt(int(ret))
	return
}

// DollarGestureEvent contains complex gesture event information.
// (https://wiki.libsdl.org/SDL_DollarGestureEvent)
type DollarGestureEvent struct {
	Type       uint32    // DOLLARGESTURE, DOLLARRECORD
	Timestamp  uint32    // timestamp of the event
	TouchID    TouchID   // the touch device id
	GestureID  GestureID // the unique id of the closest gesture to the performed stroke
	NumFingers uint32    // the number of fingers used to draw the stroke
	Error      float32   // the difference between the gesture template and the actual performed gesture (lower error is a better match)
	X          float32   // the normalized center of gesture
	Y          float32   // the normalized center of gesture
}

// GetTimestamp returns the timestamp of the event.
func (e *DollarGestureEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *DollarGestureEvent) GetType() uint32 {
	return e.Type
}

// DropEvent contains an event used to request a file open by the system.
// (https://wiki.libsdl.org/SDL_DropEvent)
type DropEvent struct {
	Type      uint32 // DROPFILE, DROPTEXT, DROPBEGIN, DROPCOMPLETE
	Timestamp uint32 // timestamp of the event
	File      string // the file name
	WindowID  uint32 // the window that was dropped on, if any
}

// GetTimestamp returns the timestamp of the event.
func (e *DropEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *DropEvent) GetType() uint32 {
	return e.Type
}

// ErrorCode is an error code used in SDL error messages.
type ErrorCode uint32

// Event is a union of all event structures used in SDL.
// (https://wiki.libsdl.org/SDL_Event)
type Event interface {
	GetType() uint32      // GetType returns the event type
	GetTimestamp() uint32 // GetTimestamp returns the timestamp of the event
}

// PollEvent polls for currently pending events.
// (https://wiki.libsdl.org/SDL_PollEvent)
func PollEvent() Event {
	var e CEvent
	ret, _, _ := pollEvent.Call(uintptr(unsafe.Pointer(&e)))
	if ret == 0 {
		return nil
	}
	return goEvent(&e)
}

// WaitEvent waits indefinitely for the next available event.
// (https://wiki.libsdl.org/SDL_WaitEvent)
func WaitEvent() Event {
	var e CEvent
	ret, _, _ := waitEvent.Call(uintptr(unsafe.Pointer(&e)))
	if ret == 0 {
		return nil
	}
	return goEvent(&e)
}

// WaitEventTimeout waits until the specified timeout (in milliseconds) for the
// next available event.
// (https://wiki.libsdl.org/SDL_WaitEventTimeout)
func WaitEventTimeout(timeout int) Event {
	var e CEvent
	ret, _, _ := waitEventTimeout.Call(
		uintptr(unsafe.Pointer(&e)),
		uintptr(timeout),
	)
	if ret == 0 {
		return nil
	}
	return goEvent(&e)
}

func goEvent(cevent *CEvent) Event {
	switch cevent.Type {
	case WINDOWEVENT:
		return (*WindowEvent)(unsafe.Pointer(cevent))
	case SYSWMEVENT:
		return (*SysWMEvent)(unsafe.Pointer(cevent))
	case KEYDOWN, KEYUP:
		return (*KeyboardEvent)(unsafe.Pointer(cevent))
	case TEXTEDITING:
		return (*TextEditingEvent)(unsafe.Pointer(cevent))
	case TEXTINPUT:
		return (*TextInputEvent)(unsafe.Pointer(cevent))
	case MOUSEMOTION:
		return (*MouseMotionEvent)(unsafe.Pointer(cevent))
	case MOUSEBUTTONDOWN, MOUSEBUTTONUP:
		return (*MouseButtonEvent)(unsafe.Pointer(cevent))
	case MOUSEWHEEL:
		return (*MouseWheelEvent)(unsafe.Pointer(cevent))
	case JOYAXISMOTION:
		return (*JoyAxisEvent)(unsafe.Pointer(cevent))
	case JOYBALLMOTION:
		return (*JoyBallEvent)(unsafe.Pointer(cevent))
	case JOYHATMOTION:
		return (*JoyHatEvent)(unsafe.Pointer(cevent))
	case JOYBUTTONDOWN, JOYBUTTONUP:
		return (*JoyButtonEvent)(unsafe.Pointer(cevent))
	case JOYDEVICEADDED:
		return (*JoyDeviceAddedEvent)(unsafe.Pointer(cevent))
	case JOYDEVICEREMOVED:
		return (*JoyDeviceRemovedEvent)(unsafe.Pointer(cevent))
	case CONTROLLERAXISMOTION:
		return (*ControllerAxisEvent)(unsafe.Pointer(cevent))
	case CONTROLLERBUTTONDOWN, CONTROLLERBUTTONUP:
		return (*ControllerButtonEvent)(unsafe.Pointer(cevent))
	case CONTROLLERDEVICEADDED, CONTROLLERDEVICEREMOVED, CONTROLLERDEVICEREMAPPED:
		return (*ControllerDeviceEvent)(unsafe.Pointer(cevent))
	case AUDIODEVICEADDED, AUDIODEVICEREMOVED:
		return (*AudioDeviceEvent)(unsafe.Pointer(cevent))
	case FINGERMOTION, FINGERDOWN, FINGERUP:
		return (*TouchFingerEvent)(unsafe.Pointer(cevent))
	case MULTIGESTURE:
		return (*MultiGestureEvent)(unsafe.Pointer(cevent))
	case DOLLARGESTURE, DOLLARRECORD:
		return (*DollarGestureEvent)(unsafe.Pointer(cevent))
	case DROPFILE, DROPTEXT, DROPBEGIN, DROPCOMPLETE:
		e := (*tDropEvent)(unsafe.Pointer(cevent))
		event := DropEvent{
			Type:      e.Type,
			Timestamp: e.Timestamp,
			File:      sdlToGoString(uintptr(e.File)),
			WindowID:  e.WindowID,
		}
		return &event
	case SENSORUPDATE:
		return (*SensorEvent)(unsafe.Pointer(cevent))
	case RENDER_TARGETS_RESET, RENDER_DEVICE_RESET:
		return (*RenderEvent)(unsafe.Pointer(cevent))
	case QUIT:
		return (*QuitEvent)(unsafe.Pointer(cevent))
	case USEREVENT:
		return (*UserEvent)(unsafe.Pointer(cevent))
	case CLIPBOARDUPDATE:
		return (*ClipboardEvent)(unsafe.Pointer(cevent))
	default:
		return (*CommonEvent)(unsafe.Pointer(cevent))
	}
}

type tDropEvent struct {
	Type      uint32
	Timestamp uint32
	File      unsafe.Pointer
	WindowID  uint32
}

// EventAction is the action to take in PeepEvents() function.
// (https://wiki.libsdl.org/SDL_PeepEvents)
type EventAction uint32

// EventFilter is the function to call when an event happens.
// (https://wiki.libsdl.org/SDL_SetEventFilter)
type EventFilter interface {
	FilterEvent(e Event, userdata interface{}) bool
}

// GetEventFilter queries the current event filter.
// (https://wiki.libsdl.org/SDL_GetEventFilter)
func GetEventFilter() EventFilter {
	return eventFilterCache
}

// EventWatchHandle is an event watch callback added with AddEventWatch().
type EventWatchHandle uintptr

// AddEventWatch adds a callback to be triggered when an event is added to the event queue.
// (https://wiki.libsdl.org/SDL_AddEventWatch)
func AddEventWatch(filter EventFilter, userdata interface{}) EventWatchHandle {
	context := newEventFilterCallbackContext(filter, userdata)
	addEventWatch.Call(
		eventFilterCallbackPtr,
		uintptr(context.handle),
	)
	//C.addEventWatch(context.cptr())
	return context.handle
}

func theEventFilterCallback(userdata, event uintptr) uintptr {
	// same sort of reasoning with goSetEventFilterCallback, userdata should
	// always be non-nil and represent a valid eventFilterCallbackContext. If
	// it doesn't a panic will let us know that there something wrong and the
	// problem can be fixed.

	context := eventWatches[EventWatchHandle(userdata)]
	return wrapEventFilterCallback(context.filter, event, context.userdata)
}

var eventFilterCallbackPtr = syscall.NewCallback(theEventFilterCallback)

func wrapEventFilterCallback(filter EventFilter, e uintptr, userdata interface{}) uintptr {
	gev := goEvent((*CEvent)(unsafe.Pointer(e)))
	result := filter.FilterEvent(gev, userdata)
	if result {
		return 1
	}
	return 0
}

func newEventFilterCallbackContext(filter EventFilter, userdata interface{}) *eventFilterCallbackContext {
	lastEventWatchHandleMutex.Lock()
	defer lastEventWatchHandleMutex.Unlock()
	// Look for the next available watch handle (this should be immediate
	// unless you're creating a LOT of handlers).
	for {
		if _, ok := eventWatches[lastEventWatchHandle]; !ok {
			break
		}
		lastEventWatchHandle++
	}
	e := &eventFilterCallbackContext{filter, lastEventWatchHandle, userdata}
	eventWatches[lastEventWatchHandle] = e
	lastEventWatchHandle++
	return e
}

var (
	eventFilterCache          EventFilter
	eventWatches              = make(map[EventWatchHandle]*eventFilterCallbackContext)
	lastEventWatchHandleMutex sync.Mutex
	lastEventWatchHandle      EventWatchHandle
)

type eventFilterCallbackContext struct {
	filter   EventFilter
	handle   EventWatchHandle
	userdata interface{}
}

// AddEventWatchFunc adds a callback function to be triggered when an event is added to the event queue.
// (https://wiki.libsdl.org/SDL_AddEventWatch)
func AddEventWatchFunc(filterFunc eventFilterFunc, userdata interface{}) EventWatchHandle {
	// TODO
	return 0
	//return AddEventWatch(filterFunc, userdata)
}

// Finger contains touch information.
type Finger struct {
	ID       FingerID // the finger id
	X        float32  // the x-axis location of the touch event, normalized (0...1)
	Y        float32  // the y-axis location of the touch event, normalized (0...1)
	Pressure float32  // the quantity of pressure applied, normalized (0...1)
}

// GetTouchFinger returns the finger object for specified touch device ID and finger index.
// (https://wiki.libsdl.org/SDL_GetTouchFinger)
func GetTouchFinger(t TouchID, index int) *Finger {
	ret, _, _ := getTouchFinger.Call(uintptr(t), uintptr(index))
	return (*Finger)(unsafe.Pointer(ret))
}

// FingerID is a finger id.
type FingerID int64

// GLContext is an opaque handle to an OpenGL context.
type GLContext uintptr

// GLattr is an OpenGL configuration attribute.
//(https://wiki.libsdl.org/SDL_GLattr)
type GLattr uint32

// GameController used to identify an SDL game controller.
type GameController struct{}

// GameControllerFromInstanceID returns the GameController associated with an instance id.
// (https://wiki.libsdl.org/SDL_GameControllerFromInstanceID)
func GameControllerFromInstanceID(joyid JoystickID) *GameController {
	ret, _, _ := gameControllerFromInstanceID.Call(uintptr(joyid))
	return (*GameController)(unsafe.Pointer(ret))
}

// GameControllerOpen opens a gamecontroller for use.
// (https://wiki.libsdl.org/SDL_GameControllerOpen)
func GameControllerOpen(index int) *GameController {
	ret, _, _ := gameControllerOpen.Call(uintptr(index))
	return (*GameController)(unsafe.Pointer(ret))
}

// Attached reports whether a controller has been opened and is currently connected.
// (https://wiki.libsdl.org/SDL_GameControllerGetAttached)
func (ctrl *GameController) Attached() bool {
	ret, _, _ := gameControllerGetAttached.Call(uintptr(unsafe.Pointer(ctrl)))
	return ret == 1
}

// Axis returns the current state of an axis control on a game controller.
// (https://wiki.libsdl.org/SDL_GameControllerGetAxis)
func (ctrl *GameController) Axis(axis GameControllerAxis) int16 {
	ret, _, _ := gameControllerGetAxis.Call(
		uintptr(unsafe.Pointer(ctrl)),
		uintptr(axis),
	)
	return int16(ret)
}

// BindForAxis returns the SDL joystick layer binding for a controller button mapping.
// (https://wiki.libsdl.org/SDL_GameControllerGetBindForAxis)
func (ctrl *GameController) BindForAxis(axis GameControllerAxis) GameControllerButtonBind {
	// TODO why does this not return a pointer?
	ret, _, _ := gameControllerGetBindForAxis.Call(
		uintptr(unsafe.Pointer(ctrl)),
		uintptr(axis),
	)
	return *((*GameControllerButtonBind)(unsafe.Pointer(ret)))
}

// BindForButton returns the SDL joystick layer binding for this controller button mapping.
// (https://wiki.libsdl.org/SDL_GameControllerGetBindForButton)
func (ctrl *GameController) BindForButton(btn GameControllerButton) GameControllerButtonBind {
	// TODO why does this not return a pointer?
	ret, _, _ := gameControllerGetBindForButton.Call(
		uintptr(unsafe.Pointer(ctrl)),
		uintptr(btn),
	)
	return *((*GameControllerButtonBind)(unsafe.Pointer(ret)))
}

// Button returns the current state of a button on a game controller.
// (https://wiki.libsdl.org/SDL_GameControllerGetButton)
func (ctrl *GameController) Button(btn GameControllerButton) byte {
	ret, _, _ := gameControllerGetButton.Call(
		uintptr(unsafe.Pointer(ctrl)),
		uintptr(btn),
	)
	return byte(ret)
}

// Close closes a game controller previously opened with GameControllerOpen().
// (https://wiki.libsdl.org/SDL_GameControllerClose)
func (ctrl *GameController) Close() {
	gameControllerClose.Call(uintptr(unsafe.Pointer(ctrl)))
}

// Joystick returns the Joystick ID from a Game Controller. The game controller builds on the Joystick API, but to be able to use the Joystick's functions with a gamepad, you need to use this first to get the joystick object.
// (https://wiki.libsdl.org/SDL_GameControllerGetJoystick)
func (ctrl *GameController) Joystick() *Joystick {
	ret, _, _ := gameControllerGetJoystick.Call(uintptr(unsafe.Pointer(ctrl)))
	return (*Joystick)(unsafe.Pointer(ret))
}

// Mapping returns the current mapping of a Game Controller.
// (https://wiki.libsdl.org/SDL_GameControllerMapping)
func (ctrl *GameController) Mapping() string {
	ret, _, _ := gameControllerMapping.Call(uintptr(unsafe.Pointer(ctrl)))
	return sdlToGoString(ret)
}

// Name returns the implementation dependent name for an opened game controller.
// (https://wiki.libsdl.org/SDL_GameControllerName)
func (ctrl *GameController) Name() string {
	ret, _, _ := gameControllerName.Call(uintptr(unsafe.Pointer(ctrl)))
	return sdlToGoString(ret)
}

// Product returns the USB product ID of an opened controller, if available, 0 otherwise.
func (ctrl *GameController) Product() int {
	ret, _, _ := gameControllerGetProduct.Call(uintptr(unsafe.Pointer(ctrl)))
	return int(ret)
}

// ProductVersion returns the product version of an opened controller, if available, 0 otherwise.
func (ctrl *GameController) ProductVersion() int {
	ret, _, _ := gameControllerGetProductVersion.Call(uintptr(unsafe.Pointer(ctrl)))
	return int(ret)
}

// Vendor returns the USB vendor ID of an opened controller, if available, 0 otherwise.
func (ctrl *GameController) Vendor() int {
	ret, _, _ := gameControllerGetVendor.Call(uintptr(unsafe.Pointer(ctrl)))
	return int(ret)
}

// GameControllerAxis is an axis on a game controller.
// (https://wiki.libsdl.org/SDL_GameControllerAxis)
type GameControllerAxis uint32

// GameControllerGetAxisFromString converts a string into an enum representation for a GameControllerAxis.
// (https://wiki.libsdl.org/SDL_GameControllerGetAxisFromString)
func GameControllerGetAxisFromString(pchString string) GameControllerAxis {
	s := append([]byte(pchString), 0)
	ret, _, _ := gameControllerGetAxisFromString.Call(uintptr(unsafe.Pointer(&s[0])))
	return GameControllerAxis(ret)
}

// GameControllerBindType is a type of game controller input.
type GameControllerBindType uint32

// GameControllerButton is a button on a game controller.
// (https://wiki.libsdl.org/SDL_GameControllerButton)
type GameControllerButton uint32

// GameControllerGetButtonFromString turns a string into a button mapping.
// (https://wiki.libsdl.org/SDL_GameControllerGetButtonFromString)
func GameControllerGetButtonFromString(pchString string) GameControllerButton {
	s := append([]byte(pchString), 0)
	ret, _, _ := gameControllerGetButtonFromString.Call(uintptr(unsafe.Pointer(&s[0])))
	return GameControllerButton(ret)
}

// GameControllerButtonBind SDL joystick layer binding for controller button/axis mapping.
type GameControllerButtonBind struct {
	bindType GameControllerBindType
	value    [8]byte // TODO this is a union of two ints in the SDL code
}

// Axis returns axis mapped for this SDL joystick layer binding.
func (bind *GameControllerButtonBind) Axis() int {
	val, _ := binary.Varint(bind.value[:4])
	return int(val)
}

// Button returns button mapped for this SDL joystick layer binding.
func (bind *GameControllerButtonBind) Button() int {
	val, _ := binary.Varint(bind.value[:4])
	return int(val)
}

// Hat returns hat mapped for this SDL joystick layer binding.
func (bind *GameControllerButtonBind) Hat() int {
	val, _ := binary.Varint(bind.value[:4])
	return int(val)
}

// HatMask returns hat mask for this SDL joystick layer binding.
func (bind *GameControllerButtonBind) HatMask() int {
	val, _ := binary.Varint(bind.value[4:8])
	return int(val)
}

// Type returns the type of game controller input for this SDL joystick layer binding.
func (bind *GameControllerButtonBind) Type() int {
	return int(bind.bindType)
}

// GestureID is the unique id of the closest gesture to the performed stroke.
type GestureID int64

// Haptic identifies an SDL haptic.
// (https://wiki.libsdl.org/CategoryForceFeedback)
type Haptic struct{}

// HapticOpen opens a haptic device for use.
// (https://wiki.libsdl.org/SDL_HapticOpen)
func HapticOpen(index int) (*Haptic, error) {
	ret, _, _ := hapticOpen.Call(uintptr(index))
	if ret == 0 {
		return nil, GetError()
	}
	return (*Haptic)(unsafe.Pointer(ret)), nil
}

// HapticOpenFromJoystick opens a haptic device for use from a joystick device.
// (https://wiki.libsdl.org/SDL_HapticOpenFromJoystick)
func HapticOpenFromJoystick(joy *Joystick) (*Haptic, error) {
	ret, _, _ := hapticOpenFromJoystick.Call(uintptr(unsafe.Pointer(joy)))
	if ret == 0 {
		return nil, GetError()
	}
	return (*Haptic)(unsafe.Pointer(ret)), nil
}

// HapticOpenFromMouse open a haptic device from the current mouse.
// (https://wiki.libsdl.org/SDL_HapticOpenFromMouse)
func HapticOpenFromMouse() (*Haptic, error) {
	ret, _, _ := hapticOpenFromMouse.Call()
	if ret == 0 {
		return nil, GetError()
	}
	return (*Haptic)(unsafe.Pointer(ret)), nil
}

// Close closes a haptic device previously opened with HapticOpen().
// (https://wiki.libsdl.org/SDL_HapticClose)
func (h *Haptic) Close() {
	hapticClose.Call(uintptr(unsafe.Pointer(h)))
}

// DestroyEffect destroys a haptic effect on the device.
// (https://wiki.libsdl.org/SDL_HapticDestroyEffect)
func (h *Haptic) DestroyEffect(effect int) {
	hapticDestroyEffect.Call(uintptr(unsafe.Pointer(h)), uintptr(effect))
}

// EffectSupported reports whether an effect is supported by a haptic device.
// Pass pointer to a Haptic struct (Constant|Periodic|Condition|Ramp|LeftRight|Custom) instead of HapticEffect union.
// (https://wiki.libsdl.org/SDL_HapticEffectSupported)
func (h *Haptic) EffectSupported(he HapticEffect) (bool, error) {
	ret, _, _ := hapticEffectSupported.Call(
		uintptr(unsafe.Pointer(h)),
		he.pointer(),
	)
	return ret == 1, errorFromInt(int(ret))
}

// GetEffectStatus returns the status of the current effect on the specified haptic device.
// (https://wiki.libsdl.org/SDL_HapticGetEffectStatus)
func (h *Haptic) GetEffectStatus(effect int) (int, error) {
	ret, _, _ := hapticGetEffectStatus.Call(
		uintptr(unsafe.Pointer(h)),
		uintptr(effect),
	)
	return int(ret), errorFromInt(int(ret))
}

// NewEffect creates a new haptic effect on a specified device.
// Pass pointer to a Haptic struct (Constant|Periodic|Condition|Ramp|LeftRight|Custom) instead of HapticEffect union.
// (https://wiki.libsdl.org/SDL_HapticNewEffect)
func (h *Haptic) NewEffect(he HapticEffect) (int, error) {
	ret, _, _ := hapticNewEffect.Call(
		uintptr(unsafe.Pointer(h)),
		he.pointer(),
	)
	return int(ret), errorFromInt(int(ret))
}

// NumAxes returns the number of haptic axes the device has.
// (https://wiki.libsdl.org/SDL_HapticNumAxes)
func (h *Haptic) NumAxes() (int, error) {
	ret, _, _ := hapticNumAxes.Call(uintptr(unsafe.Pointer(h)))
	return int(ret), errorFromInt(int(ret))
}

// NumEffects returns the number of effects a haptic device can store.
// (https://wiki.libsdl.org/SDL_HapticNumEffects)
func (h *Haptic) NumEffects() (int, error) {
	ret, _, _ := hapticNumEffects.Call(uintptr(unsafe.Pointer(h)))
	return int(ret), errorFromInt(int(ret))
}

// NumEffectsPlaying returns the number of effects a haptic device can play at the same time.
// (https://wiki.libsdl.org/SDL_HapticNumEffectsPlaying)
func (h *Haptic) NumEffectsPlaying() (int, error) {
	ret, _, _ := hapticNumEffectsPlaying.Call(uintptr(unsafe.Pointer(h)))
	return int(ret), errorFromInt(int(ret))
}

// Pause pauses a haptic device.
// (https://wiki.libsdl.org/SDL_HapticPause)
func (h *Haptic) Pause() error {
	ret, _, _ := hapticPause.Call(uintptr(unsafe.Pointer(h)))
	return errorFromInt(int(ret))
}

// Query returns haptic device's supported features in bitwise manner.
// (https://wiki.libsdl.org/SDL_HapticQuery)
func (h *Haptic) Query() (uint32, error) {
	ret, _, _ := hapticQuery.Call(uintptr(unsafe.Pointer(h)))
	if ret == 0 {
		return 0, GetError()
	}
	return uint32(ret), nil
}

// RumbleInit initializes the haptic device for simple rumble playback.
// (https://wiki.libsdl.org/SDL_HapticRumbleInit)
func (h *Haptic) RumbleInit() error {
	ret, _, _ := hapticRumbleInit.Call(uintptr(unsafe.Pointer(h)))
	return errorFromInt(int(ret))
}

// RumblePlay runs a simple rumble effect on a haptic device.
// (https://wiki.libsdl.org/SDL_HapticRumblePlay)
func (h *Haptic) RumblePlay(strength float32, length uint32) error {
	ret, _, _ := hapticRumblePlay.Call(
		uintptr(unsafe.Pointer(h)),
		uintptr(strength),
		uintptr(length),
	)
	return errorFromInt(int(ret))
}

// RumbleStop stops the simple rumble on a haptic device.
// (https://wiki.libsdl.org/SDL_HapticRumbleStop)
func (h *Haptic) RumbleStop() error {
	ret, _, _ := hapticRumbleStop.Call(uintptr(unsafe.Pointer(h)))
	return errorFromInt(int(ret))
}

// RumbleSupported reports whether rumble is supported on a haptic device.
// (https://wiki.libsdl.org/SDL_HapticRumbleSupported)
func (h *Haptic) RumbleSupported() (bool, error) {
	ret, _, _ := hapticRumbleSupported.Call(uintptr(unsafe.Pointer(h)))
	return ret == 1, errorFromInt(int(ret))
}

// RunEffect runs the haptic effect on its associated haptic device.
// (https://wiki.libsdl.org/SDL_HapticRunEffect)
func (h *Haptic) RunEffect(effect int, iterations uint32) error {
	ret, _, _ := hapticRunEffect.Call(
		uintptr(unsafe.Pointer(h)),
		uintptr(effect),
		uintptr(iterations),
	)
	return errorFromInt(int(ret))
}

// SetAutocenter sets the global autocenter of the device.
// (https://wiki.libsdl.org/SDL_HapticSetAutocenter)
func (h *Haptic) SetAutocenter(autocenter int) error {
	ret, _, _ := hapticSetAutocenter.Call(
		uintptr(unsafe.Pointer(h)),
		uintptr(autocenter),
	)
	return errorFromInt(int(ret))
}

// SetGain sets the global gain of the specified haptic device.
// (https://wiki.libsdl.org/SDL_HapticSetGain)
func (h *Haptic) SetGain(gain int) error {
	ret, _, _ := hapticSetGain.Call(
		uintptr(unsafe.Pointer(h)),
		uintptr(gain),
	)
	return errorFromInt(int(ret))
}

// StopAll stops all the currently playing effects on a haptic device.
// (https://wiki.libsdl.org/SDL_HapticStopAll)
func (h *Haptic) StopAll() error {
	ret, _, _ := hapticStopAll.Call(uintptr(unsafe.Pointer(h)))
	return errorFromInt(int(ret))
}

// StopEffect stops the haptic effect on its associated haptic device.
// (https://wiki.libsdl.org/SDL_HapticStopEffect)
func (h *Haptic) StopEffect(effect int) error {
	ret, _, _ := hapticStopEffect.Call(
		uintptr(unsafe.Pointer(h)),
		uintptr(effect),
	)
	return errorFromInt(int(ret))
}

// Unpause unpauses a haptic device.
// (https://wiki.libsdl.org/SDL_HapticUnpause)
func (h *Haptic) Unpause() error {
	ret, _, _ := hapticUnpause.Call(uintptr(unsafe.Pointer(h)))
	return errorFromInt(int(ret))
}

// UpdateEffect updates the properties of an effect.
// Pass pointer to a Haptic struct (Constant|Periodic|Condition|Ramp|LeftRight|Custom) instead of HapticEffect union.
// (https://wiki.libsdl.org/SDL_HapticUpdateEffect)
func (h *Haptic) UpdateEffect(effect int, data HapticEffect) error {
	ret, _, _ := hapticUpdateEffect.Call(
		uintptr(unsafe.Pointer(h)),
		uintptr(effect),
		data.pointer(),
	)
	return errorFromInt(int(ret))
}

// HapticCondition contains a template for a condition effect.
// (https://wiki.libsdl.org/SDL_HapticCondition)
type HapticCondition struct {
	Type       uint16          // HAPTIC_SPRING, HAPTIC_DAMPER, HAPTIC_INERTIA, HAPTIC_FRICTION
	Direction  HapticDirection // direction of the effect - not used at the moment
	Length     uint32          // duration of the effect
	Delay      uint16          // delay before starting the effect
	Button     uint16          // button that triggers the effect
	Interval   uint16          // how soon it can be triggered again after button
	RightSat   [3]uint16       // level when joystick is to the positive side; max 0xFFFF
	LeftSat    [3]uint16       // level when joystick is to the negative side; max 0xFFFF
	RightCoeff [3]int16        // how fast to increase the force towards the positive side
	LeftCoeff  [3]int16        // how fast to increase the force towards the negative side
	Deadband   [3]uint16       // size of the dead zone; max 0xFFFF: whole axis-range when 0-centered
	Center     [3]int16        // position of the dead zone
}

func (he *HapticCondition) pointer() uintptr {
	return uintptr(unsafe.Pointer(he))
}

// HapticConstant contains a template for a constant effect.
// (https://wiki.libsdl.org/SDL_HapticConstant)
type HapticConstant struct {
	Type         uint16          // HAPTIC_CONSTANT
	Direction    HapticDirection // direction of the effect
	Length       uint32          // duration of the effect
	Delay        uint16          // delay before starting the effect
	Button       uint16          // button that triggers the effect
	Interval     uint16          // how soon it can be triggered again after button
	Level        int16           // strength of the constant effect
	AttackLength uint16          // duration of the attack
	AttackLevel  uint16          // level at the start of the attack
	FadeLength   uint16          // duration of the fade
	FadeLevel    uint16          // level at the end of the fade
}

func (he *HapticConstant) pointer() uintptr {
	return uintptr(unsafe.Pointer(he))
}

// HapticCustom contains a template for a custom effect.
// (https://wiki.libsdl.org/SDL_HapticCustom)
type HapticCustom struct {
	Type         uint16          // SDL_HAPTIC_CUSTOM
	Direction    HapticDirection // direction of the effect
	Length       uint32          // duration of the effect
	Delay        uint16          // delay before starting the effect
	Button       uint16          // button that triggers the effect
	Interval     uint16          // how soon it can be triggered again after button
	Channels     uint8           // axes to use, minimum of 1
	Period       uint16          // sample periods
	Samples      uint16          // amount of samples
	Data         *uint16         // should contain channels*samples items
	AttackLength uint16          // duration of the attack
	AttackLevel  uint16          // level at the start of the attack
	FadeLength   uint16          // duration of the fade
	FadeLevel    uint16          // level at the end of the fade
}

func (he *HapticCustom) pointer() uintptr {
	return uintptr(unsafe.Pointer(he))
}

// HapticDirection contains a haptic direction.
// (https://wiki.libsdl.org/SDL_HapticDirection)
type HapticDirection struct {
	Type byte     // the type of encoding
	Dir  [3]int32 // the encoded direction
}

// HapticEffect union that contains the generic template for any haptic effect.
// (https://wiki.libsdl.org/SDL_HapticEffect)
type HapticEffect interface {
	// original: cHapticEffect() *C.SDL_HapticEffect
	pointer() uintptr
}

// HapticLeftRight contains a template for a left/right effect.
// (https://wiki.libsdl.org/SDL_HapticLeftRight)
type HapticLeftRight struct {
	Type           uint16 // HAPTIC_LEFTRIGHT
	Length         uint32 // duration of the effect
	LargeMagnitude uint16 // control of the large controller motor
	SmallMagnitude uint16 // control of the small controller motor
}

func (he *HapticLeftRight) pointer() uintptr {
	return uintptr(unsafe.Pointer(he))
}

// HapticPeriodic contains a template for a periodic effect.
// (https://wiki.libsdl.org/SDL_HapticPeriodic)
type HapticPeriodic struct {
	Type         uint16          // HAPTIC_SINE, HAPTIC_LEFTRIGHT, HAPTIC_TRIANGLE, HAPTIC_SAWTOOTHUP, HAPTIC_SAWTOOTHDOWN
	Direction    HapticDirection // direction of the effect
	Length       uint32          // duration of the effect
	Delay        uint16          // delay before starting the effect
	Button       uint16          // button that triggers the effect
	Interval     uint16          // how soon it can be triggered again after button
	Period       uint16          // period of the wave
	Magnitude    int16           // peak value; if negative, equivalent to 180 degrees extra phase shift
	Offset       int16           // mean value of the wave
	Phase        uint16          // positive phase shift given by hundredth of a degree
	AttackLength uint16          // duration of the attack
	AttackLevel  uint16          // level at the start of the attack
	FadeLength   uint16          // duration of the fade
	FadeLevel    uint16          // level at the end of the fade
}

func (he *HapticPeriodic) pointer() uintptr {
	return uintptr(unsafe.Pointer(he))
}

// HapticRamp contains a template for a ramp effect.
// (https://wiki.libsdl.org/SDL_HapticRamp)
type HapticRamp struct {
	Type         uint16          // HAPTIC_RAMP
	Direction    HapticDirection // direction of the effect
	Length       uint32          // duration of the effect
	Delay        uint16          // delay before starting the effect
	Button       uint16          // button that triggers the effect
	Interval     uint16          // how soon it can be triggered again after button
	Start        int16           // beginning strength level
	End          int16           // ending strength level
	AttackLength uint16          // duration of the attack
	AttackLevel  uint16          // level at the start of the attack
	FadeLength   uint16          // duration of the fade
	FadeLevel    uint16          // level at the end of the fade
}

func (he *HapticRamp) pointer() uintptr {
	return uintptr(unsafe.Pointer(he))
}

// HintCallback is the function to call when the hint value changes.
type HintCallback func(data interface{}, name, oldValue, newValue string)

// HintCallbackAndData contains a callback function and userdata.
type HintCallbackAndData struct {
	callback HintCallback // the function to call when the hint value changes
	data     interface{}  // data to pass to the callback function
}

// HintPriority is a hint priority used in SetHintWithPriority().
// (https://wiki.libsdl.org/SDL_HintPriority)
type HintPriority uint32

// JoyAxisEvent contains joystick axis motion event information.
// (https://wiki.libsdl.org/SDL_JoyAxisEvent)
type JoyAxisEvent struct {
	Type      uint32     // JOYAXISMOTION
	Timestamp uint32     // timestamp of the event
	Which     JoystickID // the instance id of the joystick that reported the event
	Axis      uint8      // the index of the axis that changed
	_         uint8      // padding
	_         uint8      // padding
	_         uint8      // padding
	Value     int16      // the current position of the axis (range: -32768 to 32767)
	_         uint16     // padding
}

// GetTimestamp returns the timestamp of the event.
func (e *JoyAxisEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *JoyAxisEvent) GetType() uint32 {
	return e.Type
}

// JoyBallEvent contains joystick trackball motion event information.
// (https://wiki.libsdl.org/SDL_JoyBallEvent)
type JoyBallEvent struct {
	Type      uint32     // JOYBALLMOTION
	Timestamp uint32     // timestamp of the event
	Which     JoystickID // the instance id of the joystick that reported the event
	Ball      uint8      // the index of the trackball that changed
	_         uint8      // padding
	_         uint8      // padding
	_         uint8      // padding
	XRel      int16      // the relative motion in the X direction
	YRel      int16      // the relative motion in the Y direction
}

// GetTimestamp returns the timestamp of the event.
func (e *JoyBallEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *JoyBallEvent) GetType() uint32 {
	return e.Type
}

// JoyButtonEvent contains joystick button event information.
// (https://wiki.libsdl.org/SDL_JoyButtonEvent)
type JoyButtonEvent struct {
	Type      uint32     // JOYBUTTONDOWN, JOYBUTTONUP
	Timestamp uint32     // timestamp of the event
	Which     JoystickID // the instance id of the joystick that reported the event
	Button    uint8      // the index of the button that changed
	State     uint8      // PRESSED, RELEASED
	_         uint8      // padding
	_         uint8      // padding
}

// GetTimestamp returns the timestamp of the event.
func (e *JoyButtonEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *JoyButtonEvent) GetType() uint32 {
	return e.Type
}

// JoyDeviceAddedEvent contains joystick device event information.
// (https://wiki.libsdl.org/SDL_JoyDeviceEvent)
type JoyDeviceAddedEvent struct {
	Type      uint32 // JOYDEVICEADDED
	Timestamp uint32 // the timestamp of the event
	Which     int    // the joystick device index
}

// GetTimestamp returns the timestamp of the event.
func (e *JoyDeviceAddedEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *JoyDeviceAddedEvent) GetType() uint32 {
	return e.Type
}

// JoyDeviceRemovedEvent contains joystick device event information.
// (https://wiki.libsdl.org/SDL_JoyDeviceEvent)
type JoyDeviceRemovedEvent struct {
	Type      uint32     // JOYDEVICEREMOVED
	Timestamp uint32     // the timestamp of the event
	Which     JoystickID // the instance id
}

// GetTimestamp returns the timestamp of the event.
func (e *JoyDeviceRemovedEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *JoyDeviceRemovedEvent) GetType() uint32 {
	return e.Type
}

// JoyHatEvent contains joystick hat position change event information.
// (https://wiki.libsdl.org/SDL_JoyHatEvent)
type JoyHatEvent struct {
	Type      uint32     // JOYHATMOTION
	Timestamp uint32     // timestamp of the event
	Which     JoystickID // the instance id of the joystick that reported the event
	Hat       uint8      // the index of the hat that changed
	Value     uint8      // HAT_LEFTUP, HAT_UP, HAT_RIGHTUP, HAT_LEFT, HAT_CENTERED, HAT_RIGHT, HAT_LEFTDOWN, HAT_DOWN, HAT_RIGHTDOWN
	_         uint8      // padding
	_         uint8      // padding
}

// GetTimestamp returns the timestamp of the event.
func (e *JoyHatEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *JoyHatEvent) GetType() uint32 {
	return e.Type
}

// Joystick is an SDL joystick.
type Joystick struct{}

// JoystickFromInstanceID returns the Joystick associated with an instance id.
// (https://wiki.libsdl.org/SDL_JoystickFromInstanceID)
func JoystickFromInstanceID(joyid JoystickID) *Joystick {
	ret, _, _ := joystickFromInstanceID.Call(uintptr(joyid))
	return (*Joystick)(unsafe.Pointer(ret))
}

// JoystickOpen opens a joystick for use.
// (https://wiki.libsdl.org/SDL_JoystickOpen)
func JoystickOpen(index int) *Joystick {
	ret, _, _ := joystickOpen.Call(uintptr(index))
	return (*Joystick)(unsafe.Pointer(ret))
}

// Attached returns the status of a specified joystick.
// (https://wiki.libsdl.org/SDL_JoystickGetAttached)
func (joy *Joystick) Attached() bool {
	ret, _, _ := joystickGetAttached.Call(uintptr(unsafe.Pointer(joy)))
	return ret == 1
}

// Axis returns the current state of an axis control on a joystick.
// (https://wiki.libsdl.org/SDL_JoystickGetAxis)
func (joy *Joystick) Axis(axis int) int16 {
	ret, _, _ := joystickGetAxis.Call(
		uintptr(unsafe.Pointer(joy)),
		uintptr(axis),
	)
	return int16(ret)
}

// AxisInitialState returns the initial state of an axis control on a joystick, ok is true if this axis has any initial value.
func (joy *Joystick) AxisInitialState(axis int) (state int16, ok bool) {
	ret, _, _ := joystickGetAxisInitialState.Call(
		uintptr(unsafe.Pointer(joy)),
		uintptr(axis),
		uintptr(unsafe.Pointer(&state)),
	)
	ok = ret == 1
	return
}

// Ball returns the ball axis change since the last poll.
// (https://wiki.libsdl.org/SDL_JoystickGetBall)
func (joy *Joystick) Ball(ball int, dx, dy *int32) int {
	ret, _, _ := joystickGetBall.Call(
		uintptr(unsafe.Pointer(joy)),
		uintptr(ball),
		uintptr(unsafe.Pointer(dx)),
		uintptr(unsafe.Pointer(dy)),
	)
	return int(ret)
}

// Button the current state of a button on a joystick.
// (https://wiki.libsdl.org/SDL_JoystickGetButton)
func (joy *Joystick) Button(button int) byte {
	ret, _, _ := joystickGetButton.Call(
		uintptr(unsafe.Pointer(joy)),
		uintptr(button),
	)
	return byte(ret)
}

// Close closes a joystick previously opened with JoystickOpen().
// (https://wiki.libsdl.org/SDL_JoystickClose)
func (joy *Joystick) Close() {
	joystickClose.Call(uintptr(unsafe.Pointer(joy)))
}

// CurrentPowerLevel returns the battery level of a joystick as JoystickPowerLevel.
// (https://wiki.libsdl.org/SDL_JoystickCurrentPowerLevel)
func (joy *Joystick) CurrentPowerLevel() JoystickPowerLevel {
	ret, _, _ := joystickCurrentPowerLevel.Call(uintptr(unsafe.Pointer(joy)))
	return JoystickPowerLevel(ret)
}

// GUID returns the implementation-dependent GUID for the joystick.
// (https://wiki.libsdl.org/SDL_JoystickGetGUID)
func (joy *Joystick) GUID() JoystickGUID {
	ret, _, _ := joystickGetGUID.Call(uintptr(unsafe.Pointer(joy)))
	return *(*JoystickGUID)(unsafe.Pointer(ret))
}

// Hat returns the current state of a POV hat on a joystick.
// (https://wiki.libsdl.org/SDL_JoystickGetHat)
func (joy *Joystick) Hat(hat int) byte {
	ret, _, _ := joystickGetHat.Call(
		uintptr(unsafe.Pointer(joy)),
		uintptr(hat),
	)
	return byte(ret)
}

// InstanceID returns the instance ID of an opened joystick.
// (https://wiki.libsdl.org/SDL_JoystickInstanceID)
func (joy *Joystick) InstanceID() JoystickID {
	ret, _, _ := joystickInstanceID.Call(uintptr(unsafe.Pointer(joy)))
	return JoystickID(ret)
}

// Name returns the implementation dependent name of a joystick.
// (https://wiki.libsdl.org/SDL_JoystickName)
func (joy *Joystick) Name() string {
	ret, _, _ := joystickName.Call(uintptr(unsafe.Pointer(joy)))
	return sdlToGoString(ret)
}

// NumAxes returns the number of general axis controls on a joystick.
// (https://wiki.libsdl.org/SDL_JoystickNumAxes)
func (joy *Joystick) NumAxes() int {
	ret, _, _ := joystickNumAxes.Call(uintptr(unsafe.Pointer(joy)))
	return int(ret)
}

// NumBalls returns the number of trackballs on a joystick.
// (https://wiki.libsdl.org/SDL_JoystickNumBalls)
func (joy *Joystick) NumBalls() int {
	ret, _, _ := joystickNumBalls.Call(uintptr(unsafe.Pointer(joy)))
	return int(ret)
}

// NumButtons returns the number of buttons on a joystick.
// (https://wiki.libsdl.org/SDL_JoystickNumButtons)
func (joy *Joystick) NumButtons() int {
	ret, _, _ := joystickNumButtons.Call(uintptr(unsafe.Pointer(joy)))
	return int(ret)
}

// NumHats returns the number of POV hats on a joystick.
// (https://wiki.libsdl.org/SDL_JoystickNumHats)
func (joy *Joystick) NumHats() int {
	ret, _, _ := joystickNumHats.Call(uintptr(unsafe.Pointer(joy)))
	return int(ret)
}

// Product returns the USB product ID of an opened joystick, if available, 0 otherwise.
func (joy *Joystick) Product() int {
	ret, _, _ := joystickGetProduct.Call(uintptr(unsafe.Pointer(joy)))
	return int(ret)
}

// ProductVersion returns the product version of an opened joystick, if available, 0 otherwise.
func (joy *Joystick) ProductVersion() int {
	ret, _, _ := joystickGetProductVersion.Call(uintptr(unsafe.Pointer(joy)))
	return int(ret)
}

// Type returns the the type of an opened joystick.
func (joy *Joystick) Type() JoystickType {
	ret, _, _ := joystickGetType.Call(uintptr(unsafe.Pointer(joy)))
	return JoystickType(ret)
}

// Vendor returns the USB vendor ID of an opened joystick, if available, 0 otherwise.
func (joy *Joystick) Vendor() int {
	ret, _, _ := joystickGetVendor.Call(uintptr(unsafe.Pointer(joy)))
	return int(ret)
}

// JoystickGUID is a stable unique id for a joystick device.
type JoystickGUID struct {
	data [16]byte
}

// JoystickGetDeviceGUID returns the implementation dependent GUID for the joystick at a given device index.
// (https://wiki.libsdl.org/SDL_JoystickGetDeviceGUID)
func JoystickGetDeviceGUID(index int) JoystickGUID {
	ret, _, _ := joystickGetDeviceGUID.Call(uintptr(index))
	return *(*JoystickGUID)(unsafe.Pointer(ret))
}

// JoystickGetGUIDFromString converts a GUID string into a JoystickGUID structure.
// (https://wiki.libsdl.org/SDL_JoystickGetGUIDFromString)
func JoystickGetGUIDFromString(pchGUID string) JoystickGUID {
	g := append([]byte(pchGUID), 0)
	ret, _, _ := joystickGetGUIDFromString.Call(uintptr(unsafe.Pointer(&g[0])))
	return *(*JoystickGUID)(unsafe.Pointer(ret))
}

// JoystickID is joystick's instance id.
type JoystickID int32

// JoystickGetDeviceInstanceID returns the instance ID of a joystick.
func JoystickGetDeviceInstanceID(index int) JoystickID {
	ret, _, _ := joystickGetDeviceInstanceID.Call(uintptr(index))
	return JoystickID(ret)
}

// JoystickPowerLevel is a battery level of a joystick.
type JoystickPowerLevel uint32

// JoystickType is a type of a joystick.
type JoystickType uint32

// JoystickGetDeviceType returns the type of a joystick.
func JoystickGetDeviceType(index int) JoystickType {
	ret, _, _ := joystickGetDeviceType.Call(uintptr(index))
	return JoystickType(ret)
}

// KeyboardEvent contains keyboard key down event information.
// (https://wiki.libsdl.org/SDL_KeyboardEvent)
type KeyboardEvent struct {
	Type      uint32 // KEYDOWN, KEYUP
	Timestamp uint32 // timestamp of the event
	WindowID  uint32 // the window with keyboard focus, if any
	State     uint8  // PRESSED, RELEASED
	Repeat    uint8  // non-zero if this is a key repeat
	_         uint8  // padding
	_         uint8  // padding
	Keysym    Keysym // Keysym representing the key that was pressed or released
}

// GetTimestamp returns the timestamp of the event.
func (e *KeyboardEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *KeyboardEvent) GetType() uint32 {
	return e.Type
}

// Keycode is the SDL virtual key representation.
// (https://wiki.libsdl.org/SDL_Keycode)
type Keycode int32

// GetKeyFromName returns a key code from a human-readable name.
// (https://wiki.libsdl.org/SDL_GetKeyFromName)
func GetKeyFromName(name string) Keycode {
	n := append([]byte(name), 0)
	ret, _, _ := getKeyFromName.Call(uintptr(unsafe.Pointer(&n[0])))
	return Keycode(ret)
}

// GetKeyFromScancode returns the key code corresponding to the given scancode according to the current keyboard layout.
// (https://wiki.libsdl.org/SDL_GetKeyFromScancode)
func GetKeyFromScancode(code Scancode) Keycode {
	ret, _, _ := getKeyFromScancode.Call(uintptr(code))
	return Keycode(ret)
}

// Keymod is a key modifier masks.
// (https://wiki.libsdl.org/SDL_Keymod)
type Keymod uint32

// GetModState returns the current key modifier state for the keyboard.
// (https://wiki.libsdl.org/SDL_GetModState)
func GetModState() Keymod {
	ret, _, _ := getModState.Call()
	return Keymod(ret)
}

// Keysym contains key information used in key events.
// (https://wiki.libsdl.org/SDL_Keysym)
type Keysym struct {
	Scancode Scancode // SDL physical key code
	Sym      Keycode  // SDL virtual key code
	Mod      uint16   // current key modifiers
	unused   uint32   // unused
}

// LogOutputFunction is the function to call instead of the default
type LogOutputFunction func(data interface{}, category int, pri LogPriority, message string)

// LogGetOutputFunction returns the current log output function.
// (https://wiki.libsdl.org/SDL_LogGetOutputFunction)
func LogGetOutputFunction() (LogOutputFunction, interface{}) {
	// TODO
	return nil, nil
	//return logOutputFunctionCache, logOutputDataCache
}

// LogPriority is a predefined log priority.
// (https://wiki.libsdl.org/SDL_LogPriority)
type LogPriority uint32

// LogGetPriority returns the priority of a particular log category.
// (https://wiki.libsdl.org/SDL_LogGetPriority)
func LogGetPriority(category int) LogPriority {
	ret, _, _ := logGetPriority.Call(uintptr(category))
	return LogPriority(ret)
}

// MessageBoxButtonData contains individual button data for a message box.
// (https://wiki.libsdl.org/SDL_MessageBoxButtonData)
type MessageBoxButtonData struct {
	Flags    uint32 // MESSAGEBOX_BUTTON_RETURNKEY_DEFAULT, MESSAGEBOX_BUTTON_ESCAPEKEY_DEFAULT
	ButtonID int32  // user defined button id (value returned via ShowMessageBox())
	Text     string // the UTF-8 button text
}

// MessageBoxColor contains RGB value used in an MessageBoxColorScheme.
// (https://wiki.libsdl.org/SDL_MessageBoxColor)
type MessageBoxColor struct {
	R uint8 // the red component in the range 0-255
	G uint8 // the green component in the range 0-255
	B uint8 // the blue component in the range 0-255
}

// MessageBoxColorScheme contains a set of colors to use for message box dialogs.
// (https://wiki.libsdl.org/SDL_MessageBoxColorScheme)
type MessageBoxColorScheme struct {
	Colors [5]MessageBoxColor // background, text, button border, button background, button selected
}

// MessageBoxData contains title, text, window and other data for a message box.
// (https://wiki.libsdl.org/SDL_MessageBoxData)
type MessageBoxData struct {
	Flags       uint32                 // MESSAGEBOX_ERROR, MESSAGEBOX_WARNING, MESSAGEBOX_INFORMATION
	Window      *Window                // an parent window, can be nil
	Title       string                 // an UTF-8 title
	Message     string                 // an UTF-8 message text
	NumButtons  int32                  // the number of buttons
	Buttons     []MessageBoxButtonData // an array of MessageBoxButtonData with size of numbuttons
	ColorScheme *MessageBoxColorScheme // a MessageBoxColorScheme, can be nil to use system settings
}

// MouseButtonEvent contains mouse button event information.
// (https://wiki.libsdl.org/SDL_MouseButtonEvent)
type MouseButtonEvent struct {
	Type      uint32 // MOUSEBUTTONDOWN, MOUSEBUTTONUP
	Timestamp uint32 // timestamp of the event
	WindowID  uint32 // the window with mouse focus, if any
	Which     uint32 // the mouse instance id, or TOUCH_MOUSEID
	Button    uint8  // BUTTON_LEFT, BUTTON_MIDDLE, BUTTON_RIGHT, BUTTON_X1, BUTTON_X2
	State     uint8  // PRESSED, RELEASED
	_         uint8  // padding
	_         uint8  // padding
	X         int32  // X coordinate, relative to window
	Y         int32  // Y coordinate, relative to window
}

// GetTimestamp returns the timestamp of the event.
func (e *MouseButtonEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *MouseButtonEvent) GetType() uint32 {
	return e.Type
}

// MouseMotionEvent contains mouse motion event information.
// (https://wiki.libsdl.org/SDL_MouseMotionEvent)
type MouseMotionEvent struct {
	Type      uint32 // MOUSEMOTION
	Timestamp uint32 // timestamp of the event
	WindowID  uint32 // the window with mouse focus, if any
	Which     uint32 // the mouse instance id, or TOUCH_MOUSEID
	State     uint32 // BUTTON_LEFT, BUTTON_MIDDLE, BUTTON_RIGHT, BUTTON_X1, BUTTON_X2
	X         int32  // X coordinate, relative to window
	Y         int32  // Y coordinate, relative to window
	XRel      int32  // relative motion in the X direction
	YRel      int32  // relative motion in the Y direction
}

// GetTimestamp returns the timestamp of the event.
func (e *MouseMotionEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *MouseMotionEvent) GetType() uint32 {
	return e.Type
}

// MouseWheelEvent contains mouse wheel event information.
// (https://wiki.libsdl.org/SDL_MouseWheelEvent)
type MouseWheelEvent struct {
	Type      uint32 // MOUSEWHEEL
	Timestamp uint32 // timestamp of the event
	WindowID  uint32 // the window with mouse focus, if any
	Which     uint32 // the mouse instance id, or TOUCH_MOUSEID
	X         int32  // the amount scrolled horizontally, positive to the right and negative to the left
	Y         int32  // the amount scrolled vertically, positive away from the user and negative toward the user
	Direction uint32 // MOUSEWHEEL_NORMAL, MOUSEWHEEL_FLIPPED (>= SDL 2.0.4)
}

// GetTimestamp returns the timestamp of the event.
func (e *MouseWheelEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *MouseWheelEvent) GetType() uint32 {
	return e.Type
}

// MultiGestureEvent contains multiple finger gesture event information.
// (https://wiki.libsdl.org/SDL_MultiGestureEvent)
type MultiGestureEvent struct {
	Type       uint32  // MULTIGESTURE
	Timestamp  uint32  // timestamp of the event
	TouchID    TouchID // the touch device id
	DTheta     float32 // the amount that the fingers rotated during this motion
	DDist      float32 // the amount that the fingers pinched during this motion
	X          float32 // the normalized center of gesture
	Y          float32 // the normalized center of gesture
	NumFingers uint16  // the number of fingers used in the gesture
	_          uint16  // padding
}

// GetTimestamp returns the timestamp of the event.
func (e *MultiGestureEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *MultiGestureEvent) GetType() uint32 {
	return e.Type
}

// Mutex is the SDL mutex structure.
type Mutex struct {
	Recursive int
	Owner     ThreadID
	Sem       *Sem
}

// CreateMutex creates a new mutex.
// (https://wiki.libsdl.org/SDL_CreateMutex)
func CreateMutex() (*Mutex, error) {
	ret, _, _ := createMutex.Call()
	if ret == 0 {
		return nil, GetError()
	}
	return (*Mutex)(unsafe.Pointer(ret)), nil
}

// Destroy destroys a mutex created with CreateMutex().
// (https://wiki.libsdl.org/SDL_DestroyMutex)
func (mutex *Mutex) Destroy() {
	destroyMutex.Call(uintptr(unsafe.Pointer(mutex)))
}

// Lock locks a mutex created with CreateMutex().
// (https://wiki.libsdl.org/SDL_LockMutex)
func (mutex *Mutex) Lock() error {
	ret, _, _ := lockMutex.Call(uintptr(unsafe.Pointer(mutex)))
	if ret != 0 {
		return GetError()
	}
	return nil
}

// TryLock tries to lock a mutex without blocking.
// (https://wiki.libsdl.org/SDL_TryLockMutex)
func (mutex *Mutex) TryLock() error {
	ret, _, _ := tryLockMutex.Call(uintptr(unsafe.Pointer(mutex)))
	if ret != 0 {
		return GetError()
	}
	return nil
}

// Unlock unlocks a mutex created with CreateMutex().
// (https://wiki.libsdl.org/SDL_UnlockMutex)
func (mutex *Mutex) Unlock() error {
	ret, _, _ := unlockMutex.Call(uintptr(unsafe.Pointer(mutex)))
	if ret != 0 {
		return GetError()
	}
	return nil
}

// OSEvent contains OS specific event information.
type OSEvent struct {
	Type      uint32 // the event type
	Timestamp uint32 // timestamp of the event
}

// GetTimestamp returns the timestamp of the event.
func (e *OSEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *OSEvent) GetType() uint32 {
	return e.Type
}

// Palette contains palette information.
// (https://wiki.libsdl.org/SDL_Palette)
type Palette struct {
	Ncolors  int32  // the number of colors in the palette
	Colors   *Color // an array of Color structures representing the palette (https://wiki.libsdl.org/SDL_Color)
	version  uint32 // incrementally tracks changes to the palette (internal use)
	refCount int32  // reference count (internal use)
}

// AllocPalette creates a palette structure with the specified number of color entries.
// (https://wiki.libsdl.org/SDL_AllocPalette)
func AllocPalette(ncolors int) (*Palette, error) {
	ret, _, _ := allocPalette.Call(uintptr(ncolors))
	if ret == 0 {
		return nil, GetError()
	}
	return (*Palette)(unsafe.Pointer(ret)), nil
}

// Free frees the palette created with AllocPalette().
// (https://wiki.libsdl.org/SDL_FreePalette)
func (palette *Palette) Free() {
	freePalette.Call(uintptr(unsafe.Pointer(palette)))
}

// SetColors sets a range of colors in the palette.
// (https://wiki.libsdl.org/SDL_SetPaletteColors)
func (palette *Palette) SetColors(colors []Color) error {
	ret, _, _ := setPaletteColors.Call(
		uintptr(unsafe.Pointer(palette)),
		uintptr(unsafe.Pointer(&colors[0])),
		0,
		uintptr(len(colors)),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// PixelFormat contains pixel format information.
// (https://wiki.libsdl.org/SDL_PixelFormat)
type PixelFormat struct {
	Format        uint32       // one of the PIXELFORMAT values (https://wiki.libsdl.org/SDL_PixelFormatEnum)
	Palette       *Palette     // palette structure associated with this pixel format, or nil if the format doesn't have a palette (https://wiki.libsdl.org/SDL_Palette)
	BitsPerPixel  uint8        // the number of significant bits in a pixel value, eg: 8, 15, 16, 24, 32
	BytesPerPixel uint8        // the number of bytes required to hold a pixel value, eg: 1, 2, 3, 4
	_             [2]uint8     // padding
	Rmask         uint32       // a mask representing the location of the red component of the pixel
	Gmask         uint32       // a mask representing the location of the green component of the pixel
	Bmask         uint32       // a mask representing the location of the blue component of the pixel
	Amask         uint32       // a mask representing the location of the alpha component of the pixel or 0 if the pixel format doesn't have any alpha information
	rLoss         uint8        // (internal use)
	gLoss         uint8        // (internal use)
	bLoss         uint8        // (internal use)
	aLoss         uint8        // (internal use)
	rShift        uint8        // (internal use)
	gShift        uint8        // (internal use)
	bShift        uint8        // (internal use)
	aShift        uint8        // (internal use)
	refCount      int32        // (internal use)
	next          *PixelFormat // (internal use)
}

// AllocFormat creates a PixelFormat structure corresponding to a pixel format.
// (https://wiki.libsdl.org/SDL_AllocFormat)
func AllocFormat(format uint) (*PixelFormat, error) {
	ret, _, _ := allocFormat.Call(uintptr(format))
	if ret == 0 {
		return nil, GetError()
	}
	return (*PixelFormat)(unsafe.Pointer(ret)), nil
}

// Free frees the PixelFormat structure allocated by AllocFormat().
// (https://wiki.libsdl.org/SDL_FreeFormat)
func (format *PixelFormat) Free() {
	freeFormat.Call(uintptr(unsafe.Pointer(format)))
}

// SetPalette sets the palette for the pixel format structure.
// (https://wiki.libsdl.org/SDL_SetPixelFormatPalette)
func (format *PixelFormat) SetPalette(palette *Palette) error {
	ret, _, _ := setPixelFormatPalette.Call(
		uintptr(unsafe.Pointer(format)),
		uintptr(unsafe.Pointer(palette)),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// Point defines a two dimensional point.
// (https://wiki.libsdl.org/SDL_Point)
type Point struct {
	X int32 // the x coordinate of the point
	Y int32 // the y coordinate of the point
}

// InRect reports whether the point resides inside a rectangle.
// (https://wiki.libsdl.org/SDL_PointInRect)
func (p *Point) InRect(r *Rect) bool {
	if (p.X >= r.X) && (p.X < (r.X + r.W)) &&
		(p.Y >= r.Y) && (p.Y < (r.Y + r.H)) {
		return true
	}
	return false
}

// PowerState is the basic state for the system's power supply.
// (https://wiki.libsdl.org/SDL_PowerState)
type PowerState uint32

// QuitEvent contains the "quit requested" event.
// (https://wiki.libsdl.org/SDL_QuitEvent)
type QuitEvent struct {
	Type      uint32 // QUIT
	Timestamp uint32 // timestamp of the event
}

// GetTimestamp returns the timestamp of the event.
func (e *QuitEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *QuitEvent) GetType() uint32 {
	return e.Type
}

// RWops provides an abstract interface to stream I/O. Applications can generally ignore the specifics of this structure's internals and treat them as opaque pointers. The details are important to lower-level code that might need to implement one of these, however.
// (https://wiki.libsdl.org/SDL_RWops)
type RWops struct {
	size  uintptr
	seek  uintptr
	read  uintptr
	write uintptr
	close uintptr
	typ   uint32
}

// AllocRW allocates an empty, unpopulated RWops structure.
// (https://wiki.libsdl.org/SDL_AllocRW)
func AllocRW() *RWops {
	ret, _, _ := allocRW.Call()
	return (*RWops)(unsafe.Pointer(ret))
}

// RWFromFile creates a new RWops structure for reading from and/or writing to a named file.
// (https://wiki.libsdl.org/SDL_RWFromFile)
func RWFromFile(file, mode string) *RWops {
	f := append([]byte(file), 0)
	m := append([]byte(mode), 0)
	ret, _, _ := rwFromFile.Call(
		uintptr(unsafe.Pointer(&f[0])),
		uintptr(unsafe.Pointer(&m[0])),
	)
	return (*RWops)(unsafe.Pointer(ret))
}

// RWFromMem prepares a read-write memory buffer for use with RWops.
// (https://wiki.libsdl.org/SDL_RWFromMem)
func RWFromMem(mem []byte) (*RWops, error) {
	if len(mem) == 0 {
		return nil, ErrInvalidParameters
	}
	ret, _, _ := rwFromMem.Call(
		uintptr(unsafe.Pointer(&mem[0])),
		uintptr(len(mem)),
	)
	if ret == 0 {
		return nil, GetError()
	}
	return (*RWops)(unsafe.Pointer(ret)), nil
}

// Close closes and frees the allocated RWops structure.
// (https://wiki.libsdl.org/SDL_RWclose)
func (rwops *RWops) Close() error {
	ret, _, _ := rwClose.Call(uintptr(unsafe.Pointer(rwops)))
	if rwops != nil && ret != 0 {
		return GetError()
	}
	return nil
}

// Free frees the RWops structure allocated by AllocRW().
// (https://wiki.libsdl.org/SDL_FreeRW)
func (rwops *RWops) Free() error {
	if rwops == nil {
		return ErrInvalidParameters
	}
	freeRW.Call(uintptr(unsafe.Pointer(rwops)))
	return nil
}

// LoadFile_RW loads all the data from an SDL data stream.
// (https://wiki.libsdl.org/SDL_LoadFile_RW)
func (src *RWops) LoadFileRW(freesrc bool) (data []byte, size int) {
	// TODO
	return
	//var _size C.size_t
	//var _freesrc C.int = 0
	//
	//if freesrc {
	//	_freesrc = 1
	//}
	//
	//_data := C.SDL_LoadFile_RW(src.cptr(), &_size, _freesrc)
	//sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	//sliceHeader.Cap = int(_size)
	//sliceHeader.Len = int(_size)
	//sliceHeader.Data = uintptr(_data)
	//size = int(_size)
	//return
}

// Read reads from a data source.
// (https://wiki.libsdl.org/SDL_RWread)
func (rwops *RWops) Read(buf []byte) (n int, err error) {
	return rwops.Read2(buf, 1, uint(len(buf)))
}

// Read2 reads from a data source (native).
// (https://wiki.libsdl.org/SDL_RWread)
func (rwops *RWops) Read2(buf []byte, size, maxnum uint) (n int, err error) {
	if rwops == nil || buf == nil {
		return 0, ErrInvalidParameters
	}
	ret, _, _ := syscall.Syscall6(
		rwops.read,
		4,
		uintptr(unsafe.Pointer(rwops)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(size),
		uintptr(maxnum),
		0,
		0,
	)
	if ret == 0 {
		err = GetError()
	}
	n = int(ret)
	return
}

// ReadBE16 read 16 bits of big-endian data from the RWops and returns in native format.
// (https://wiki.libsdl.org/SDL_ReadBE16)
func (rwops *RWops) ReadBE16() uint16 {
	if rwops == nil {
		return 0
	}
	ret, _, _ := readBE16.Call(uintptr(unsafe.Pointer(rwops)))
	return uint16(ret)
}

// ReadBE32 reads 32 bits of big-endian data from the RWops and returns in native format.
// (https://wiki.libsdl.org/SDL_ReadBE32)
func (rwops *RWops) ReadBE32() uint32 {
	if rwops == nil {
		return 0
	}
	ret, _, _ := readBE32.Call(uintptr(unsafe.Pointer(rwops)))
	return uint32(ret)
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

// ReadLE16 reads 16 bits of little-endian data from the RWops and returns in native format.
// (https://wiki.libsdl.org/SDL_ReadLE16)
func (rwops *RWops) ReadLE16() uint16 {
	if rwops == nil {
		return 0
	}
	ret, _, _ := readLE16.Call(uintptr(unsafe.Pointer(rwops)))
	return uint16(ret)
}

// ReadLE32 reads 32 bits of little-endian data from the RWops and returns in native format.
// (https://wiki.libsdl.org/SDL_ReadLE32)
func (rwops *RWops) ReadLE32() uint32 {
	if rwops == nil {
		return 0
	}
	ret, _, _ := readLE32.Call(uintptr(unsafe.Pointer(rwops)))
	return uint32(ret)
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

// ReadU8 reads a byte from the RWops.
// (https://wiki.libsdl.org/SDL_ReadU8)
func (rwops *RWops) ReadU8() uint8 {
	if rwops == nil {
		return 0
	}
	ret, _, _ := readU8.Call(uintptr(unsafe.Pointer(rwops)))
	return uint8(ret)
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

// Tell returns the current read/write offset in the RWops data stream.
// (https://wiki.libsdl.org/SDL_RWtell)
func (rwops *RWops) Tell() (int64, error) {
	if rwops == nil {
		return 0, ErrInvalidParameters
	}
	ret, _, _ := syscall.Syscall(
		rwops.seek,
		3,
		uintptr(unsafe.Pointer(rwops)),
		0,
		uintptr(RW_SEEK_CUR),
	)
	ofs := int64(ret)
	if ofs < 0 {
		return ofs, GetError()
	}
	return ofs, nil
}

// Write writes to the RWops data stream.
// (https://wiki.libsdl.org/SDL_RWwrite)
func (rwops *RWops) Write(buf []byte) (n int, err error) {
	return rwops.Write2(buf, 1, uint(len(buf)))
}

// Write2 writes to the RWops data stream (native).
// (https://wiki.libsdl.org/SDL_RWwrite)
func (rwops *RWops) Write2(buf []byte, size, num uint) (n int, err error) {
	if rwops == nil || buf == nil {
		return 0, ErrInvalidParameters
	}
	ret, _, _ := syscall.Syscall6(
		rwops.write,
		4,
		uintptr(unsafe.Pointer(rwops)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(size),
		uintptr(num),
		0,
		0,
	)
	n = int(ret)
	if n < int(num) {
		err = GetError()
	}
	return
}

// WriteBE16 writes 16 bits in native format to the RWops as big-endian data.
// (https://wiki.libsdl.org/SDL_WriteBE16)
func (rwops *RWops) WriteBE16(value uint16) uint {
	if rwops == nil {
		return 0
	}
	ret, _, _ := writeBE16.Call(uintptr(unsafe.Pointer(rwops)), uintptr(value))
	return uint(ret)
}

// WriteBE32 writes 32 bits in native format to the RWops as big-endian data.
// (https://wiki.libsdl.org/SDL_WriteBE32)
func (rwops *RWops) WriteBE32(value uint32) uint {
	if rwops == nil {
		return 0
	}
	ret, _, _ := writeBE32.Call(uintptr(unsafe.Pointer(rwops)), uintptr(value))
	return uint(ret)
}

// WriteBE64 writes 64 bits in native format to the RWops as big-endian data.
// (https://wiki.libsdl.org/SDL_WriteBE64)
func (rwops *RWops) WriteBE64(value uint64) uint {
	if rwops == nil {
		return 0
	}
	// TODO what about 32 bit OS? uintptr is 32 bit there
	ret, _, _ := writeBE64.Call(uintptr(unsafe.Pointer(rwops)), uintptr(value))
	return uint(ret)
}

// WriteLE16 writes 16 bits in native format to the RWops as little-endian data.
// (https://wiki.libsdl.org/SDL_WriteLE16)
func (rwops *RWops) WriteLE16(value uint16) uint {
	if rwops == nil {
		return 0
	}
	ret, _, _ := writeLE16.Call(uintptr(unsafe.Pointer(rwops)), uintptr(value))
	return uint(ret)
}

// WriteLE32 writes 32 bits in native format to the RWops as little-endian data.
// (https://wiki.libsdl.org/SDL_WriteLE32)
func (rwops *RWops) WriteLE32(value uint32) uint {
	if rwops == nil {
		return 0
	}
	ret, _, _ := writeLE32.Call(uintptr(unsafe.Pointer(rwops)), uintptr(value))
	return uint(ret)
}

// WriteLE64 writes 64 bits in native format to the RWops as little-endian data.
// (https://wiki.libsdl.org/SDL_WriteLE64)
func (rwops *RWops) WriteLE64(value uint64) uint {
	if rwops == nil {
		return 0
	}
	// TODO what about 32 bit OS? uintptr is 32 bit there
	ret, _, _ := writeLE64.Call(uintptr(unsafe.Pointer(rwops)), uintptr(value))
	return uint(ret)
}

// WriteU8 writes a byte to the RWops.
// (https://wiki.libsdl.org/SDL_WriteU8)
func (rwops *RWops) WriteU8(value uint8) uint {
	if rwops == nil {
		return 0
	}
	ret, _, _ := writeU8.Call(uintptr(unsafe.Pointer(rwops)), uintptr(value))
	return uint(ret)
}

// Rect contains the definition of a rectangle, with the origin at the upper left.
// (https://wiki.libsdl.org/SDL_Rect)
type Rect struct {
	X int32 // the x location of the rectangle's upper left corner
	Y int32 // the y location of the rectangle's upper left corner
	W int32 // the width of the rectangle
	H int32 // the height of the rectangle
}

// EnclosePoints calculates a minimal rectangle that encloses a set of points.
// (https://wiki.libsdl.org/SDL_EnclosePoints)
func EnclosePoints(points []Point, clip *Rect) (Rect, bool) {
	var result Rect

	if len(points) == 0 {
		return result, false
	}

	var minX, minY, maxX, maxY int32
	if clip != nil {
		added := false
		clipMinX := clip.X
		clipMinY := clip.Y
		clipMaxX := clip.X + clip.W - 1
		clipMaxY := clip.Y + clip.H - 1

		// If the clip has no size, we're done
		if clip.Empty() {
			return result, false
		}

		for _, val := range points {
			// Check if the point is inside the clip rect
			if val.X < clipMinX || val.X > clipMaxX || val.Y < clipMinY || val.Y > clipMaxY {
				continue
			}

			if !added {
				// If it's the first point
				minX = val.X
				maxX = val.X
				minY = val.Y
				maxY = val.Y
				added = true
			}

			// Find mins and maxes
			if val.X < minX {
				minX = val.X
			} else if val.X > maxX {
				maxX = val.X
			}
			if val.Y < minY {
				minY = val.Y
			} else if val.Y > maxY {
				maxY = val.Y
			}
		}
	} else {
		for i, val := range points {
			if i == 0 {
				// Populate the first point
				minX = val.X
				maxX = val.X
				minY = val.Y
				maxY = val.Y
				continue
			}

			// Find mins and maxes
			if val.X < minX {
				minX = val.X
			} else if val.X > maxX {
				maxX = val.X
			}
			if val.Y < minY {
				minY = val.Y
			} else if val.Y > maxY {
				maxY = val.Y
			}
		}
	}

	result.X = minX
	result.Y = minY
	result.W = (maxX - minX) + 1
	result.H = (maxY - minY) + 1

	return result, true
}

// GetDisplayBounds returns the desktop area represented by a display, with the primary display located at 0,0.
// (https://wiki.libsdl.org/SDL_GetDisplayBounds)
func GetDisplayBounds(displayIndex int) (rect Rect, err error) {
	ret, _, _ := getDisplayBounds.Call(
		uintptr(displayIndex),
		uintptr(unsafe.Pointer(&rect)),
	)
	err = errorFromInt(int(ret))
	return
}

// GetDisplayUsableBounds returns the usable desktop area represented by a display, with the primary display located at 0,0.
// (https://wiki.libsdl.org/SDL_GetDisplayUsableBounds)
func GetDisplayUsableBounds(displayIndex int) (rect Rect, err error) {
	ret, _, _ := getDisplayUsableBounds.Call(
		uintptr(displayIndex),
		uintptr(unsafe.Pointer(&rect)),
	)
	err = errorFromInt(int(ret))
	return
}

// Empty reports whether a rectangle has no area.
// (https://wiki.libsdl.org/SDL_RectEmpty)
func (a *Rect) Empty() bool {
	return a == nil || a.W <= 0 || a.H <= 0
}

// Equals reports whether two rectangles are equal.
// (https://wiki.libsdl.org/SDL_RectEquals)
func (a *Rect) Equals(b *Rect) bool {
	if (a != nil) && (b != nil) &&
		(a.X == b.X) && (a.Y == b.Y) &&
		(a.W == b.W) && (a.H == b.H) {
		return true
	}
	return false
}

// HasIntersection reports whether two rectangles intersect.
// (https://wiki.libsdl.org/SDL_HasIntersection)
func (a *Rect) HasIntersection(b *Rect) bool {
	if a == nil || b == nil {
		return false
	}

	// Special case for empty rects
	if a.Empty() || b.Empty() {
		return false
	}

	if a.X >= b.X+b.W || a.X+a.W <= b.X || a.Y >= b.Y+b.H || a.Y+a.H <= b.Y {
		return false
	}

	return true
}

// Intersect calculates the intersection of two rectangles.
// (https://wiki.libsdl.org/SDL_IntersectRect)
func (a *Rect) Intersect(b *Rect) (Rect, bool) {
	var result Rect

	if a == nil || b == nil {
		return result, false
	}

	// Special case for empty rects
	if a.Empty() || b.Empty() {
		result.W = 0
		result.H = 0
		return result, false
	}

	aMin := a.X
	aMax := aMin + a.W
	bMin := b.X
	bMax := bMin + b.W
	if bMin > aMin {
		aMin = bMin
	}
	result.X = aMin
	if bMax < aMax {
		aMax = bMax
	}
	result.W = aMax - aMin

	aMin = a.Y
	aMax = aMin + a.H
	bMin = b.Y
	bMax = bMin + b.H
	if bMin > aMin {
		aMin = bMin
	}
	result.Y = aMin
	if bMax < aMax {
		aMax = bMax
	}
	result.H = aMax - aMin

	return result, !result.Empty()
}

// IntersectLine calculates the intersection of a rectangle and a line segment.
// (https://wiki.libsdl.org/SDL_IntersectRectAndLine)
func (a *Rect) IntersectLine(X1, Y1, X2, Y2 *int32) bool {
	if a.Empty() {
		return false
	}

	x1 := *X1
	y1 := *Y1
	x2 := *X2
	y2 := *Y2
	rectX1 := a.X
	rectY1 := a.Y
	rectX2 := a.X + a.W - 1
	rectY2 := a.Y + a.H - 1

	// Check if the line is entirely inside the rect
	if x1 >= rectX1 && x1 <= rectX2 && x2 >= rectX1 && x2 <= rectX2 &&
		y1 >= rectY1 && y1 <= rectY2 && y2 >= rectY1 && y2 <= rectY2 {
		return true
	}

	// Check if the line is entirely outside the rect
	if (x1 < rectX1 && x2 < rectX1) || (x1 > rectX2 && x2 > rectX2) ||
		(y1 < rectY1 && y2 < rectY1) || (y1 > rectY2 && y2 > rectY2) {
		return false
	}

	// Check if the line is horizontal
	if y1 == y2 {
		if x1 < rectX1 {
			*X1 = rectX1
		} else if x1 > rectX2 {
			*X1 = rectX2
		}
		if x2 < rectX1 {
			*X2 = rectX1
		} else if x2 > rectX2 {
			*X2 = rectX2
		}

		return true
	}

	// Check if the line is vertical
	if x1 == x2 {
		if y1 < rectY1 {
			*Y1 = rectY1
		} else if y1 > rectY2 {
			*Y1 = rectY2
		}
		if y2 < rectY1 {
			*Y2 = rectY1
		} else if y2 > rectY2 {
			*Y2 = rectY2
		}

		return true
	}

	// Use Cohen-Sutherland algorithm when all shortcuts fail
	outCode1 := computeOutCode(a, x1, y1)
	outCode2 := computeOutCode(a, x2, y2)
	for outCode1 != 0 || outCode2 != 0 {
		if outCode1&outCode2 != 0 {
			return false
		}

		if outCode1 != 0 {
			var x, y int32
			if outCode1&codeTop != 0 {
				y = rectY1
				x = x1 + ((x2-x1)*(y-y1))/(y2-y1)
			} else if outCode1&codeBottom != 0 {
				y = rectY2
				x = x1 + ((x2-x1)*(y-y1))/(y2-y1)
			} else if outCode1&codeLeft != 0 {
				x = rectX1
				y = y1 + ((y2-y1)*(x-x1))/(x2-x1)
			} else if outCode1&codeRight != 0 {
				x = rectX2
				y = y1 + ((y2-y1)*(x-x1))/(x2-x1)
			}

			x1 = x
			y1 = y
			outCode1 = computeOutCode(a, x, y)
		} else {
			var x, y int32
			if outCode2&codeTop != 0 {
				y = rectY1
				x = x1 + ((x2-x1)*(y-y1))/(y2-y1)
			} else if outCode2&codeBottom != 0 {
				y = rectY2
				x = x1 + ((x2-x1)*(y-y1))/(y2-y1)
			} else if outCode2&codeLeft != 0 {
				x = rectX1
				y = y1 + ((y2-y1)*(x-x1))/(x2-x1)
			} else if outCode2&codeRight != 0 {
				x = rectX2
				y = y1 + ((y2-y1)*(x-x1))/(x2-x1)
			}

			x2 = x
			y2 = y
			outCode2 = computeOutCode(a, x, y)
		}
	}

	*X1 = x1
	*Y1 = y1
	*X2 = x2
	*Y2 = y2

	return true
}

const (
	codeBottom = 1
	codeTop    = 2
	codeLeft   = 4
	codeRight  = 8
)

func computeOutCode(rect *Rect, x, y int32) int {
	code := 0
	if y < rect.Y {
		code |= codeTop
	} else if y >= rect.Y+rect.H {
		code |= codeBottom
	}
	if x < rect.X {
		code |= codeLeft
	} else if x >= rect.X+rect.W {
		code |= codeRight
	}

	return code
}

// Union calculates the union of two rectangles.
// (https://wiki.libsdl.org/SDL_UnionRect)
func (a *Rect) Union(b *Rect) Rect {
	var result Rect

	if a == nil || b == nil {
		return result
	}

	// Special case for empty rects
	if a.Empty() {
		return *b
	} else if b.Empty() {
		return *a
	} else if a.Empty() && b.Empty() {
		return result
	}

	aMin := a.X
	aMax := aMin + a.W
	bMin := b.X
	bMax := bMin + b.W
	if bMin < aMin {
		aMin = bMin
	}
	result.X = aMin
	if bMax > aMax {
		aMax = bMax
	}
	result.W = aMax - aMin

	aMin = a.Y
	aMax = aMin + a.H
	bMin = b.Y
	bMax = bMin + b.H
	if bMin < aMin {
		aMin = bMin
	}
	result.Y = aMin
	if bMax > aMax {
		aMax = bMax
	}
	result.H = aMax - aMin

	return result
}

// RenderEvent contains render event information.
// (https://wiki.libsdl.org/SDL_EventType)
type RenderEvent struct {
	Type      uint32 // the event type
	Timestamp uint32 // timestamp of the event
}

// GetTimestamp returns the timestamp of the event.
func (e *RenderEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *RenderEvent) GetType() uint32 {
	return e.Type
}

// Renderer contains a rendering state.
// (https://wiki.libsdl.org/SDL_Renderer)
type Renderer struct{}

// CreateRenderer returns a new 2D rendering context for a window.
// (https://wiki.libsdl.org/SDL_CreateRenderer)
func CreateRenderer(window *Window, index int, flags uint32) (*Renderer, error) {
	ret, _, _ := createRenderer.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(index),
		uintptr(flags),
	)
	if ret == 0 {
		return nil, GetError()
	}
	return (*Renderer)(unsafe.Pointer(ret)), nil
}

// CreateSoftwareRenderer returns a new 2D software rendering context for a surface.
// (https://wiki.libsdl.org/SDL_CreateSoftwareRenderer)
func CreateSoftwareRenderer(surface *Surface) (*Renderer, error) {
	ret, _, _ := createSoftwareRenderer.Call(uintptr(unsafe.Pointer(surface)))
	if ret == 0 {
		return nil, GetError()
	}
	return (*Renderer)(unsafe.Pointer(ret)), nil
}

// Clear clears the current rendering target with the drawing color.
// (https://wiki.libsdl.org/SDL_RenderClear)
func (renderer *Renderer) Clear() error {
	ret, _, _ := renderClear.Call(uintptr(unsafe.Pointer(renderer)))
	return errorFromInt(int(ret))
}

// Copy copies a portion of the texture to the current rendering target.
// (https://wiki.libsdl.org/SDL_RenderCopy)
func (renderer *Renderer) Copy(texture *Texture, src, dst *Rect) error {
	ret, _, _ := renderCopy.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(texture)),
		uintptr(unsafe.Pointer(src)),
		uintptr(unsafe.Pointer(dst)),
	)
	return errorFromInt(int(ret))
}

// CopyEx copies a portion of the texture to the current rendering target, optionally rotating it by angle around the given center and also flipping it top-bottom and/or left-right.
// (https://wiki.libsdl.org/SDL_RenderCopyEx)
func (renderer *Renderer) CopyEx(texture *Texture, src, dst *Rect, angle float64, center *Point, flip RendererFlip) error {
	ret, _, _ := renderCopyEx.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(texture)),
		uintptr(unsafe.Pointer(src)),
		uintptr(unsafe.Pointer(dst)),
		uintptr(angle),
		uintptr(unsafe.Pointer(center)),
		uintptr(flip),
	)
	return errorFromInt(int(ret))
}

// CreateTexture returns a new texture for a rendering context.
// (https://wiki.libsdl.org/SDL_CreateTexture)
func (renderer *Renderer) CreateTexture(format uint32, access int, w, h int32) (*Texture, error) {
	ret, _, _ := createTexture.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(format),
		uintptr(access),
		uintptr(w),
		uintptr(h),
	)
	if ret == 0 {
		return nil, GetError()
	}
	return (*Texture)(unsafe.Pointer(ret)), nil
}

// CreateTextureFromSurface returns a new texture from an existing surface.
// (https://wiki.libsdl.org/SDL_CreateTextureFromSurface)
func (renderer *Renderer) CreateTextureFromSurface(surface *Surface) (*Texture, error) {
	ret, _, _ := createTextureFromSurface.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(surface)),
	)
	if ret == 0 {
		return nil, GetError()
	}
	return (*Texture)(unsafe.Pointer(ret)), nil
}

// Destroy destroys the rendering context for a window and free associated textures.
// (https://wiki.libsdl.org/SDL_DestroyRenderer)
func (renderer *Renderer) Destroy() error {
	lastErr := GetError()
	ClearError()
	destroyRenderer.Call(uintptr(unsafe.Pointer(renderer)))
	err := GetError()
	if err != nil {
		return err
	}
	SetError(lastErr)
	return nil
}

// DrawLine draws a line on the current rendering target.
// (https://wiki.libsdl.org/SDL_RenderDrawLine)
func (renderer *Renderer) DrawLine(x1, y1, x2, y2 int32) error {
	ret, _, _ := renderDrawLine.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(x1),
		uintptr(y1),
		uintptr(x2),
		uintptr(y2),
	)

	return errorFromInt(int(ret))
}

// DrawLines draws a series of connected lines on the current rendering target.
// (https://wiki.libsdl.org/SDL_RenderDrawLines)
func (renderer *Renderer) DrawLines(points []Point) error {
	if points == nil {
		return nil // TODO this check should be in the original
	}
	ret, _, _ := renderDrawLines.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(&points[0])),
		uintptr(len(points)),
	)
	return errorFromInt(int(ret))
}

// DrawPoint draws a point on the current rendering target.
// (https://wiki.libsdl.org/SDL_RenderDrawPoint)
func (renderer *Renderer) DrawPoint(x, y int32) error {
	ret, _, _ := renderDrawPoint.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(x),
		uintptr(y),
	)
	return errorFromInt(int(ret))
}

// DrawPoints draws multiple points on the current rendering target.
// (https://wiki.libsdl.org/SDL_RenderDrawPoints)
func (renderer *Renderer) DrawPoints(points []Point) error {
	if points == nil {
		return nil // TODO this should be in the original as well
	}
	ret, _, _ := renderDrawPoints.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(&points[0])),
		uintptr(len(points)),
	)
	return errorFromInt(int(ret))
}

// DrawRect draws a rectangle on the current rendering target.
// (https://wiki.libsdl.org/SDL_RenderDrawRect)
func (renderer *Renderer) DrawRect(rect *Rect) error {
	ret, _, _ := renderDrawRect.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(rect)),
	)
	return errorFromInt(int(ret))
}

// DrawRects draws some number of rectangles on the current rendering target.
// (https://wiki.libsdl.org/SDL_RenderDrawRects)
func (renderer *Renderer) DrawRects(rects []Rect) error {
	if rects == nil {
		return nil // TODO this should be in the original as well
	}
	ret, _, _ := renderDrawRects.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(&rects[0])),
		uintptr(len(rects)),
	)
	return errorFromInt(int(ret))
}

// FillRect fills a rectangle on the current rendering target with the drawing color.
// (https://wiki.libsdl.org/SDL_RenderFillRect)
func (renderer *Renderer) FillRect(rect *Rect) error {
	ret, _, _ := renderFillRect.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(rect)),
	)
	return errorFromInt(int(ret))
}

// FillRects fills some number of rectangles on the current rendering target with the drawing color.
// (https://wiki.libsdl.org/SDL_RenderFillRects)
func (renderer *Renderer) FillRects(rects []Rect) error {
	if rects == nil {
		return nil // TODO this should be in the original as well
	}
	ret, _, _ := renderFillRects.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(&rects[0])),
		uintptr(len(rects)),
	)
	return errorFromInt(int(ret))
}

// GetClipRect returns the clip rectangle for the current target.
// (https://wiki.libsdl.org/SDL_RenderGetClipRect)
func (renderer *Renderer) GetClipRect() (rect Rect) {
	renderGetClipRect.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(&rect)),
	)
	return
}

// GetDrawBlendMode returns the blend mode used for drawing operations.
// (https://wiki.libsdl.org/SDL_GetRenderDrawBlendMode)
func (renderer *Renderer) GetDrawBlendMode(bm *BlendMode) error {
	ret, _, _ := getRenderDrawBlendMode.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(bm)),
	)
	return errorFromInt(int(ret))
}

// GetDrawColor returns the color used for drawing operations (Rect, Line and Clear).
// (https://wiki.libsdl.org/SDL_GetRenderDrawColor)
func (renderer *Renderer) GetDrawColor() (r, g, b, a uint8, err error) {
	ret, _, _ := getRenderDrawColor.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(&r)),
		uintptr(unsafe.Pointer(&g)),
		uintptr(unsafe.Pointer(&b)),
		uintptr(unsafe.Pointer(&a)),
	)
	err = errorFromInt(int(ret))
	return
}

// GetInfo returns information about a rendering context.
// (https://wiki.libsdl.org/SDL_GetRendererInfo)
func (renderer *Renderer) GetInfo() (r RendererInfo, err error) {
	ret, _, _ := getRendererInfo.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(&r)),
	)
	if ret != 0 {
		err = GetError()
	}
	return
}

// GetLogicalSize returns device independent resolution for rendering.
// (https://wiki.libsdl.org/SDL_RenderGetLogicalSize)
func (renderer *Renderer) GetLogicalSize() (w, h int32) {
	renderGetLogicalSize.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(&w)),
		uintptr(unsafe.Pointer(&h)),
	)
	return
}

// GetMetalCommandEncoder gets the Metal command encoder for the current frame
// (https://wiki.libsdl.org/SDL_RenderGetMetalCommandEncoder)
func (renderer *Renderer) GetMetalCommandEncoder() (encoder unsafe.Pointer, err error) {
	ret, _, _ := renderGetMetalCommandEncoder.Call(uintptr(unsafe.Pointer(renderer)))
	if ret == 0 {
		err = GetError()
	}
	encoder = unsafe.Pointer(ret)
	return
}

// GetMetalLayer gets the CAMetalLayer associated with the given Metal renderer
// (https://wiki.libsdl.org/SDL_RenderGetMetalLayer)
func (renderer *Renderer) GetMetalLayer() (layer unsafe.Pointer, err error) {
	ret, _, _ := renderGetMetalLayer.Call(uintptr(unsafe.Pointer(renderer)))
	if ret == 0 {
		err = GetError()
	}
	layer = unsafe.Pointer(ret)
	return
}

// GetOutputSize returns the output size in pixels of a rendering context.
// (https://wiki.libsdl.org/SDL_GetRendererOutputSize)
func (renderer *Renderer) GetOutputSize() (w, h int32, err error) {
	ret, _, _ := getRendererOutputSize.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(&w)),
		uintptr(unsafe.Pointer(&h)),
	)
	err = errorFromInt(int(ret))
	return
}

// GetRenderTarget returns the current render target.
// (https://wiki.libsdl.org/SDL_GetRenderTarget)
func (renderer *Renderer) GetRenderTarget() *Texture {
	ret, _, _ := getRenderTarget.Call(uintptr(unsafe.Pointer(renderer)))
	return (*Texture)(unsafe.Pointer(ret))
}

// GetScale returns the drawing scale for the current target.
// (https://wiki.libsdl.org/SDL_RenderGetScale)
func (renderer *Renderer) GetScale() (scaleX, scaleY float32) {
	renderGetScale.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(&scaleX)),
		uintptr(unsafe.Pointer(&scaleY)),
	)
	return
}

// GetViewport returns the drawing area for the current target.
// (https://wiki.libsdl.org/SDL_RenderGetViewport)
func (renderer *Renderer) GetViewport() (rect Rect) {
	renderGetViewport.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(&rect)),
	)
	return
}

// Present updates the screen with any rendering performed since the previous call.
// (https://wiki.libsdl.org/SDL_RenderPresent)
func (renderer *Renderer) Present() {
	renderPresent.Call(uintptr(unsafe.Pointer(renderer)))
}

// ReadPixels reads pixels from the current rendering target.
// (https://wiki.libsdl.org/SDL_RenderReadPixels)
func (renderer *Renderer) ReadPixels(rect *Rect, format uint32, pixels unsafe.Pointer, pitch int) error {
	ret, _, _ := renderReadPixels.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(rect)),
		uintptr(format),
		uintptr(pixels),
		uintptr(pitch),
	)
	return errorFromInt(int(ret))

}

// RenderTargetSupported reports whether a window supports the use of render targets.
// (https://wiki.libsdl.org/SDL_RenderTargetSupported)
func (renderer *Renderer) RenderTargetSupported() bool {
	ret, _, _ := renderTargetSupported.Call(uintptr(unsafe.Pointer(renderer)))
	return ret != 0
}

// SetClipRect sets the clip rectangle for rendering on the specified target.
// (https://wiki.libsdl.org/SDL_RenderSetClipRect)
func (renderer *Renderer) SetClipRect(rect *Rect) error {
	ret, _, _ := renderSetClipRect.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(rect)),
	)
	return errorFromInt(int(ret))
}

// SetDrawBlendMode sets the blend mode used for drawing operations (Fill and Line).
// (https://wiki.libsdl.org/SDL_SetRenderDrawBlendMode)
func (renderer *Renderer) SetDrawBlendMode(bm BlendMode) error {
	ret, _, _ := setRenderDrawBlendMode.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(bm),
	)
	return errorFromInt(int(ret))
}

// SetDrawColor sets the color used for drawing operations (Rect, Line and Clear).
// (https://wiki.libsdl.org/SDL_SetRenderDrawColor)
func (renderer *Renderer) SetDrawColor(r, g, b, a uint8) error {
	ret, _, _ := setRenderDrawColor.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(r),
		uintptr(g),
		uintptr(b),
		uintptr(a),
	)
	return errorFromInt(int(ret))
}

// SetDrawColorArray is a custom variant of SetDrawColor.
func (renderer *Renderer) SetDrawColorArray(bs ...uint8) error {
	// TODO
	return nil
	//_bs := []C.Uint8{0, 0, 0, 255}
	//for i := 0; i < len(_bs) && i < len(bs); i++ {
	//	_bs[i] = C.Uint8(bs[i])
	//}
	//return errorFromInt(int(
	//	C.SDL_SetRenderDrawColor(
	//		renderer.cptr(),
	//		_bs[0],
	//		_bs[1],
	//		_bs[2],
	//		_bs[3])))
}

// SetLogicalSize sets a device independent resolution for rendering.
// (https://wiki.libsdl.org/SDL_RenderSetLogicalSize)
func (renderer *Renderer) SetLogicalSize(w, h int32) error {
	ret, _, _ := renderSetLogicalSize.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(w),
		uintptr(h),
	)
	return errorFromInt(int(ret))
}

// SetRenderTarget sets a texture as the current rendering target.
// (https://wiki.libsdl.org/SDL_SetRenderTarget)
func (renderer *Renderer) SetRenderTarget(texture *Texture) error {
	ret, _, _ := setRenderTarget.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(texture)),
	)
	return errorFromInt(int(ret))
}

// SetScale sets the drawing scale for rendering on the current target.
// (https://wiki.libsdl.org/SDL_RenderSetScale)
func (renderer *Renderer) SetScale(scaleX, scaleY float32) error {
	ret, _, _ := renderSetScale.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(scaleX),
		uintptr(scaleY),
	)
	return errorFromInt(int(ret))
}

// SetViewport sets the drawing area for rendering on the current target.
// (https://wiki.libsdl.org/SDL_RenderSetViewport)
func (renderer *Renderer) SetViewport(rect *Rect) error {
	ret, _, _ := renderSetViewport.Call(
		uintptr(unsafe.Pointer(renderer)),
		uintptr(unsafe.Pointer(rect)),
	)
	return errorFromInt(int(ret))
}

// RendererFlip is an enumeration of flags that can be used in the flip parameter for Renderer.CopyEx().
// (https://wiki.libsdl.org/SDL_RendererFlip)
type RendererFlip uint32

// RendererInfo contains information on the capabilities of a render driver or the current render context.
// (https://wiki.libsdl.org/SDL_RendererInfo)
type RendererInfo struct {
	Name string // the name of the renderer
	RendererInfoData
}

// RendererInfoData contains information on the capabilities of a render driver or the current render context.
// (https://wiki.libsdl.org/SDL_RendererInfo)
type RendererInfoData struct {
	Flags             uint32    // a mask of supported renderer flags
	NumTextureFormats uint32    // the number of available texture formats
	TextureFormats    [16]int32 // the available texture formats
	MaxTextureWidth   int32     // the maximum texture width
	MaxTextureHeight  int32     // the maximum texture height
}

// Scancode is an SDL keyboard scancode representation.
// (https://wiki.libsdl.org/SDL_Scancode)
type Scancode uint32

// GetScancodeFromKey returns the scancode corresponding to the given key code according to the current keyboard layout.
// (https://wiki.libsdl.org/SDL_GetScancodeFromKey)
func GetScancodeFromKey(code Keycode) Scancode {
	ret, _, _ := getScancodeFromKey.Call(uintptr(code))
	return Scancode(ret)
}

// GetScancodeFromName returns a scancode from a human-readable name.
// (https://wiki.libsdl.org/SDL_GetScancodeFromName)
func GetScancodeFromName(name string) Scancode {
	n := append([]byte(name), 0)
	ret, _, _ := getScancodeFromName.Call(uintptr(unsafe.Pointer(&n[0])))
	return Scancode(ret)
}

// Sem is the SDL semaphore structure.
type Sem struct {
	Count        uint32
	WaitersCount uint32
	CountLock    *Mutex
	CountNonzero *Cond
}

// CreateSemaphore creates a semaphore.
// (https://wiki.libsdl.org/SDL_CreateSemaphore)
func CreateSemaphore(initialValue uint32) (*Sem, error) {
	ret, _, _ := createSemaphore.Call(uintptr(initialValue))
	if ret == 0 {
		return nil, GetError()
	}
	return (*Sem)(unsafe.Pointer(ret)), nil
}

// Destroy destroys a semaphore.
// (https://wiki.libsdl.org/SDL_DestroySemaphore)
func (sem *Sem) Destroy() {
	destroySemaphore.Call(uintptr(unsafe.Pointer(sem)))
}

// Post atomically increments a semaphore's value and wake waiting threads.
// (https://wiki.libsdl.org/SDL_SemPost)
func (sem *Sem) Post() error {
	ret, _, _ := semPost.Call(uintptr(unsafe.Pointer(sem)))
	if ret != 0 {
		return GetError()
	}
	return nil
}

// TryWait sees if a semaphore has a positive value and decrement it if it does.
// (https://wiki.libsdl.org/SDL_SemTryWait)
func (sem *Sem) TryWait() error {
	ret, _, _ := semTryWait.Call(uintptr(unsafe.Pointer(sem)))
	if ret != 0 {
		return GetError()
	}
	return nil
}

// Value returns the current value of a semaphore.
// (https://wiki.libsdl.org/SDL_SemValue)
func (sem *Sem) Value() uint32 {
	ret, _, _ := semValue.Call(uintptr(unsafe.Pointer(sem)))
	return uint32(ret)
}

// Wait waits until a semaphore has a positive value and then decrements it.
// (https://wiki.libsdl.org/SDL_SemWait)
func (sem *Sem) Wait() error {
	ret, _, _ := semWait.Call(uintptr(unsafe.Pointer(sem)))
	if ret != 0 {
		return GetError()
	}
	return nil
}

// WaitTimeout waits until a semaphore has a positive value and then decrements it.
// (https://wiki.libsdl.org/SDL_SemWaitTimeout)
func (sem *Sem) WaitTimeout(ms uint32) error {
	ret, _, _ := semWaitTimeout.Call(
		uintptr(unsafe.Pointer(sem)),
		uintptr(ms),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

type Sensor struct{}

// SensorFromInstanceID returns the Sensor associated with an instance id.
// (https://wiki.libsdl.org/SDL_SensorFromInstanceID)
func SensorFromInstanceID(id SensorID) (sensor *Sensor) {
	ret, _, _ := sensorFromInstanceID.Call(uintptr(id))
	return (*Sensor)(unsafe.Pointer(ret))
}

// SensorOpen opens a sensor for use.
//
// The index passed as an argument refers to the N'th sensor on the system.
//
// Returns a sensor identifier, or nil if an error occurred.
// (https://wiki.libsdl.org/SDL_SensorOpen)
func SensorOpen(deviceIndex int) (sensor *Sensor) {
	ret, _, _ := sensorOpen.Call(uintptr(deviceIndex))
	return (*Sensor)(unsafe.Pointer(ret))
}

// Close closes a sensor previously opened with SensorOpen()
// (https://wiki.libsdl.org/SDL_SensorClose)
func (sensor *Sensor) Close() {
	sensorClose.Call(uintptr(unsafe.Pointer(sensor)))
}

// GetData gets the current state of an opened sensor.
//
// The number of values and interpretation of the data is sensor dependent.
// (https://wiki.libsdl.org/SDL_SensorGetData)
func (sensor *Sensor) GetData(data []float32) (err error) {
	if data == nil {
		return nil // TODO should this be in the original too?
	}
	ret, _, _ := sensorGetData.Call(
		uintptr(unsafe.Pointer(sensor)),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
	)
	return errorFromInt(int(ret))
}

// GetInstanceID gets the instance ID of a sensor.
//
// This can be called before any sensors are opened.
//
// Returns the sensor instance ID, or -1 if the sensor is nil.
// (https://wiki.libsdl.org/SDL_SensorGetInstanceID)
func (sensor *Sensor) GetInstanceID() (id SensorID) {
	ret, _, _ := sensorGetInstanceID.Call(
		uintptr(unsafe.Pointer(sensor)),
		uintptr(id),
	)
	return SensorID(ret)
}

// GetName gets the implementation dependent name of a sensor.
//
// Returns the sensor name, or empty string if the sensor is nil.
// (https://wiki.libsdl.org/SDL_SensorGetName)
func (sensor *Sensor) GetName() (name string) {
	ret, _, _ := sensorGetName.Call(uintptr(unsafe.Pointer(sensor)))
	return sdlToGoString(ret)
}

// GetNonPortableType gets the platform dependent type of a sensor.
//
// This can be called before any sensors are opened.
//
// Returns the sensor platform dependent type, or -1 if the sensor is nil.
// (https://wiki.libsdl.org/SDL_SensorGetNonPortableType)
func (sensor *Sensor) GetNonPortableType() (typ int) {
	ret, _, _ := sensorGetNonPortableType.Call(uintptr(unsafe.Pointer(sensor)))
	return int(ret)
}

// GetType gets the type of a sensor.
//
// This can be called before any sensors are opened.
//
// Returns the sensor type, or SENSOR_INVALID if the sensor is nil.
// (https://wiki.libsdl.org/SDL_SensorGetType)
func (sensor *Sensor) GetType() (typ SensorType) {
	ret, _, _ := sensorGetType.Call(uintptr(unsafe.Pointer(sensor)))
	return SensorType(ret)
}

// SensorEvent contains data from sensors such as accelerometer and gyroscope
// (https://wiki.libsdl.org/SDL_SensorEvent)
type SensorEvent struct {
	Type      uint32     // SDL_SENSORUPDATE
	Timestamp uint32     // In milliseconds, populated using SDL_GetTicks()
	Which     int32      // The instance ID of the sensor
	Data      [6]float32 // Up to 6 values from the sensor - additional values can be queried using SDL_SensorGetData()
}

// GetTimestamp returns the timestamp of the event.
func (e *SensorEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *SensorEvent) GetType() uint32 {
	return e.Type
}

type SensorID int32

// SensorGetDeviceInstanceID gets the instance ID of a sensor.
//
// This can be called before any sensors are opened.
//
// Returns the sensor instance ID, or -1 if deviceIndex is out of range.
// (https://wiki.libsdl.org/SDL_SensorGetDeviceInstanceID)
func SensorGetDeviceInstanceID(deviceIndex int) (id SensorID) {
	ret, _, _ := sensorGetDeviceInstanceID.Call(uintptr(deviceIndex))
	return SensorID(ret)
}

type SensorType int

// The different sensors defined by SDL
//
// Additional sensors may be available, using platform dependent semantics.
//
// Here are the additional Android sensors:
// https://developer.android.com/reference/android/hardware/SensorEvent.html#values
const (
	SENSOR_INVALID SensorType = -1 // Returned for an invalid sensor
	SENSOR_UNKNOWN                 // Unknown sensor type
	SENSOR_ACCEL                   // Accelerometer
	SENSOR_GYRO                    // Gyroscope
)

//  SensorGetDeviceType gets the type of a sensor.
//
//  This can be called before any sensors are opened.
//
//  Returns the sensor type, or SDL_SENSOR_INVALID if deviceIndex is out of range.
// (https://wiki.libsdl.org/SDL_SensorGetDeviceType)
func SensorGetDeviceType(deviceIndex int) (typ SensorType) {
	ret, _, _ := sensorGetDeviceType.Call(uintptr(deviceIndex))
	return SensorType(ret)
}

// SharedObject is a pointer to the object handle.
type SharedObject uintptr

// LoadObject dynamically loads a shared object and returns a pointer to the object handle.
// (https://wiki.libsdl.org/SDL_LoadObject)
func LoadObject(sofile string) SharedObject {
	s := append([]byte(sofile), 0)
	ret, _, _ := loadObject.Call(uintptr(unsafe.Pointer(&s[0])))
	return SharedObject(ret)
}

// LoadFunction returns a pointer to the named function from the shared object.
// (https://wiki.libsdl.org/SDL_LoadFunction)
func (handle SharedObject) LoadFunction(name string) unsafe.Pointer {
	n := append([]byte(name), 0)
	ret, _, _ := loadFunction.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&n[0])),
	)
	return unsafe.Pointer(ret)
}

// Unload unloads a shared object from memory.
// (https://wiki.libsdl.org/SDL_UnloadObject)
func (handle SharedObject) Unload() {
	unloadObject.Call(uintptr(handle))
}

// Surface contains a collection of pixels used in software blitting.
// (https://wiki.libsdl.org/SDL_Surface)
type Surface struct {
	flags    uint32         // (internal use)
	Format   *PixelFormat   // the format of the pixels stored in the surface (read-only) (https://wiki.libsdl.org/SDL_PixelFormat)
	W        int32          // the width in pixels (read-only)
	H        int32          // the height in pixels (read-only)
	Pitch    int32          // the length of a row of pixels in bytes (read-only)
	pixels   unsafe.Pointer // the pointer to the actual pixel data; use Pixels() for access
	UserData unsafe.Pointer // an arbitrary pointer you can set
	locked   int32          // used for surfaces that require locking (internal use)
	lockData unsafe.Pointer // used for surfaces that require locking (internal use)
	ClipRect Rect           // a Rect structure used to clip blits to the surface which can be set by SetClipRect() (read-only)
	_        unsafe.Pointer // map; info for fast blit mapping to other surfaces (internal use)
	RefCount int32          // reference count that can be incremented by the application
}

// CreateRGBSurface allocates a new RGB surface.
// (https://wiki.libsdl.org/SDL_CreateRGBSurface)
func CreateRGBSurface(flags uint32, width, height, depth int32, Rmask, Gmask, Bmask, Amask uint32) (*Surface, error) {
	ret, _, _ := createRGBSurface.Call(
		uintptr(flags),
		uintptr(width),
		uintptr(height),
		uintptr(depth),
		uintptr(Rmask),
		uintptr(Gmask),
		uintptr(Bmask),
		uintptr(Amask),
	)
	if ret == 0 {
		return nil, GetError()
	}
	return (*Surface)(unsafe.Pointer(ret)), nil
}

// CreateRGBSurfaceFrom allocate a new RGB surface with existing pixel data.
// (https://wiki.libsdl.org/SDL_CreateRGBSurfaceFrom)
func CreateRGBSurfaceFrom(pixels unsafe.Pointer, width, height int32, depth, pitch int, Rmask, Gmask, Bmask, Amask uint32) (*Surface, error) {
	ret, _, _ := createRGBSurfaceFrom.Call(
		uintptr(pixels),
		uintptr(width),
		uintptr(height),
		uintptr(depth),
		uintptr(pitch),
		uintptr(Rmask),
		uintptr(Gmask),
		uintptr(Bmask),
		uintptr(Amask),
	)
	if ret == 0 {
		return nil, GetError()
	}
	return (*Surface)(unsafe.Pointer(ret)), nil
}

// CreateRGBSurfaceWithFormat allocates an RGB surface.
// (https://wiki.libsdl.org/SDL_CreateRGBSurfaceWithFormat)
func CreateRGBSurfaceWithFormat(flags uint32, width, height, depth int32, format uint32) (*Surface, error) {
	ret, _, _ := createRGBSurfaceWithFormat.Call(
		uintptr(flags),
		uintptr(width),
		uintptr(height),
		uintptr(depth),
		uintptr(format),
	)
	if ret == 0 {
		return nil, GetError()
	}
	return (*Surface)(unsafe.Pointer(ret)), nil
}

// CreateRGBSurfaceWithFormatFrom allocates an RGB surface from provided pixel data.
// (https://wiki.libsdl.org/SDL_CreateRGBSurfaceWithFormatFrom)
func CreateRGBSurfaceWithFormatFrom(pixels unsafe.Pointer, width, height, depth, pitch int32, format uint32) (*Surface, error) {
	ret, _, _ := createRGBSurfaceWithFormatFrom.Call(
		uintptr(pixels),
		uintptr(width),
		uintptr(height),
		uintptr(depth),
		uintptr(pitch),
		uintptr(format),
	)
	if ret == 0 {
		return nil, GetError()
	}
	return (*Surface)(unsafe.Pointer(ret)), nil
}

// LoadBMP loads a surface from a BMP file.
// (https://wiki.libsdl.org/SDL_LoadBMP)
func LoadBMP(file string) (*Surface, error) {
	return LoadBMPRW(RWFromFile(file, "rb"), true)
}

// LoadBMPRW loads a BMP image from a seekable SDL data stream (memory or file).
// (https://wiki.libsdl.org/SDL_LoadBMP_RW)
func LoadBMPRW(src *RWops, freeSrc bool) (*Surface, error) {
	ret, _, _ := loadBMP_RW.Call(
		uintptr(unsafe.Pointer(src)),
		uintptr(Btoi(freeSrc)),
	)
	if ret == 0 {
		return nil, GetError()
	}
	return (*Surface)(unsafe.Pointer(ret)), nil
}

// At returns the pixel color at (x, y)
func (surface *Surface) At(x, y int) color.Color {
	pix := surface.Pixels()
	i := int32(y)*surface.Pitch + int32(x)*int32(surface.Format.BytesPerPixel)
	switch surface.Format.Format {
	/*
		case PIXELFORMAT_ARGB8888:
			return color.RGBA{pix[i+3], pix[i], pix[i+1], pix[i+2]}
		case PIXELFORMAT_ABGR8888:
			return color.RGBA{pix[i], pix[i+3], pix[i+2], pix[i+1]}
	*/
	case PIXELFORMAT_RGB888:
		return color.RGBA{pix[i], pix[i+1], pix[i+2], 0xff}
	default:
		panic("Not implemented yet")
	}
}

// Blit performs a fast surface copy to a destination surface.
// (https://wiki.libsdl.org/SDL_BlitSurface)
func (surface *Surface) Blit(srcRect *Rect, dst *Surface, dstRect *Rect) error {
	ret, _, _ := blitSurface.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(srcRect)),
		uintptr(unsafe.Pointer(dst)),
		uintptr(unsafe.Pointer(dstRect)),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// BlitScaled performs a scaled surface copy to a destination surface.
// (https://wiki.libsdl.org/SDL_BlitScaled)
func (surface *Surface) BlitScaled(srcRect *Rect, dst *Surface, dstRect *Rect) error {
	ret, _, _ := blitScaled.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(srcRect)),
		uintptr(unsafe.Pointer(dst)),
		uintptr(unsafe.Pointer(dstRect)),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// Bounds return the bounds of this surface. Currently, it always starts at
// (0,0), but this is not guaranteed in the future so don't rely on it.
func (surface *Surface) Bounds() image.Rectangle {
	return image.Rect(0, 0, int(surface.W), int(surface.H))
}

// BytesPerPixel return the number of significant bits in a pixel values of the surface.
func (surface *Surface) BytesPerPixel() int {
	return int(surface.Format.BytesPerPixel)
}

// ColorModel returns the color model used by this Surface.
func (surface *Surface) ColorModel() color.Model {
	switch surface.Format.Format {
	case PIXELFORMAT_ARGB8888, PIXELFORMAT_ABGR8888:
		return color.RGBAModel
	case PIXELFORMAT_RGB888:
		return color.RGBAModel
	default:
		panic("Not implemented yet")
	}
}

// Convert copies the existing surface into a new one that is optimized for blitting to a surface of a specified pixel format.
// (https://wiki.libsdl.org/SDL_ConvertSurface)
func (surface *Surface) Convert(fmt *PixelFormat, flags uint32) (*Surface, error) {
	ret, _, _ := convertSurface.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(fmt)),
		uintptr(flags),
	)
	if ret != 0 {
		return nil, GetError()
	}
	return (*Surface)(unsafe.Pointer(ret)), nil
}

// ConvertFormat copies the existing surface to a new surface of the specified format.
// (https://wiki.libsdl.org/SDL_ConvertSurfaceFormat)
func (surface *Surface) ConvertFormat(pixelFormat uint32, flags uint32) (*Surface, error) {
	ret, _, _ := convertSurfaceFormat.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(pixelFormat),
		uintptr(flags),
	)
	if ret == 0 {
		return nil, GetError()
	}
	return (*Surface)(unsafe.Pointer(ret)), nil
}

// Data returns the pointer to the actual pixel data of the surface.
func (surface *Surface) Data() unsafe.Pointer {
	return surface.pixels
}

// Duplicate creates a new surface identical to the existing surface
func (surface *Surface) Duplicate() (newSurface *Surface, err error) {
	ret, _, _ := duplicateSurface.Call(uintptr(unsafe.Pointer(surface)))
	if ret == 0 {
		return nil, GetError()
	}
	return (*Surface)(unsafe.Pointer(ret)), nil
}

// FillRect performs a fast fill of a rectangle with a specific color.
// (https://wiki.libsdl.org/SDL_FillRect)
func (surface *Surface) FillRect(rect *Rect, color uint32) error {
	ret, _, _ := fillRect.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(rect)),
		uintptr(color),
	)
	if ret == 0 {
		return GetError()
	}
	return nil
}

// FillRects performs a fast fill of a set of rectangles with a specific color.
// (https://wiki.libsdl.org/SDL_FillRects)
func (surface *Surface) FillRects(rects []Rect, color uint32) error {
	if rects == nil {
		return nil // TODO this should be in the original as well
	}
	ret, _, _ := fillRects.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(&rects[0])),
		uintptr(len(rects)),
		uintptr(color),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// Free frees the RGB surface.
// (https://wiki.libsdl.org/SDL_FreeSurface)
func (surface *Surface) Free() {
	freeSurface.Call(uintptr(unsafe.Pointer(surface)))
}

// GetAlphaMod returns the additional alpha value used in blit operations.
// (https://wiki.libsdl.org/SDL_GetSurfaceAlphaMod)
func (surface *Surface) GetAlphaMod() (alpha uint8, err error) {
	ret, _, _ := getSurfaceAlphaMod.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(&alpha)),
	)
	if ret != 0 {
		err = GetError()
	}
	return
}

// GetBlendMode returns the blend mode used for blit operations.
// (https://wiki.libsdl.org/SDL_GetSurfaceBlendMode)
func (surface *Surface) GetBlendMode() (bm BlendMode, err error) {
	ret, _, _ := getSurfaceBlendMode.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(&bm)),
	)
	if ret != 0 {
		err = GetError()
	}
	return
}

// GetClipRect returns the clipping rectangle for a surface.
// (https://wiki.libsdl.org/SDL_GetClipRect)
func (surface *Surface) GetClipRect(rect *Rect) {
	getClipRect.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(rect)),
	)
}

// GetColorKey retruns the color key (transparent pixel) for the surface.
// (https://wiki.libsdl.org/SDL_GetColorKey)
func (surface *Surface) GetColorKey() (key uint32, err error) {
	ret, _, _ := getColorKey.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(&key)),
	)
	if ret != 0 {
		err = GetError()
	}
	return
}

// GetColorMod returns the additional color value multiplied into blit operations.
// (https://wiki.libsdl.org/SDL_GetSurfaceColorMod)
func (surface *Surface) GetColorMod() (r, g, b uint8, err error) {
	ret, _, _ := getSurfaceColorMod.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(&r)),
		uintptr(unsafe.Pointer(&g)),
		uintptr(unsafe.Pointer(&b)),
	)
	if ret != 0 {
		err = GetError()
	}
	return
}

// Lock sets up the surface for directly accessing the pixels.
// (https://wiki.libsdl.org/SDL_LockSurface)
func (surface *Surface) Lock() error {
	ret, _, _ := lockSurface.Call(uintptr(unsafe.Pointer(surface)))
	if ret != 0 {
		return GetError()
	}
	return nil
}

// LowerBlit performs low-level surface blitting only.
// (https://wiki.libsdl.org/SDL_LowerBlit)
func (surface *Surface) LowerBlit(srcRect *Rect, dst *Surface, dstRect *Rect) error {
	ret, _, _ := lowerBlit.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(srcRect)),
		uintptr(unsafe.Pointer(dst)),
		uintptr(unsafe.Pointer(dstRect)),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// LowerBlitScaled performs low-level surface scaled blitting only.
// (https://wiki.libsdl.org/SDL_LowerBlitScaled)
func (surface *Surface) LowerBlitScaled(srcRect *Rect, dst *Surface, dstRect *Rect) error {
	ret, _, _ := lowerBlitScaled.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(srcRect)),
		uintptr(unsafe.Pointer(dst)),
		uintptr(unsafe.Pointer(dstRect)),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// MustLock reports whether the surface must be locked for access.
// (https://wiki.libsdl.org/SDL_MUSTLOCK)
func (surface *Surface) MustLock() bool {
	return (surface.flags & RLEACCEL) != 0
}

// PixelNum returns the number of pixels stored in the surface.
func (surface *Surface) PixelNum() int {
	return int(surface.W * surface.H)
}

// Pixels returns the actual pixel data of the surface.
func (surface *Surface) Pixels() []byte {
	var b []byte
	length := int(surface.W*surface.H) * int(surface.Format.BytesPerPixel)
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sliceHeader.Cap = int(length)
	sliceHeader.Len = int(length)
	sliceHeader.Data = uintptr(surface.pixels)
	return b
}

// SaveBMP saves the surface to a BMP file.
// (https://wiki.libsdl.org/SDL_SaveBMP)
func (surface *Surface) SaveBMP(file string) error {
	return surface.SaveBMPRW(RWFromFile(file, "wb"), true)
}

// SaveBMPRW save the surface to a seekable SDL data stream (memory or file) in BMP format.
// (https://wiki.libsdl.org/SDL_SaveBMP_RW)
func (surface *Surface) SaveBMPRW(dst *RWops, freeDst bool) error {
	ret, _, _ := saveBMP_RW.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(dst)),
		uintptr(Btoi(freeDst)),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// Set the color of the pixel at (x, y) using this surface's color model to
// convert c to the appropriate color. This method is required for the
// draw.Image interface. The surface may require locking before calling Set.
func (surface *Surface) Set(x, y int, c color.Color) {
	pix := surface.Pixels()
	i := int32(y)*surface.Pitch + int32(x)*int32(surface.Format.BytesPerPixel)
	switch surface.Format.Format {
	case PIXELFORMAT_ARGB8888:
		col := surface.ColorModel().Convert(c).(color.RGBA)
		pix[i+0] = col.R
		pix[i+1] = col.G
		pix[i+2] = col.B
		pix[i+3] = col.A
	case PIXELFORMAT_ABGR8888:
		col := surface.ColorModel().Convert(c).(color.RGBA)
		pix[i+3] = col.R
		pix[i+2] = col.G
		pix[i+1] = col.B
		pix[i+0] = col.A
	case PIXELFORMAT_RGB24, PIXELFORMAT_RGB888:
		col := surface.ColorModel().Convert(c).(color.RGBA)
		pix[i+0] = col.B
		pix[i+1] = col.G
		pix[i+2] = col.R
	case PIXELFORMAT_BGR24, PIXELFORMAT_BGR888:
		col := surface.ColorModel().Convert(c).(color.RGBA)
		pix[i+2] = col.R
		pix[i+1] = col.G
		pix[i+0] = col.B
	default:
		panic("Unknown pixel format!")
	}
}

// SetAlphaMod sets an additional alpha value used in blit operations.
// (https://wiki.libsdl.org/SDL_SetSurfaceAlphaMod)
func (surface *Surface) SetAlphaMod(alpha uint8) error {
	ret, _, _ := setSurfaceAlphaMod.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(alpha),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// SetBlendMode sets the blend mode used for blit operations.
// (https://wiki.libsdl.org/SDL_SetSurfaceBlendMode)
func (surface *Surface) SetBlendMode(bm BlendMode) error {
	ret, _, _ := setSurfaceBlendMode.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(bm),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// SetClipRect sets the clipping rectangle for the surface
// (https://wiki.libsdl.org/SDL_SetClipRect)
func (surface *Surface) SetClipRect(rect *Rect) bool {
	ret, _, _ := setClipRect.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(rect)),
	)
	return ret != 0
}

// SetColorKey sets the color key (transparent pixel) in the surface.
// (https://wiki.libsdl.org/SDL_SetColorKey)
func (surface *Surface) SetColorKey(flag bool, key uint32) error {
	ret, _, _ := setColorKey.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(Btoi(flag)),
		uintptr(key),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// SetColorMod sets an additional color value multiplied into blit operations.
// (https://wiki.libsdl.org/SDL_SetSurfaceColorMod)
func (surface *Surface) SetColorMod(r, g, b uint8) error {
	ret, _, _ := setSurfaceColorMod.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(r),
		uintptr(g),
		uintptr(b),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// SetPalette sets the palette used by the surface.
// (https://wiki.libsdl.org/SDL_SetSurfacePalette)
func (surface *Surface) SetPalette(palette *Palette) error {
	ret, _, _ := setSurfacePalette.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(palette)),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// SetRLE sets the RLE acceleration hint for the surface.
// (https://wiki.libsdl.org/SDL_SetSurfaceRLE)
func (surface *Surface) SetRLE(flag bool) error {
	ret, _, _ := setSurfaceRLE.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(Btoi(flag)),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// SoftStretch has been replaced by BlitScaled().
// (https://wiki.libsdl.org/SDL_SoftStretch)
func (surface *Surface) SoftStretch(srcRect *Rect, dst *Surface, dstRect *Rect) error {
	ret, _, _ := softStretch.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(srcRect)),
		uintptr(unsafe.Pointer(dst)),
		uintptr(unsafe.Pointer(dstRect)),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// Unlock releases the surface after directly accessing the pixels.
// (https://wiki.libsdl.org/SDL_UnlockSurface)
func (surface *Surface) Unlock() {
	unlockSurface.Call(uintptr(unsafe.Pointer(surface)))
}

// UpperBlit has been replaced by Blit().
// (https://wiki.libsdl.org/SDL_UpperBlit)
func (surface *Surface) UpperBlit(srcRect *Rect, dst *Surface, dstRect *Rect) error {
	ret, _, _ := upperBlit.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(srcRect)),
		uintptr(unsafe.Pointer(dst)),
		uintptr(unsafe.Pointer(dstRect)),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// UpperBlitScaled has been replaced by BlitScaled().
// (https://wiki.libsdl.org/SDL_UpperBlitScaled)
func (surface *Surface) UpperBlitScaled(srcRect *Rect, dst *Surface, dstRect *Rect) error {
	ret, _, _ := upperBlitScaled.Call(
		uintptr(unsafe.Pointer(surface)),
		uintptr(unsafe.Pointer(srcRect)),
		uintptr(unsafe.Pointer(dst)),
		uintptr(unsafe.Pointer(dstRect)),
	)
	if ret != 0 {
		return GetError()
	}
	return nil
}

// SysWMEvent contains a video driver dependent system event.
// (https://wiki.libsdl.org/SDL_SysWMEvent)
type SysWMEvent struct {
	Type      uint32         // SYSWMEVENT
	Timestamp uint32         // timestamp of the event
	msg       unsafe.Pointer // driver dependent data, defined in SDL_syswm.h
}

// GetTimestamp returns the timestamp of the event.
func (e *SysWMEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *SysWMEvent) GetType() uint32 {
	return e.Type
}

// SysWMInfo contains system-dependent information about a window.
// (https://wiki.libsdl.org/SDL_SysWMinfo)
type SysWMInfo struct {
	Version   Version  // a Version structure that contains the current SDL version
	Subsystem uint32   // the windowing system type
	dummy     [24]byte // unused (to help compilers when no specific system is available)
}

// GetCocoaInfo returns Apple Mac OS X window information.
func (info *SysWMInfo) GetCocoaInfo() *CocoaInfo {
	return (*CocoaInfo)(unsafe.Pointer(&info.dummy[0]))
}

// GetDFBInfo returns DirectFB window information.
func (info *SysWMInfo) GetDFBInfo() *DFBInfo {
	return (*DFBInfo)(unsafe.Pointer(&info.dummy[0]))
}

// GetUIKitInfo returns Apple iOS window information.
func (info *SysWMInfo) GetUIKitInfo() *UIKitInfo {
	return (*UIKitInfo)(unsafe.Pointer(&info.dummy[0]))
}

// GetWindowsInfo returns Microsoft Windows window information.
func (info *SysWMInfo) GetWindowsInfo() *WindowsInfo {
	return (*WindowsInfo)(unsafe.Pointer(&info.dummy[0]))
}

// GetX11Info returns X Window System window information.
func (info *SysWMInfo) GetX11Info() *X11Info {
	return (*X11Info)(unsafe.Pointer(&info.dummy[0]))
}

// SystemCursor is a system cursor created by CreateSystemCursor().
type SystemCursor uint32

// TextEditingEvent contains keyboard text editing event information.
// (https://wiki.libsdl.org/SDL_TextEditingEvent)
type TextEditingEvent struct {
	Type      uint32                         // TEXTEDITING
	Timestamp uint32                         // timestamp of the event
	WindowID  uint32                         // the window with keyboard focus, if any
	Text      [TEXTINPUTEVENT_TEXT_SIZE]byte // the null-terminated editing text in UTF-8 encoding
	Start     int32                          // the location to begin editing from
	Length    int32                          // the number of characters to edit from the start point
}

// GetTimestamp returns the timestamp of the event.
func (e *TextEditingEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *TextEditingEvent) GetType() uint32 {
	return e.Type
}

// TextInputEvent contains keyboard text input event information.
// (https://wiki.libsdl.org/SDL_TextInputEvent)
type TextInputEvent struct {
	Type      uint32                         // TEXTINPUT
	Timestamp uint32                         // timestamp of the event
	WindowID  uint32                         // the window with keyboard focus, if any
	Text      [TEXTINPUTEVENT_TEXT_SIZE]byte // the null-terminated input text in UTF-8 encoding
}

// GetTimestamp returns the timestamp of the event.
func (e *TextInputEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *TextInputEvent) GetType() uint32 {
	return e.Type
}

// Texture contains an efficient, driver-specific representation of pixel data.
// (https://wiki.libsdl.org/SDL_Texture)
type Texture struct{}

// Destroy destroys the specified texture.
// (https://wiki.libsdl.org/SDL_DestroyTexture)
func (texture *Texture) Destroy() error {
	lastErr := GetError()
	ClearError()
	destroyTexture.Call(uintptr(unsafe.Pointer(texture)))
	err := GetError()
	if err != nil {
		return err
	}
	SetError(lastErr)
	return nil
}

// GLBind binds an OpenGL/ES/ES2 texture to the current context for use with OpenGL instructions when rendering OpenGL primitives directly.
// (https://wiki.libsdl.org/SDL_GL_BindTexture)
func (texture *Texture) GLBind(texw, texh *float32) error {
	ret, _, _ := gl_BindTexture.Call(
		uintptr(unsafe.Pointer(texture)),
		uintptr(unsafe.Pointer(texw)),
		uintptr(unsafe.Pointer(texh)),
	)
	return errorFromInt(int(ret))
}

// GLUnbind unbinds an OpenGL/ES/ES2 texture from the current context.
// (https://wiki.libsdl.org/SDL_GL_UnbindTexture)
func (texture *Texture) GLUnbind() error {
	ret, _, _ := gl_UnbindTexture.Call(uintptr(unsafe.Pointer(texture)))
	return errorFromInt(int(ret))
}

// GetAlphaMod returns the additional alpha value multiplied into render copy operations.
// (https://wiki.libsdl.org/SDL_GetTextureAlphaMod)
func (texture *Texture) GetAlphaMod() (alpha uint8, err error) {
	ret, _, _ := getTextureAlphaMod.Call(
		uintptr(unsafe.Pointer(texture)),
		uintptr(unsafe.Pointer(&alpha)),
	)
	err = errorFromInt(int(ret))
	return
}

// GetBlendMode returns the blend mode used for texture copy operations.
// (https://wiki.libsdl.org/SDL_GetTextureBlendMode)
func (texture *Texture) GetBlendMode() (bm BlendMode, err error) {
	ret, _, _ := getTextureBlendMode.Call(
		uintptr(unsafe.Pointer(texture)),
		uintptr(unsafe.Pointer(&bm)),
	)
	err = errorFromInt(int(ret))
	return
}

// Lock locks a portion of the texture for write-only pixel access.
// (https://wiki.libsdl.org/SDL_LockTexture)
func (texture *Texture) Lock(rect *Rect) ([]byte, int, error) {
	var pitch int
	var pixels unsafe.Pointer
	ret, _, _ := lockTexture.Call(
		uintptr(unsafe.Pointer(texture)),
		uintptr(unsafe.Pointer(rect)),
		uintptr(unsafe.Pointer(&pixels)),
		uintptr(unsafe.Pointer(&pitch)),
	)
	if ret != 0 {
		return nil, pitch, GetError()
	}

	_, _, w, h, err := texture.Query()
	if err != nil {
		return nil, pitch, GetError()
	}

	var b []byte
	var length int
	if rect != nil {
		bytesPerPixel := int32(pitch) / w
		length = int(bytesPerPixel * (w*rect.H - rect.X - (w - rect.X - rect.W)))
	} else {
		length = pitch * int(h)
	}
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sliceHeader.Cap = int(length)
	sliceHeader.Len = int(length)
	sliceHeader.Data = uintptr(pixels)

	return b, pitch, nil
}

// Query returns the attributes of a texture.
// (https://wiki.libsdl.org/SDL_QueryTexture)
func (texture *Texture) Query() (format uint32, access int, width, height int32, err error) {
	ret, _, _ := queryTexture.Call(
		uintptr(unsafe.Pointer(texture)),
		uintptr(unsafe.Pointer(&format)),
		uintptr(unsafe.Pointer(&access)),
		uintptr(unsafe.Pointer(&width)),
		uintptr(unsafe.Pointer(&height)),
	)
	if ret != 0 {
		return 0, 0, 0, 0, GetError()
	}
	return
}

// SetAlphaMod sets an additional alpha value multiplied into render copy operations.
// (https://wiki.libsdl.org/SDL_SetTextureAlphaMod)
func (texture *Texture) SetAlphaMod(alpha uint8) error {
	ret, _, _ := setTextureAlphaMod.Call(
		uintptr(unsafe.Pointer(texture)),
		uintptr(alpha),
	)
	return errorFromInt(int(ret))
}

// SetBlendMode sets the blend mode for a texture, used by Renderer.Copy().
// (https://wiki.libsdl.org/SDL_SetTextureBlendMode)
func (texture *Texture) SetBlendMode(bm BlendMode) error {
	ret, _, _ := setTextureBlendMode.Call(
		uintptr(unsafe.Pointer(texture)),
		uintptr(bm),
	)
	return errorFromInt(int(ret))
}

// SetColorMod sets an additional color value multiplied into render copy operations.
// (https://wiki.libsdl.org/SDL_SetTextureColorMod)
func (texture *Texture) SetColorMod(r uint8, g uint8, b uint8) error {
	ret, _, _ := setTextureColorMod.Call(
		uintptr(unsafe.Pointer(texture)),
		uintptr(r),
		uintptr(g),
		uintptr(b),
	)
	return errorFromInt(int(ret))
}

// Unlock unlocks a texture, uploading the changes to video memory, if needed.
// (https://wiki.libsdl.org/SDL_UnlockTexture)
func (texture *Texture) Unlock() {
	unlockTexture.Call(uintptr(unsafe.Pointer(texture)))
}

// Update updates the given texture rectangle with new pixel data.
// (https://wiki.libsdl.org/SDL_UpdateTexture)
func (texture *Texture) Update(rect *Rect, pixels []byte, pitch int) error {
	if pixels == nil {
		return nil // TODO this should be in the original as well
	}
	ret, _, _ := updateTexture.Call(
		uintptr(unsafe.Pointer(texture)),
		uintptr(unsafe.Pointer(rect)),
		uintptr(unsafe.Pointer(&pixels[0])),
		uintptr(pitch),
	)
	return errorFromInt(int(ret))
}

// UpdateRGBA updates the given texture rectangle with new uint32 pixel data.
// (https://wiki.libsdl.org/SDL_UpdateTexture)
func (texture *Texture) UpdateRGBA(rect *Rect, pixels []uint32, pitch int) error {
	if pixels == nil {
		return nil // TODO this should be in the original as well
	}
	ret, _, _ := updateTexture.Call(
		uintptr(unsafe.Pointer(texture)),
		uintptr(unsafe.Pointer(rect)),
		uintptr(unsafe.Pointer(&pixels[0])),
		uintptr(4*pitch), // 4 bytes in one uint32
	)
	return errorFromInt(int(ret))
}

// UpdateYUV updates a rectangle within a planar YV12 or IYUV texture with new pixel data.
// (https://wiki.libsdl.org/SDL_UpdateYUVTexture)
func (texture *Texture) UpdateYUV(rect *Rect, yPlane []byte, yPitch int, uPlane []byte, uPitch int, vPlane []byte, vPitch int) error {
	// TODO handle nil []bytes in the original as well
	var yPlanePtr, uPlanePtr, vPlanePtr uintptr
	if yPlane != nil {
		yPlanePtr = uintptr(unsafe.Pointer(&yPlane[0]))
	}
	if uPlane != nil {
		uPlanePtr = uintptr(unsafe.Pointer(&uPlane[0]))
	}
	if vPlane != nil {
		vPlanePtr = uintptr(unsafe.Pointer(&vPlane[0]))
	}
	ret, _, _ := updateYUVTexture.Call(
		uintptr(unsafe.Pointer(texture)),
		uintptr(unsafe.Pointer(rect)),
		yPlanePtr,
		uintptr(yPitch),
		uPlanePtr,
		uintptr(uPitch),
		vPlanePtr,
		uintptr(vPitch),
	)
	return errorFromInt(int(ret))
}

// ThreadID is the thread identifier for a thread.
type ThreadID uint64

// TouchFingerEvent contains finger touch event information.
// (https://wiki.libsdl.org/SDL_TouchFingerEvent)
type TouchFingerEvent struct {
	Type      uint32   // FINGERMOTION, FINGERDOWN, FINGERUP
	Timestamp uint32   // timestamp of the event
	TouchID   TouchID  // the touch device id
	FingerID  FingerID // the finger id
	X         float32  // the x-axis location of the touch event, normalized (0...1)
	Y         float32  // the y-axis location of the touch event, normalized (0...1)
	DX        float32  // the distance moved in the x-axis, normalized (-1...1)
	DY        float32  // the distance moved in the y-axis, normalized (-1...1)
	Pressure  float32  // the quantity of pressure applied, normalized (0...1)
}

// GetTimestamp returns the timestamp of the event.
func (e *TouchFingerEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *TouchFingerEvent) GetType() uint32 {
	return e.Type
}

// TouchID is the ID of a touch device.
type TouchID int64

// GetTouchDevice returns the touch ID with the given index.
// (https://wiki.libsdl.org/SDL_GetTouchDevice)
func GetTouchDevice(index int) TouchID {
	ret, _, _ := getTouchDevice.Call(uintptr(index))
	return TouchID(ret)
}

// UIKitInfo contains Apple iOS window information.
type UIKitInfo struct {
	Window unsafe.Pointer // the UIKit window
}

// UserEvent contains an application-defined event type.
// (https://wiki.libsdl.org/SDL_UserEvent)
type UserEvent struct {
	Type      uint32         // value obtained from RegisterEvents()
	Timestamp uint32         // timestamp of the event
	WindowID  uint32         // the associated window, if any
	Code      int32          // user defined event code
	Data1     unsafe.Pointer // user defined data pointer
	Data2     unsafe.Pointer // user defined data pointer
}

// GetTimestamp returns the timestamp of the event.
func (e *UserEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *UserEvent) GetType() uint32 {
	return e.Type
}

// Version contains information about the version of SDL in use.
// (https://wiki.libsdl.org/SDL_version)
type Version struct {
	Major uint8 // major version
	Minor uint8 // minor version
	Patch uint8 // update version (patchlevel)
}

// Window is a type used to identify a window.
type Window struct{}

// CreateWindow creates a window with the specified position, dimensions, and flags.
// (https://wiki.libsdl.org/SDL_CreateWindow)
func CreateWindow(title string, x, y, w, h int32, flags uint32) (*Window, error) {
	t := append([]byte(title), 0)
	ret, _, _ := createWindow.Call(
		uintptr(unsafe.Pointer(&t[0])),
		uintptr(x),
		uintptr(y),
		uintptr(w),
		uintptr(h),
		uintptr(flags),
	)
	if ret == 0 {
		return nil, GetError()
	}
	return (*Window)(unsafe.Pointer(ret)), nil
}

// CreateWindowFrom creates an SDL window from an existing native window.
// (https://wiki.libsdl.org/SDL_CreateWindowFrom)
func CreateWindowFrom(data unsafe.Pointer) (*Window, error) {
	ret, _, _ := createWindowFrom.Call(uintptr(data))
	if ret == 0 {
		return nil, GetError()
	}
	return (*Window)(unsafe.Pointer(ret)), nil
}

// GetKeyboardFocus returns the window which currently has keyboard focus.
// (https://wiki.libsdl.org/SDL_GetKeyboardFocus)
func GetKeyboardFocus() *Window {
	ret, _, _ := getKeyboardFocus.Call()
	return (*Window)(unsafe.Pointer(ret))
}

// GetMouseFocus returns the window which currently has mouse focus.
// (https://wiki.libsdl.org/SDL_GetMouseFocus)
func GetMouseFocus() *Window {
	ret, _, _ := getMouseFocus.Call()
	return (*Window)(unsafe.Pointer(ret))
}

// GetWindowFromID returns a window from a stored ID.
// (https://wiki.libsdl.org/SDL_GetWindowFromID)
func GetWindowFromID(id uint32) (*Window, error) {
	ret, _, _ := getWindowFromID.Call(uintptr(id))
	if ret == 0 {
		return nil, GetError()
	}
	return (*Window)(unsafe.Pointer(ret)), nil
}

// Destroy destroys the window.
// (https://wiki.libsdl.org/SDL_DestroyWindow)
func (window *Window) Destroy() error {
	lastErr := GetError()
	ClearError()
	destroyWindow.Call(uintptr(unsafe.Pointer(window)))
	err := GetError()
	if err != nil {
		return err
	}
	SetError(lastErr)
	return nil
}

// GLCreateContext creates an OpenGL context for use with an OpenGL window, and make it current.
// (https://wiki.libsdl.org/SDL_GL_CreateContext)
func (window *Window) GLCreateContext() (GLContext, error) {
	ret, _, _ := gl_CreateContext.Call(uintptr(unsafe.Pointer(window)))
	if ret == 0 {
		return 0, GetError()
	}
	return GLContext(ret), nil
}

// GLGetDrawableSize returns the size of a window's underlying drawable in pixels (for use with glViewport).
// (https://wiki.libsdl.org/SDL_GL_GetDrawableSize)
func (window *Window) GLGetDrawableSize() (w, h int32) {
	gl_GetDrawableSize.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(unsafe.Pointer(&w)),
		uintptr(unsafe.Pointer(&h)),
	)
	return
}

// GLMakeCurrent sets up an OpenGL context for rendering into an OpenGL window.
// (https://wiki.libsdl.org/SDL_GL_MakeCurrent)
func (window *Window) GLMakeCurrent(glcontext GLContext) error {
	ret, _, _ := gl_MakeCurrent.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(glcontext),
	)
	return errorFromInt(int(ret))
}

// GLSwap updates a window with OpenGL rendering.
// (https://wiki.libsdl.org/SDL_GL_SwapWindow)
func (window *Window) GLSwap() {
	gl_SwapWindow.Call(uintptr(unsafe.Pointer(window)))
}

// GetBrightness returns the brightness (gamma multiplier) for the display that owns the given window.
// (https://wiki.libsdl.org/SDL_GetWindowBrightness)
func (window *Window) GetBrightness() float32 {
	ret, _, _ := getWindowBrightness.Call(uintptr(unsafe.Pointer(window)))
	return float32(ret)
}

// GetData returns the data pointer associated with the window.
// (https://wiki.libsdl.org/SDL_GetWindowData)
func (window *Window) GetData(name string) unsafe.Pointer {
	n := append([]byte(name), 0)
	ret, _, _ := getWindowData.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(unsafe.Pointer(&n[0])),
	)
	return unsafe.Pointer(ret)
}

// GetDisplayIndex returns the index of the display associated with the window.
// (https://wiki.libsdl.org/SDL_GetWindowDisplayIndex)
func (window *Window) GetDisplayIndex() (int, error) {
	ret, _, _ := getWindowDisplayIndex.Call(uintptr(unsafe.Pointer(window)))
	return int(ret), errorFromInt(int(ret))
}

// GetDisplayMode fills in information about the display mode to use when the window is visible at fullscreen.
// (https://wiki.libsdl.org/SDL_GetWindowDisplayMode)
func (window *Window) GetDisplayMode() (mode DisplayMode, err error) {
	ret, _, _ := getWindowDisplayMode.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(unsafe.Pointer(&mode)),
	)
	err = errorFromInt(int(ret))
	return
}

// GetFlags returns the window flags.
// (https://wiki.libsdl.org/SDL_GetWindowFlags)
func (window *Window) GetFlags() uint32 {
	ret, _, _ := getWindowFlags.Call(uintptr(unsafe.Pointer(window)))
	return uint32(ret)
}

// GetGammaRamp returns the gamma ramp for the display that owns a given window.
// (https://wiki.libsdl.org/SDL_GetWindowGammaRamp)
func (window *Window) GetGammaRamp() (red, green, blue *[256]uint16, err error) {
	ret, _, _ := getWindowGammaRamp.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(unsafe.Pointer(red)),
		uintptr(unsafe.Pointer(green)),
		uintptr(unsafe.Pointer(blue)),
	)
	err = errorFromInt(int(ret))
	return
}

// GetGrab returns the window's input grab mode.
// (https://wiki.libsdl.org/SDL_GetWindowGrab)
func (window *Window) GetGrab() bool {
	ret, _, _ := getWindowGrab.Call(uintptr(unsafe.Pointer(window)))
	return ret != 0
}

// GetID returns the numeric ID of the window, for logging purposes.
//  (https://wiki.libsdl.org/SDL_GetWindowID)
func (window *Window) GetID() (uint32, error) {
	ret, _, _ := getWindowID.Call(uintptr(unsafe.Pointer(window)))
	if ret == 0 {
		return 0, GetError()
	}
	return uint32(ret), nil
}

// GetMaximumSize returns the maximum size of the window's client area.
// (https://wiki.libsdl.org/SDL_GetWindowMaximumSize)
func (window *Window) GetMaximumSize() (w, h int32) {
	getWindowMaximumSize.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(unsafe.Pointer(&w)),
		uintptr(unsafe.Pointer(&h)),
	)
	return
}

// GetMinimumSize returns the minimum size of the window's client area.
// (https://wiki.libsdl.org/SDL_GetWindowMinimumSize)
func (window *Window) GetMinimumSize() (w, h int32) {
	getWindowMinimumSize.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(unsafe.Pointer(&w)),
		uintptr(unsafe.Pointer(&h)),
	)
	return
}

// GetPixelFormat returns the pixel format associated with the window.
// (https://wiki.libsdl.org/SDL_GetWindowPixelFormat)
func (window *Window) GetPixelFormat() (uint32, error) {
	ret, _, _ := getWindowPixelFormat.Call(uintptr(unsafe.Pointer(window)))
	if ret == PIXELFORMAT_UNKNOWN {
		return PIXELFORMAT_UNKNOWN, GetError()
	}
	return uint32(ret), nil
}

// GetPosition returns the position of the window.
// (https://wiki.libsdl.org/SDL_GetWindowPosition)
func (window *Window) GetPosition() (x, y int32) {
	getWindowPosition.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(unsafe.Pointer(&x)),
		uintptr(unsafe.Pointer(&y)),
	)
	return
}

// GetRenderer returns the renderer associated with a window.
// (https://wiki.libsdl.org/SDL_GetRenderer)
func (window *Window) GetRenderer() (*Renderer, error) {
	ret, _, _ := getRenderer.Call(uintptr(unsafe.Pointer(window)))
	if ret == 0 {
		return nil, GetError()
	}
	return (*Renderer)(unsafe.Pointer(ret)), nil
}

// GetSize returns the size of the window's client area.
// (https://wiki.libsdl.org/SDL_GetWindowSize)
func (window *Window) GetSize() (w, h int32) {
	getWindowSize.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(unsafe.Pointer(&w)),
		uintptr(unsafe.Pointer(&h)),
	)
	return
}

// GetSurface returns the SDL surface associated with the window.
// (https://wiki.libsdl.org/SDL_GetWindowSurface)
func (window *Window) GetSurface() (*Surface, error) {
	ret, _, _ := getWindowSurface.Call(uintptr(unsafe.Pointer(window)))
	if ret == 0 {
		return nil, GetError()
	}
	return (*Surface)(unsafe.Pointer(ret)), nil
}

// GetTitle returns the title of the window.
// (https://wiki.libsdl.org/SDL_GetWindowTitle)
func (window *Window) GetTitle() string {
	ret, _, _ := getWindowTitle.Call(uintptr(unsafe.Pointer(window)))
	return sdlToGoString(ret)
}

// GetWMInfo returns driver specific information about a window.
// (https://wiki.libsdl.org/SDL_GetWindowWMInfo)
func (window *Window) GetWMInfo() (*SysWMInfo, error) {
	var info SysWMInfo
	ret, _, _ := getWindowWMInfo.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(unsafe.Pointer(&info)),
	)
	if ret == 0 {
		return nil, GetError()
	}
	return &info, nil
}

// GetWindowOpacity returns the opacity of the window.
// (https://wiki.libsdl.org/SDL_GetWindowOpacity)
func (window *Window) GetWindowOpacity() (opacity float32, err error) {
	ret, _, _ := getWindowOpacity.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(unsafe.Pointer(&opacity)),
	)

	return opacity, errorFromInt(int(ret))
}

// Hide hides the window.
// (https://wiki.libsdl.org/SDL_HideWindow)
func (window *Window) Hide() {
	hideWindow.Call(uintptr(unsafe.Pointer(window)))
}

// Maximize makes the window as large as possible.
// (https://wiki.libsdl.org/SDL_MaximizeWindow)
func (window *Window) Maximize() {
	maximizeWindow.Call(uintptr(unsafe.Pointer(window)))
}

// Minimize minimizes the window to an iconic representation.
// (https://wiki.libsdl.org/SDL_MinimizeWindow)
func (window *Window) Minimize() {
	minimizeWindow.Call(uintptr(unsafe.Pointer(window)))
}

// Raise raises the window above other windows and set the input focus.
// (https://wiki.libsdl.org/SDL_RaiseWindow)
func (window *Window) Raise() {
	raiseWindow.Call(uintptr(unsafe.Pointer(window)))
}

// Restore restores the size and position of a minimized or maximized window.
// (https://wiki.libsdl.org/SDL_RestoreWindow)
func (window *Window) Restore() {
	restoreWindow.Call(uintptr(unsafe.Pointer(window)))
}

// SetBordered sets the border state of the window.
// (https://wiki.libsdl.org/SDL_SetWindowBordered)
func (window *Window) SetBordered(bordered bool) {
	setWindowBordered.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(Btoi(bordered)),
	)
}

// SetBrightness sets the brightness (gamma multiplier) for the display that owns the given window.
// (https://wiki.libsdl.org/SDL_SetWindowBrightness)
func (window *Window) SetBrightness(brightness float32) error {
	ret, _, _ := setWindowBrightness.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(brightness),
	)
	return errorFromInt(int(ret))
}

// SetData associates an arbitrary named pointer with the window.
// (https://wiki.libsdl.org/SDL_SetWindowData)
func (window *Window) SetData(name string, userdata unsafe.Pointer) unsafe.Pointer {
	n := append([]byte(name), 0)
	ret, _, _ := setWindowData.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(unsafe.Pointer(&n[0])),
		uintptr(userdata),
	)
	return unsafe.Pointer(ret)
}

// SetDisplayMode sets the display mode to use when the window is visible at fullscreen.
// (https://wiki.libsdl.org/SDL_SetWindowDisplayMode)
func (window *Window) SetDisplayMode(mode *DisplayMode) error {
	ret, _, _ := setWindowDisplayMode.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(unsafe.Pointer(mode)),
	)
	return errorFromInt(int(ret))
}

// SetFullscreen sets the window's fullscreen state.
// (https://wiki.libsdl.org/SDL_SetWindowFullscreen)
func (window *Window) SetFullscreen(flags uint32) error {
	ret, _, _ := setWindowFullscreen.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(flags),
	)
	return errorFromInt(int(ret))
}

// SetGammaRamp sets the gamma ramp for the display that owns the given window.
// (https://wiki.libsdl.org/SDL_SetWindowGammaRamp)
func (window *Window) SetGammaRamp(red, green, blue *[256]uint16) error {
	ret, _, _ := setWindowGammaRamp.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(unsafe.Pointer(red)),
		uintptr(unsafe.Pointer(green)),
		uintptr(unsafe.Pointer(blue)),
	)
	return errorFromInt(int(ret))
}

// SetGrab sets the window's input grab mode.
// (https://wiki.libsdl.org/SDL_SetWindowGrab)
func (window *Window) SetGrab(grabbed bool) {
	setWindowGrab.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(Btoi(grabbed)),
	)
}

// SetIcon sets the icon for the window.
// (https://wiki.libsdl.org/SDL_SetWindowIcon)
func (window *Window) SetIcon(icon *Surface) {
	setWindowIcon.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(unsafe.Pointer(icon)),
	)
}

// SetMaximumSize sets the maximum size of the window's client area.
// (https://wiki.libsdl.org/SDL_SetWindowMaximumSize)
func (window *Window) SetMaximumSize(maxW, maxH int32) {
	setWindowMaximumSize.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(maxW),
		uintptr(maxH),
	)
}

// SetMinimumSize sets the minimum size of the window's client area.
// (https://wiki.libsdl.org/SDL_SetWindowMinimumSize)
func (window *Window) SetMinimumSize(minW, minH int32) {
	setWindowMinimumSize.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(minW),
		uintptr(minH),
	)
}

// SetPosition sets the position of the window.
// (https://wiki.libsdl.org/SDL_SetWindowPosition)
func (window *Window) SetPosition(x, y int32) {
	setWindowPosition.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(x),
		uintptr(y),
	)
}

// SetResizable sets the user-resizable state of the window.
// (https://wiki.libsdl.org/SDL_SetWindowResizable)
func (window *Window) SetResizable(resizable bool) {
	setWindowResizable.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(Btoi(resizable)),
	)
}

// SetSize sets the size of the window's client area.
// (https://wiki.libsdl.org/SDL_SetWindowSize)
func (window *Window) SetSize(w, h int32) {
	setWindowSize.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(w),
		uintptr(h),
	)
}

// SetTitle sets the title of the window.
// (https://wiki.libsdl.org/SDL_SetWindowTitle)
func (window *Window) SetTitle(title string) {
	t := append([]byte(title), 0)
	setWindowTitle.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(unsafe.Pointer(&t[0])),
	)
}

// SetWindowOpacity sets the opacity of the window.
// (https://wiki.libsdl.org/SDL_SetWindowOpacity)
func (window *Window) SetWindowOpacity(opacity float32) error {
	ret, _, _ := setWindowOpacity.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(opacity),
	)
	return errorFromInt(int(ret))
}

// Show shows the window.
// (https://wiki.libsdl.org/SDL_ShowWindow)
func (window *Window) Show() {
	showWindow.Call(uintptr(unsafe.Pointer(window)))
}

// UpdateSurface copies the window surface to the screen.
// (https://wiki.libsdl.org/SDL_UpdateWindowSurface)
func (window *Window) UpdateSurface() error {
	ret, _, _ := updateWindowSurface.Call(uintptr(unsafe.Pointer(window)))
	return errorFromInt(int(ret))
}

// UpdateSurfaceRects copies areas of the window surface to the screen.
// (https://wiki.libsdl.org/SDL_UpdateWindowSurfaceRects)
func (window *Window) UpdateSurfaceRects(rects []Rect) error {
	if rects == nil {
		return nil // TODO this should be in the original as well
	}
	ret, _, _ := updateWindowSurfaceRects.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(unsafe.Pointer(&rects[0])),
		uintptr(len(rects)),
	)
	return errorFromInt(int(ret))
}

// VulkanCreateSurface creates a Vulkan rendering surface for a window.
// (https://wiki.libsdl.org/SDL_Vulkan_CreateSurface)
func (window *Window) VulkanCreateSurface(instance interface{}) (surface uintptr, err error) {
	// TODO
	return 0, nil
	//if instance == nil {
	//	return 0, errors.New("vulkan: instance is nil")
	//}
	//val := reflect.ValueOf(instance)
	//if val.Kind() != reflect.Ptr {
	//	return 0, errors.New("vulkan: instance is not a VkInstance (expected kind Ptr, got " + val.Kind().String() + ")")
	//}
	//var vulkanSurface C.VkSurfaceKHR
	//if C.SDL_Vulkan_CreateSurface(window.cptr(),
	//	(C.VkInstance)(unsafe.Pointer(val.Pointer())),
	//	(*C.VkSurfaceKHR)(unsafe.Pointer(&vulkanSurface))) == C.SDL_FALSE {
	//
	//	return 0, GetError()
	//}
	//return uintptr(unsafe.Pointer(&vulkanSurface)), nil
}

// VulkanGetDrawableSize gets the size of a window's underlying drawable in pixels (for use with setting viewport, scissor & etc).
// (https://wiki.libsdl.org/SDL_Vulkan_GetDrawableSize)
func (window *Window) VulkanGetDrawableSize() (w, h int32) {
	vulkan_GetDrawableSize.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(unsafe.Pointer(&w)),
		uintptr(unsafe.Pointer(&h)),
	)
	return
}

// VulkanGetInstanceExtensions gets the names of the Vulkan instance extensions needed to create a surface with VulkanCreateSurface().
// (https://wiki.libsdl.org/SDL_Vulkan_GetInstanceExtensions)
func (window *Window) VulkanGetInstanceExtensions() []string {
	// TODO
	return nil
	//var count C.uint
	//C.SDL_Vulkan_GetInstanceExtensions(window.cptr(), &count, nil)
	//if count == 0 {
	//	return nil
	//}
	//
	//strptrs := make([]*C.char, uint(count))
	//C.SDL_Vulkan_GetInstanceExtensions(window.cptr(), &count, &strptrs[0])
	//extensions := make([]string, uint(count))
	//for i := range strptrs {
	//	extensions[i] = C.GoString(strptrs[i])
	//}
	//return extensions
}

// WarpMouseInWindow moves the mouse to the given position within the window.
// (https://wiki.libsdl.org/SDL_WarpMouseInWindow)
func (window *Window) WarpMouseInWindow(x, y int32) {
	warpMouseInWindow.Call(
		uintptr(unsafe.Pointer(window)),
		uintptr(x),
		uintptr(y),
	)
}

// WindowEvent contains window state change event data.
// (https://wiki.libsdl.org/SDL_WindowEvent)
type WindowEvent struct {
	Type      uint32 // WINDOWEVENT
	Timestamp uint32 // timestamp of the event
	WindowID  uint32 // the associated window
	Event     uint8  // (https://wiki.libsdl.org/SDL_WindowEventID)
	_         uint8  // padding
	_         uint8  // padding
	_         uint8  // padding
	Data1     int32  // event dependent data
	Data2     int32  // event dependent data
}

// GetTimestamp returns the timestamp of the event.
func (e *WindowEvent) GetTimestamp() uint32 {
	return e.Timestamp
}

// GetType returns the event type.
func (e *WindowEvent) GetType() uint32 {
	return e.Type
}

// WindowsInfo contains Microsoft Windows window information.
type WindowsInfo struct {
	Window unsafe.Pointer // the window handle
}

// X11Info contains X Window System window information.
type X11Info struct {
	Display unsafe.Pointer // the X11 display
	Window  uint           // the X11 window
}

type YUV_CONVERSION_MODE uint32

// GetYUVConversionMode gets the YUV conversion mode
// TODO: (https://wiki.libsdl.org/SDL_GetYUVConversionMode)
func GetYUVConversionMode() YUV_CONVERSION_MODE {
	ret, _, _ := getYUVConversionMode.Call()
	return YUV_CONVERSION_MODE(ret)
}

// GetYUVConversionModeForResolution gets the YUV conversion mode
// TODO: (https://wiki.libsdl.org/SDL_GetYUVConversionModeForResolution)
func GetYUVConversionModeForResolution(width, height int) YUV_CONVERSION_MODE {
	ret, _, _ := getYUVConversionModeForResolution.Call(
		uintptr(width),
		uintptr(height),
	)
	return YUV_CONVERSION_MODE(ret)
}

func sdlToGoString(p uintptr) string {
	if p == 0 {
		return ""
	}
	var buf []byte
	for b := *((*byte)(unsafe.Pointer(p))); b != 0; b = *((*byte)(unsafe.Pointer(p))) {
		buf = append(buf, b)
		p++
	}
	return string(buf)
}
