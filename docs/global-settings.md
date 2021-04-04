# Outbound IP

Most users don't need to set the Outbound IP address, and can leave it blank which auto-discovers the IP address to use.  If you are using a VPN or have multiple network adapters/networks (for example a work network and personal network) on your computer then the auto-discovery may choose the wrong network adapter.

## Getting your LIFX Devices IP Address
Many routers offer the ability to get the device list, it might be under a page like "attached devices" or "connected devices" for example below is my LIFX Beam named Beam on the page with the IP Address of `192.168.2.102`

![img.png](imgs/lifx-beam-ip.png)

## Windows
On Windows open command prompt by hitting start and typing in `cmd` once that loads, run `ipconfig`

![img.png](imgs/ipconfig.png)

Based on the screenshot above I would enter `192.168.2.124` into the Outbound IP box as the first 3 numbers match with my LIFX Bulb, which would cause it to use this network adapter to manage the devices.
## Mac
On mac open your `terminal` application then run the `ifconfig` command

![img.png](imgs/ifconfig.png)

Based on the screenshot above I would enter `192.168.2.122` into the Outbound IP box as the first 3 numbers match with my LIFX Bulb, which would cause it to use this network adapter to manage the devices.

# Direct Communication

This is a setting that usually doesn't need to be changed which makes it so that all requests are sent to your devices directly instead of using broadcasts.  With this enabled if the IP address of your devices change, you will need to perform a rediscovery.

This setting is required when configuring custom devices in a future release.