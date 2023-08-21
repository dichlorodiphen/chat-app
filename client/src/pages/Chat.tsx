import './Chat.css'

type ChatProps = {
    token: string
};

function Chat({ token }: ChatProps) {
    return (
        <div className="chat">
            <h1>Chat!</h1>
            <p>{token}</p>
        </div>
    );
}

export default Chat;