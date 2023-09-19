import "./Chat.css";

import { useEffect, useState } from "react";

import Logo from "../components/Logo";
import Message from "../components/Message";
import { useCookies } from "react-cookie";
import { useWebSocket } from "react-use-websocket/dist/lib/use-websocket";

type ChatProps = {
    token: string;
};

type Message = {
    id: string;
    author: string;
    content: string;
    votes: string;
    created: string;
};

const WS_URL = "ws://127.0.0.1:8000/ws";

function Chat({ token }: ChatProps) {
    // Current message being composed.
    const [text, setText] = useState("");

    const [history, setHistory] = useState<Message[]>([]);

    const removeCookie = useCookies(["token"])[2];

    // Set up websocket.
    const { sendMessage, lastMessage, readyState } = useWebSocket(WS_URL, {
        onOpen: () => {
            console.log("Websocket connection established.");
            sendMessage(token);
        },
        onMessage: (m) => {
            console.log(`Received data: ${m.data}`);
            const message: Message = JSON.parse(m.data);
            setHistory([...history, message]);
        },
    });

    function onTextChange(e: React.FormEvent<HTMLInputElement>) {
        setText(e.currentTarget.value);
    }

    async function onKeyUp(e: React.KeyboardEvent<HTMLInputElement>) {
        if (e.key === "Enter") {
            console.log(`Sending message ${text}`);
            setText("");
            await createMessage(text);
        }
    }

    async function createMessage(message: string) {
        const response = await fetch("http://127.0.0.1:8000/messages", {
            method: "POST",
            headers: {
                Authorization: "Bearer " + token,
            },
            body: JSON.stringify({
                content: message,
            }),
        });
        console.log(response);
    }

    async function getAllMessages(): Promise<Message[]> {
        console.log(`trying to get all messages`);

        const response = await fetch("http://127.0.0.1:8000/messages", {
            method: "GET",
            headers: {
                Authorization: "Bearer " + token,
            },
        });

        return JSON.parse(await response.text());
    }

    useEffect(() => {
        getAllMessages()
            .then((arr) => {
                setHistory(arr);
            })
            .catch((e) => {
                console.error(e);
            });
    }, []);

    return (
        <div>
            <Logo></Logo>
            <div className="chat">
                <div className="chat-history">
                    {history.map((m: Message) => {
                        return (
                            <Message
                                id={m.id}
                                author={m.author}
                                content={m.content}
                                votes={m.votes}
                                created={m.created}
                                token={token}
                            ></Message>
                        );
                    })}
                </div>
                <input
                    className="chat-input"
                    onKeyUp={onKeyUp}
                    value={text}
                    onChange={onTextChange}
                    type="text"
                ></input>
            </div>
        </div>
    );
}

export default Chat;
