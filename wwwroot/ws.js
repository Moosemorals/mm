
function init() {
    var conn = new WebSocket("wss://localhost:8443/ws")

    conn.onopen = function () {
        conn.send(JSON.stringify({
            MsgType: "test",
            From: "me",
            Payload: {
                left: "1"
            }
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
