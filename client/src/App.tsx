import './App.css';

import logo from './logo.svg';
import { useWebSocket } from 'react-use-websocket/dist/lib/use-websocket';

const WS_URL = 'ws://127.0.0.1:8000/ws'

function pingBackend() {
    console.log("Pinging backend")
    fetch("http://127.0.0.1:8000/ping")
}

async function signUp() {
    const username = "testusername"
    const password = "testpassword"
    console.log(`Signing up with username ${username} and password ${password}`)

    const data = await fetch("http://127.0.0.1:8000/users/signup", {
        method: "POST",
        body: JSON.stringify({
            "username": username,
            "password": password,
        }),
    })

    console.log(`Got the following from signup endpoint: ${await data.text()}`)
}

async function logIn() {
    const username = "testusername"
    const password = "testpassword"
    console.log(`Trying to log in with username ${username} and password ${password}`)

    const data = await fetch("http://127.0.0.1:8000/users/login", {
        method: "POST",
        body: JSON.stringify({
            "username": username,
            "password": password,
        }),
    })

    console.log(`Got the following from login endpoint: ${await data.text()}`)
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
    })

    return (
        <div className="App">
            <header className="App-header">
                <img src={logo} className="App-logo" alt="logo" />
                <p>
                    Edit <code>src/App.tsx</code> and save to reload.
                </p>
                <button onClick={pingBackend}>Ping the backend!</button>
                <button onClick={signUp}>Sign up</button>
                <button onClick={logIn}>Log in</button>
                <button onClick={getAllMessages}>Get all messages</button>
                <button onClick={createMessage}>Create message</button>
                <button onClick={updateMessage}>Update message</button>
                <a
                    className="App-link"
                    href="https://reactjs.org"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    Learn React
                </a>
            </header>
        </div>
    );
}

export default App;
