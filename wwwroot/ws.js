
function init() {
    var conn = new WebSocket("wss://localhost:8443/ws")

    conn.onopen = function () {
        conn.send(JSON.stringify({
            Target: "targt",
            From: "me",
            When: "now"
        }))
    }

    conn.onerror = function (err) {
        console.error("Websocket", err)
    }

    conn.onmessage = function (e) {
        console.log("Message: ", e.data)
    }
}



window.addEventListener("DOMContentLoaded", init)
