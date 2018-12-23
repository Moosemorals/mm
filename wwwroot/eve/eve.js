
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

function getPrices() {
    return apiGet(makeApiPath("/latest/markets/prices")).then(json => {
        const prices = {}
        for (let i = 0; i < json.length; i += 1) {
            prices[json[i].type_id] = {
                avg: json[i].average_price,
                adj: json[i].adjusted_price
            }
        }

        $$("[data-price-id").forEach(x => {
            const id = x.dataset.priceId
            if (id in prices) {
                let value = 0
                if (x.classList.contains("avg")) {
                    value = prices[id].avg
                } else if (x.classList.contains("adj")) {
                    value = prices[id].adj
                } else if (x.classList.contains("val")) {
                    const count = parseInt(x.dataset.count, 10)
                    value = count * (prices[id].adj || prices[id].avg || 0)
                }

                empty(x).appendChild(textNode(formatPrice(value)))
            }
        })

    })

}

function formatPrice(price) {
    if (typeof price === "number") {
        return price.toLocaleString()
    }
    return "-"
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
            buildElement("td", { rowspan: locCount }, buildElement("span", { "data-name-id": id })),
            buildElement("td", { rowspan: locCount }, types[id].count),
            buildElement("td", { rowspan: locCount, class: "price" }, buildElement("span", { class: "avg", "data-price-id": id })),
            buildElement("td", { rowspan: locCount, class: "price" }, buildElement("span", { class: "adj", "data-price-id": id })),
            buildElement("td", { rowspan: locCount, class: "price" }, buildElement("span", { class: "val", "data-price-id": id, "data-count": types[id].count }))
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
                buildElement("th", { colspan: 2 }, "Price"),
                buildElement("th", undefined, "Value"),
                buildElement("th", undefined, "Location"),
                buildElement("th", undefined, "Count")
            )),
            tbody
        )
    )

    getPublicNames(Object.keys(toLookup.public))
    getPrivateNames(user, Object.keys(toLookup.private).map(x => parseInt(x, 10)))
    getPrices()
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
