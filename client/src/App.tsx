import './fonts.css'

import Auth from './pages/Auth';
import Chat from './pages/Chat';
import { useState } from 'react';
import { useWebSocket } from 'react-use-websocket/dist/lib/use-websocket';

const WS_URL = 'ws://127.0.0.1:8000/ws'

function pingBackend() {
    console.log("Pinging backend")
    fetch("http://127.0.0.1:8000/ping")
}


async function getAllMessages() {
    console.log("trying to get all messages")
    const token = "123123123"

    const response = await fetch("http://127.0.0.1:8000/messages", {
        method: "GET",
        headers: {
            "Authorization": "Bearer " + token,
        },
    })

    console.log(`Got the following from getAllMessages endpoint: ${await response.text()}`)
}

function createMessage() {
    console.log("trying to create message NOT IMPLEMENTED")

    return
}

function updateMessage() {
    console.log("trying to update message NOT IMPLEMENTED")

    return
}

function App() {
    const { sendMessage, lastMessage, readyState } = useWebSocket(WS_URL, {
        onOpen: () => {
            console.log("Websocket connection established.")
            sendMessage("secret")
        }
    });

    const [token, setToken] = useState("");

    if (!token) {
        return <Auth setToken={setToken}></Auth>
    }

    return <Chat token={token}></Chat>

}

export default App;
