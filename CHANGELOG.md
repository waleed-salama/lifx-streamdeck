# vNext (WIP)
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