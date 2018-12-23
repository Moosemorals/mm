
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

function buildAssetRow(asset) {
    return buildElement("tr", undefined,
        buildElement("td", undefined, buildElement("span", { "data-name-id": asset.location_id })),
        buildElement("td", undefined, buildElement("span", { "data-name-id": asset.type_id })),
        buildElement("td", undefined, asset.quantity)
    )
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
    apiPost(makeApiPath("/latest/characters/" + user.ID + "/assets/names/"), ids)
        .then(json => {
            const names = {}
            for (let i = 0; i < json.length; i += 1) {
                names[json[i].item_id] = json[i].name
            }
            updateNames(names)
        })
}

function getPublicNames(ids) {
    apiPost(makeApiPath("/latest/universe/names"), ids)
        .then(json => {
            const names = {}
            for (let i = 0; i < json.length; i += 1) {
                names[json[i].id] = json[i].name
            }
            updateNames(names)
        })
}

function getAssets(user) {
    apiGet(makeApiPath("/latest/characters/" + user.ID + "/assets/"))
        .then(json => {

            const toLookup = {
                public: {},
                private: {}
            }
            const tbody = buildElement("tbody")

            for (let i = 0; i < json.length; i += 1) {
                const row = json[i]
                if (row.location_type === "other") {
                    if (row.location_flag !== "Hangar") {
                        toLookup.private[row.location_id] = true
                    }
                } else {
                    toLookup.public[row.location_id] = true
                }
                toLookup.public[row.type_id] = true
                tbody.appendChild(buildAssetRow(row))
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

        })
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
