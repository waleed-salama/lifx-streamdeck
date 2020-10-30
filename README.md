# LIFX Stream Deck
[![pipeline status](https://gitlab.com/wwsean08/lifx-streamdeck/badges/main/pipeline.svg)](https://gitlab.com/wwsean08/lifx-streamdeck/-/commits/main)
[![discord](https://img.shields.io/discord/493162062524973056?label=discord)](https://discord.gg/PPVYMeP)
[![license](https://img.shields.io/badge/license-MIT-brightgreen)](https://gitlab.com/wwsean08/lifx-streamdeck/-/blob/master/LICENSE)

This plugin allows for an Elgato Stream Deck to control LIFX devices at the click of a button.  Currently, it is capable of turning devices on or off, as well as setting the color of the light.

## Contributions
All contributions are welcome, whether it's creating/updating docs, creating issues, requesting new features, or opening pull requests.  Feedback is welcome both via GitLab issues or via discord in the lifx-streamdeck-feedback channel. 

## Installation and Upgrading
### Installing
Installing the application is as simple as downloading the latest version of the application from the [releases page](https://gitlab.com/wwsean08/lifx-streamdeck/-/releases) and then running it and accepting the warning:

![Example of the warning](screenshots/warning.png)

### Upgrading
Upgrading requires uninstalling the previous version, don't worry, your settings will be persisted by the Stream Deck.  To uninstall it, go to "More Actions..." in the Stream Deck application, and click Uninstall:

![Example uninstall](screenshots/uninstall.png)

Once you have uninstalled the plugin you can follow the Installation instructions.

## Supported Actions
### Setting the devices power state
You are able to turn devices on or off using the respective actions.  Doing this can control more than one device at once

### Setting the devices color
You are able to set a device or multiple devices color at the push of a button.  In order to make configuration of the color simpler, you are able to choose a device and copy its current settings and store them as the settings for that action.  You can also leverage this in a multi-action to turn on a device, set it's color, or maybe eve change between a few colors every couple seconds for a short period of time.

### Setting the brightness independent of color
You can change the devices brightness independent of it's current color (as brightness is one of the arguments of setting a color).

### Coming Soon: Setting a "wave" effect


## Known Issues
1. Sometimes the device mac address is shown instead of the devices label, if the devices label isn't returned in time then the mac address will be shown as opposed to nothing.