# Agent Notes

- For coding work in this repo, follow the local `karpathy-guidelines`: make scoped changes, surface assumptions, and verify with concrete commands.
- The installed Stream Deck plugin bundle for local testing is:
  `/Users/waleedsalama/Library/Application Support/com.elgato.StreamDeck/Plugins/com.waleed-salama.lifx-plus.sdPlugin`.
- To make local action changes available in Stream Deck, build the macOS binary, copy `lifx-streamdeck`, `manifest.json`, `property-inspector/*`, `images/*`, and `LICENSE` into the installed `.sdPlugin` bundle, then restart Stream Deck.
- The `Animated Scene` action is exposed through `manifest.json` and `property-inspector/set-scene-pi.html`; Stream Deck may need a full app restart before the action appears in `LIFX+ Controls`.
- `plutil -lint` is not reliable for this manifest JSON on this macOS setup; use `node -e 'JSON.parse(...)'` to validate the manifest instead.
- Useful validation after feature work:
  `go test ./...`
  `go build ./...`
  `git diff --check`
  parse any new inline property-inspector scripts with Node.
- LIFX commands addressed only by MAC fall back to UDP broadcast when the device IP is unknown. If discovery returns no devices, custom device IPs in global settings are the reliable path for bulbs that do not respond by broadcast.
- Scene animation commands must be sequential per physical device. Sending overlapping color/power commands to the same bulb can drop UDP acknowledgements and leave bulbs partway through an animation.
- In property inspectors, stale saved devices must keep checkbox `value` set to the MAC, not the label. Unquoted label values such as `Top Light` can be saved as `mac: "Top"` and break the backend action.
- This office network can expose LIFX bulbs across multiple active Mac interfaces on the same subnet. If direct unicast fails for a discovered bulb, retrying the same command via MAC-only broadcast is a useful fallback.
