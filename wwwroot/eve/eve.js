
function empty(el) {
    while (el.firstChild) {
        el.removeChild(el.firstChild)
    }
    return el
}

function textNode(text) {
    return document.createTextNode(text)
}

function appendChildren(el, ...content) {

    for (let i = 0; i < content.length; i += 1) {
        const x = content[i]

        switch (typeof x) {
            case "string":
            case "number":
            case "boolean":
                el.appendChild(textNode(x))
                break;
            case "object":
                if (x instanceof HTMLElement) {
                    el.appendChild(x)
                } else {
                    x.appendChild(textNode(JSON.stringify(x)))
                }
                break;
        }
    }

    return el
}

function buildElement(tag, options, ...content) {
    const el = document.createElement(tag)

    if (typeof options === "string") {
        el.setAtttribute("class", options)
    } else if (typeof options === "object") {
        for (let key in options) {
            if (options.hasOwnProperty(key) && options[key] !== undefined) {
                el.setAttribute(key, options[key])
            }
        }
    }

    return appendChildren(el, ...content)
}

function $(selector) {
    return document.querySelector(selector)
}

function $$(selector) {
    return Array.from(document.querySelectorAll(selector))
}

function hide(el) {
    el.classList.add("hidden")
}

function show(el) {
    el.classList.remove("hidden")
}

function apiPost(target, body) {
    return fetch(target + "&m=POST", {
        method: "POST",
        credentials: "same-origin",
        redirect: "follow",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(body)
    })
        .then(r => r.json())
}

function apiGet(target) {
    return fetch(target, {
        credentials: "same-origin",
        redirect: "follow"
    })
        .then(r => r.json())

}

function makeApiPath(path) {
    return "/eveapi/api?p=" + encodeURIComponent(path)
}



function updateNames(names) {
    $$("[data-name-id]").forEach(x => {
        const id = x.dataset.nameId

        if (id in names) {
            empty(x).appendChild(textNode(names[id]))
        }

    })
}

function getPrivateNames(user, ids) {
    apiPost(makeApiPath("/latest/characters/" + user.ID + "/assets/names/"), ids).then(json => {
        const names = {}
        for (let i = 0; i < json.length; i += 1) {
            names[json[i].item_id] = json[i].name
        }
        updateNames(names)
    })
}

function getPublicNames(ids) {
    apiPost(makeApiPath("/latest/universe/names"), ids).then(json => {
        const names = {}
        for (let i = 0; i < json.length; i += 1) {
            names[json[i].id] = json[i].name
        }
        updateNames(names)
    })
}

function showAssetsRaw(user, assets) {
    function _buildRow(asset) {
        return buildElement("tr", undefined,
            buildElement("td", undefined, buildElement("span", { "data-name-id": asset.location_id })),
            buildElement("td", undefined, buildElement("span", { "data-name-id": asset.type_id })),
            buildElement("td", undefined, asset.quantity)
        )
    }
    const toLookup = {
        public: {},
        private: {}
    }
    const tbody = buildElement("tbody")

    for (let i = 0; i < assets.length; i += 1) {
        const row = assets[i]
        if (row.location_type === "other") {
            if (row.location_flag !== "Hangar") {
                toLookup.private[row.location_id] = true
            }
        } else {
            toLookup.public[row.location_id] = true
        }
        toLookup.public[row.type_id] = true
        tbody.appendChild(_buildRow(row))
    }

    appendChildren($("#holder"), buildElement("table", undefined,
        buildElement("thead", undefined, buildElement("tr", undefined,
            buildElement("th", undefined, "Location"),
            buildElement("th", undefined, "Type"),
            buildElement("th", undefined, "Quantity")
        )),
        tbody
    ))

    getPublicNames(Object.keys(toLookup.public))
    getPrivateNames(user, Object.keys(toLookup.private).map(x => parseInt(x, 10)))

}

function showAssetsByType(user, assets) {
    const types = {}
    const toLookup = {
        public: {},
        private: {}
    }

    for (let i = 0; i < assets.length; i += 1) {
        const row = assets[i]
        const t = row.type_id
        const l = row.location_id
        const q = row.quantity

        if (!(t in types)) {
            types[t] = {
                count: 0,
                locations: {}
            }
        }

        types[t].count += q
        if (!(l in types[t].locations)) {
            types[t].locations[l] = 0
        }
        types[t].locations[l] += q

        if (row.location_type === "other") {
            if (row.location_flag !== "Hangar") {
                toLookup.private[row.location_id] = true
            }
        } else {
            toLookup.public[row.location_id] = true
        }
        toLookup.public[t] = true
    }

    const tbody = buildElement("tbody")

    for (let id in types) {
        const locCount = Object.keys(types[id].locations).length
        const firstRow = buildElement("tr", undefined,
            buildElement("td", { rowspan: locCount  }, buildElement("span", { "data-name-id": id })),
            buildElement("td", { rowspan: locCount  }, types[id].count)
        )
        let first = true
        for (let loc in types[id].locations) {
            if (first) {
                appendChildren(firstRow,
                    buildElement("td", undefined, buildElement("span", { "data-name-id": loc })),
                    buildElement("td", undefined, types[id].locations[loc])
                )
                tbody.appendChild(firstRow)
                first = false
            } else {
                tbody.appendChild(
                    buildElement("tr", undefined,
                        buildElement("td", undefined, buildElement("span", { "data-name-id": loc })),
                        buildElement("td", undefined, types[id].locations[loc])
                    )
                )
            }
        }
    }

    $("#holder").appendChild(
        buildElement("table", undefined,
            buildElement("thead", undefined, buildElement("tr", undefined,
                buildElement("th", undefined, "Type"),
                buildElement("th", undefined, "Total"),
                buildElement("th", undefined, "Location"),
                buildElement("th", undefined, "Count")
            )),
            tbody
        )
    )

    getPublicNames(Object.keys(toLookup.public))
    getPrivateNames(user, Object.keys(toLookup.private).map(x => parseInt(x, 10)))
}

function getAssets(user) {
    apiGet(makeApiPath("/latest/characters/" + user.ID + "/assets/")).then(json => showAssetsByType(user, json))
}

function init() {
    apiGet("/eveapi/").then(json => {
        if ("Name" in json) {
            hide($("#register"))
            show($("#holder"))
            appendChildren(empty($("#holder")), "Hello ", json.Name)
            getAssets(json)
        } else if ("AuthURL" in json) {
            $("#authURL").setAttribute("href", json.AuthURL)
        }
    })
}

window.addEventListener("DOMContentLoaded", init)
