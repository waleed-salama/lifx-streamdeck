function UpdateSettings() {
    if (websocket) {
        const json = {
            "action": actionInfo.action,
            "event": "setSettings",
            "context": uuid,
            "payload": actionInfo.payload.settings
        };
        websocket.send(JSON.stringify(json));
    }
}

function UpdateGlobalSettings(payload) {
    if (websocket) {
        const json = {
            "action": actionInfo.action,
            "event": "setGlobalSettings",
            "context": uuid,
            "payload": payload
        };
        websocket.send(JSON.stringify(json));
    } else {
        console.error("websocket null")
    }
}

function getUUID() {
    return uuid;
}

function getWS() {
    return websocket;
}

// our method to pass values to the plugin
function sendValueToPlugin(payload) {
    if (websocket) {
        const json = {
            "action": actionInfo.action,
            "event": "sendToPlugin",
            "context": uuid,
            "payload": payload
        };
        websocket.send(JSON.stringify(json));
    }
}

function RemoveDevice(deviceMac) {
    let foundIndex = -1;
    const devices = actionInfo.payload.settings["devices"];
    for (let index = 0; index < devices.length; index++) {
        const device = devices[index];
        if (device["mac"] === deviceMac) {
            foundIndex = index;
        }
    }
    if (foundIndex >= 0) {
        actionInfo.payload.settings["devices"].splice(foundIndex, 1)
    } else {
        console.log("No device found with mac ", deviceMac)
    }
    UpdateSettings();
}

function AddDevice(deviceMac, deviceLabel) {
    device = {"mac": deviceMac, "label": deviceLabel};
    actionInfo.payload.settings["devices"].push(device);
    UpdateSettings();
}


function DiscoverDevices() {
    sendValueToPlugin({ 'type': 'discovery' });
    let container = document.getElementById('discover-container');
    let discoveryButton = document.getElementById('discovery');
    let loader = document.createElement('div');
    loader.setAttribute('class', 'loader');
    discoveryButton.setAttribute('disabled', 'disabled');
    container.append(loader);
}
