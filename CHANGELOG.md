# v0.2.3 (WIP)
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