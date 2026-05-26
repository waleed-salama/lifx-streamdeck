# v0.8.0
## Features
* Add an Animated Scene action with device placement, position-aware presets, reversible exit scenes, per-device timing offsets, target color controls, and customizable timing.
* Add per-device scene colors and a Capture Current Scene button so each selected bulb can keep its own hue, saturation, brightness, and Kelvin target.
* Capture and restore per-device power states so animated scenes can intentionally leave selected bulbs off.
* Add a scene-switch option that starts animation from the bulbs' current state instead of fading in from black.
* Add scene-aware cancellation so normal power, color, brightness, toggle, and waveform actions stop any in-progress scene before sending new commands.
* Add fallback handling for bulbs that respond more reliably through broadcast than direct unicast.
* Improve Get Settings failure handling so the property inspector stops loading and reports lookup errors instead of spinning indefinitely.
* Keep the plugin-owned crash log bounded by rotating it when it grows beyond 1 MiB.

## Fixes
* Preserve selected device MAC addresses in property inspectors so saved settings do not accidentally store labels such as `Top Light` as invalid MAC values.

# v0.7.2
## Fixes 
* Generate crash log for mac as well if there is a "panic" to help debugging in the future potentially.
* Make help links open to the docs that the specific version of the plugin installed (or the main branch of develop)
* Some minor fixes based on static code analysis.

# v0.7.1
## Features
* Update the debug action to include more log output.

## Fixes
* Fixed a bug around auto-rediscovery where it might not happen.
  * Found via new unit tests

# v0.7.0
## Fixes
* Remove extra debug logs that were introduced on accident.
* Remove viper to fix CVE-2019-11254 as viper was only being used for debugging purposes
* Fixed some migration issues for settings

## Features
* Add a global setting to not use broadcasts for all light communications via the `Direct Connection` checkbox.
* Automatically rediscover devices on network every hour.
* Add the ability to add custom IP addresses of devices for special network configurations.

# v0.6.2
## Fixes
* Adds some new debug information

# v0.6.1
## Fixes
* Added a link to documentation around the `Outbound IP` setting to hopefully reduce confusion.

# Unrelated Changes
* Upgraded to go 1.16

# v0.6.0
## Features
* Create a cache to reduce the occurrences of [issue 1](https://gitlab.com/wwsean08/lifx-streamdeck/-/issues/1) where mac addresses are shown instead of the device label. 

## Unrelated Changes
* Added links to my kofi, discord, and twitch into the global settings which will open in your default browser.

# v0.5.2
## Fixes
* Fixes [issue #10](https://gitlab.com/wwsean08/lifx-streamdeck/-/issues/10) by allowing you to enter the outbound IP to use.  This is available via the global settings option.  Future versions might be slightly nicer for input.

# v0.5.1
## Fixes
* The order of devices is now sorted alphabetically for discovered devices.  This will result in more consistent locations for device selection.
* Fix the debug action for MacOS.

# v0.5.0 
## New Features
* Add a debug action to allow me to better assist users running into issues.

# v0.4.1
## New Features
* Upgrade compiler to go 1.15 reducing binary sizes to about half their existing size.

## Fixes
* Fix a bug where devices would be duplicated in dropdowns on rediscovery.

# v0.4.0
## New Features
* Add a toggle button to allow toggling device(s) on/off at a click of a button.
  * This will toggle each individual light and does not keep a state, so if some are on, and some are off, each will turn on or off depending on their current power state.
* Allow you to specify the transition time for on/off between 0 (default) and 15 seconds (in milliseconds) like setting the brightness, and color.

## Fixes
* Performance improvement which should keep lights more in sync when pressing a button that affects multiple lights.

# v0.3.0
## New Features
* Add support for LIFX Waveform functionality.

## Fixes
* Update golifx library to support more devices.
* Add tooltip for set brightness action.
* Tech debt cleanup type work.

# v0.2.3
## New Features
* Setting the color or brightness of your lights now supports a transition time.  Existing configs will be set to 0 so it acts as it did before of changing instantly.  The max transition time is set to 15 seconds.
* Automate the same virus scan that Elgato runs on every build to prevent submitting a version that will be rejected.

## Fixes
* No longer "panics" during a normal shutdown/closing of the Stream Deck application

# v0.2.2
## Fixes
* Fix (probably) false positive virus detection for the v0.2.1 release by downgrading to go v1.14 

## Known Issues:
* [Sometimes mac addresses are shown](https://gitlab.com/wwsean08/lifx-streamdeck/-/issues/1)

# v0.2.1
## Fixes
* Properly handles websocket errors

## Known Issues:
* [Sometimes mac addresses are shown](https://gitlab.com/wwsean08/lifx-streamdeck/-/issues/1)

# v0.2.0
## New Features
* [Issue #4](https://gitlab.com/wwsean08/lifx-streamdeck/-/issues/4): Add an action to set the brightness independent of the color the light currently is.  There may be a slight delay in this as it requires looking up the current color due to LIFX combining those pieces of data in their API.

## Fixes
* Add licensing and attribution to the Stream Deck package during the build for compliance and legal reasons.

## Unrelated Changes
* New issue and merge request templates created and added to the repository.

## Known Issues:
* [Sometimes mac addresses are shown](https://gitlab.com/wwsean08/lifx-streamdeck/-/issues/1)

# v0.1.1
## Fixes
* Fix high CPU Usage reported by Luke Last - [issue 2](https://gitlab.com/wwsean08/lifx-streamdeck/-/issues/2)
* Handle IP address change of devices by using broadcasts instead of direct messages - [issue 3](https://gitlab.com/wwsean08/lifx-streamdeck/-/issues/3)

## Known Issues:
* [Sometimes mac addresses are shown](https://gitlab.com/wwsean08/lifx-streamdeck/-/issues/1)

# v0.1.0
* Initial Release

## Known Issues:
* [Sometimes mac addresses are shown](https://gitlab.com/wwsean08/lifx-streamdeck/-/issues/1)
