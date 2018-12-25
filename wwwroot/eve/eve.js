"use strict";

window.Handler = (function () {

    const handlers = {}

    function _handler(type) {
        return e => {
            for (let s in handlers[type]) {
                const target = e.target.closest(s)
                if (target !== null) {
                    handlers[type][s].apply(target, e)
                }
            }
        }
    }

    function _add(type, selector, fn) {
        if (!(type in handlers)) {
            handlers[type] = {}
            document.addEventListener(type, _handler(type))
        }
        handlers[type][selector] = fn
    }

    return {
        on: _add
    }

})()

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
        el.setAttribute("class", options)
    } else if (typeof options === "object") {
        for (let key in options) {
            if (options.hasOwnProperty(key) && options[key] !== undefined) {
                el.setAttribute(key, options[key])
            }
        }
    }

    return appendChildren(el, ...content)
}

function $(a, b) {
    if (b === undefined) {
        b = a
        a = document
    }
    return a.querySelector(b)
}

function $$(a, b) {
    if (b === undefined) {
        b = a
        a = document
    }
    return Array.from(a.querySelectorAll(b))
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

function staticGet(target) {
    return fetch(target).then(r => r.json())
}

function staticPost(target, body) {
    return fetch(target, {
        method: "POST",
        body: body,
        headers: {
            "Content-Type": "application/x-www-form-urlencoded"
        }
    }).then(r => r.json())
}

function makeParam(param) {
    let result = ""

    for (let key in param) {
        if (param.hasOwnProperty(key)) {
            if (result !== "") {
                result += "&"
            }
            result += encodeURIComponent(key)
            if (param[key] !== undefined) {
                result += "=" + encodeURIComponent(param[key])
            }
        }
    }

    return result
}

function makeApiPath(path) {
    return "/eveapi/api?p=" + encodeURIComponent(path)
}

function makeStaticPath(path) {
    return "/eveapi/static" + path
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
    return apiPost(makeApiPath("/latest/characters/" + user.ID + "/assets/names/"), ids).then(json => {
        const names = {}
        for (let i = 0; i < json.length; i += 1) {
            names[json[i].item_id] = json[i].name
        }
        updateNames(names)
    })
}

function getTypes(types) {
    return staticPost(makeStaticPath(""), makeParam({
        ids: types.join(",")
    }))
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
        return prices
    })
}

function formatPrice(price) {
    if (typeof price === "number") {
        return price.toLocaleString("en-GB", {
            minimumFractionDigits: 2,
            maximumFractionDigits: 2
        }) + " ISK"
    }
    return "-"
}

function getText(el) {
    if (el !== null && el !== undefined && el.firstChild !== null && el.firstChild.nodeType === Node.TEXT_NODE) {
        return el.firstChild.nodeValue
    }
    return ""
}

window.SortAssets = (function () {

    let dir, last

    function _sortFunction(s, numeric) {
        return function (a, b) {
            let left = getText($(a, s))
            let right = getText($(b, s))
            if (numeric) {
                left = parseFloat(left.replace(/[^0-9.-]/g, ""))
                right = parseFloat(right.replace(/[^0-9.-]/g, ""))
                if (isNaN(left) && isNaN(right)) {
                    return 0
                } else if (isNaN(left)) {
                    return 1
                } else if (isNaN(right)) {
                    return -1
                }
            }

            if (left < right) {
                return dir ? -1 : 1
            } else if (left > right) {
                return dir ? 1 : -1
            } else {
                return 0
            }
        }
    }

    function _sort() {
        const table = $("#holder table")
        const chunks = []
        $$(table, "tbody").forEach(b => {
            chunks.push(b.parentNode.removeChild(b))
        })

        let selector, num
        if ("dataset" in this) {
            selector = this.dataset.sort
            num = this.classList.contains("sort-number")
        } else {
            selector = ".name"
            num = false
        }

        if (last === selector) {
            dir = !dir
        } else {
            last = selector
            dir = true
        }

        chunks.sort(_sortFunction(selector, num))

        appendChildren(table, ...chunks)
        show(table)
    }

    return {
        sort: _sort
    }
})()

function buildMaterialsTable(types, materials) {
    const row = buildElement("tr", undefined,
        buildElement("th", undefined, "Type"),
        buildElement("th", undefined, "Quantity"),
        buildElement("th", undefined, "Avg Market Price"),
        buildElement("th", undefined, "Adj Market Price"),
        buildElement("th", undefined, "Recyled Price")
    )

    for (let i = 0; i < materials.length; i += 1) {
        const id = materials[i]
        row.appendChild(
            buildElement("th", undefined, types[id].name)
        )
    }

    return buildElement("table", undefined,
        buildElement("colgroup", undefined,
            buildElement("col", {
                span: 1,
                class: "col-name"
            }),
            buildElement("col", {
                span: 1,
                class: "col-number"
            }),
            buildElement("col", {
                span: 3,
                class: "col-currency"
            }),
            buildElement("col", {
                span: materials.length,
                class: "col-mats"
            })
        ),
        buildElement("thead", undefined, row)
    )
}

function buildMaterialsRow(materials, t) {
    const row = buildElement("tr", undefined,
        buildElement("td", undefined, t.name),
        buildElement("td", undefined, t.held),
        buildElement("td", undefined, formatPrice(t.avg)),
        buildElement("td", undefined, formatPrice(t.adj)),
        buildElement("td", undefined, formatPrice(t.matPrice))

    )

    return row
}


function showMaterials(user, assets) {
    const quants = {}
    for (let i = 0; i < assets.length; i += 1) {
        const asset = assets[i]
        const t = asset.type_id
        if (!(t in quants)) {
            quants[t] = 0
        }
        quants[t] += asset.quantity
    }

    Promise.all([
        getTypes(Object.keys(quants)),
        getPrices()
    ]).then(results => {
        const types = results[0]
        const prices = results[1]
        const materials = {}
        const result = {}
        for (let id in quants) {
            if (!(id in types)) {
                continue
            }
            const t = types[id]
            const p = prices[id]

            result[id] = {
                name: t.name,
                desc: t.desc,
                held: quants[id]
            }

            if (id in prices) {
                result[id].adj = prices[id].adj
                result[id].avg = prices[id].avg
            }

            if (t.mats !== null) {
                result[id].mats = t.mats
                result[id].matPrice = 0
                for (let i = 0; i < t.mats.length; i += 1) {
                    const mat = t.mats[i];
                    materials[mat.id] = true
                    if (mat.id in prices) {
                        result[id].matPrice += mat.v * (prices[mat.id].adj || prices[mat.id].avg || 0)
                    } else {
                        delete result[id].matPrice
                        break
                    }
                }
            }
        }
        const matList = Object.keys(materials)

        const table = buildMaterialsTable(types, matList)
        const tbody = buildElement("tbody")
        for (let id in result) {
            tbody.appendChild(buildMaterialsRow(matList, result[id]))
        }

        table.appendChild(tbody)

        $("#holder").appendChild(table)
    })
}


function showAssetsByType(user, assets) {
    const types = {}
    const toLookup = {
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
        }
    }

    const table = buildElement("table", "hidden",
        buildElement("thead", undefined, buildElement("tr", undefined,
            buildElement("th", {
                class: "sortable",
                "data-sort": ".name"
            }, "Type"),
            buildElement("th", {
                class: "sortable sort-number",
                "data-sort": ".count"
            }, "Count"),
            buildElement("th", {
                class: "sortable sort-number",
                "data-sort": ".avg"
            }, "Avg Price"),
            buildElement("th", {
                class: "sortable sort-number",
                "data-sort": ".adj"
            }, "Adj Price"),
            buildElement("th", {
                class: "sortable sort-number",
                "data-sort": ".val"
            }, "Value"),
            buildElement("th", undefined, "Location"),
            buildElement("th", undefined, "Count")
        ))
    )

    for (let id in types) {
        const tbody = buildElement("tbody")
        const locCount = Object.keys(types[id].locations).length
        const firstRow = buildElement("tr", undefined,
            buildElement("td", {
                rowspan: locCount
            }, buildElement("span", {
                class: "name",
                "data-name-id": id
            })),
            buildElement("td", {
                rowspan: locCount,
                class: "count"
            }, types[id].count),
            buildElement("td", {
                rowspan: locCount,
                class: "price"
            }, buildElement("span", {
                class: "avg",
                "data-price-id": id
            })),
            buildElement("td", {
                rowspan: locCount,
                class: "price"
            }, buildElement("span", {
                class: "adj",
                "data-price-id": id
            })),
            buildElement("td", {
                rowspan: locCount,
                class: "price"
            }, buildElement("span", {
                class: "val",
                "data-price-id": id,
                "data-count": types[id].count
            }))
        )
        let first = true
        for (let loc in types[id].locations) {
            if (first) {
                appendChildren(firstRow,
                    buildElement("td", undefined, buildElement("span", {
                        "data-name-id": loc
                    })),
                    buildElement("td", undefined, types[id].locations[loc])
                )
                tbody.appendChild(firstRow)
                first = false
            } else {
                tbody.appendChild(
                    buildElement("tr", undefined,
                        buildElement("td", undefined, buildElement("span", {
                            "data-name-id": loc
                        })),
                        buildElement("td", undefined, types[id].locations[loc])
                    )
                )
            }
        }
        table.appendChild(tbody)
    }

    Handler.on("click", ".sortable", SortAssets.sort)

    $("#holder").appendChild(table)

    Promise.all([
        getTypes(Object.keys(types)),
        getPrivateNames(user, Object.keys(toLookup.private).map(x => parseInt(x, 10))),
        getPrices()
    ]).then(() => SortAssets.sort())
}

function getAssets(user) {
    apiGet(makeApiPath("/latest/characters/" + user.ID + "/assets/")).then(json => showMaterials(user, json))
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
