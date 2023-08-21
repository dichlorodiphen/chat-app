import './Message.css'

type MessageProps = {
    id: string,
    author: string,
    content: string,
    votes: string,
    created: string,
    token: string
};

function Message({ id, author, content, votes, created, token }: MessageProps) {
    async function updateMessage(vote: number) {
        const response = await fetch("http://127.0.0.1:8000/messages/" + id, {
            method: "PATCH",
            headers: {
                "Authorization": "Bearer " + token,
            },
            body: JSON.stringify({
                "vote": vote
            })
        });
    }

    async function upvote() {
        await updateMessage(1);
    }

    async function downvote() {
        await updateMessage(-1);
    }

    return (
        <div className="message">
            <p>ID: {id}, Author: {author}, Created: {created}, Votes: {votes}</p>
            <p>Content: {content}</p><br />
            <button onClick={upvote}>Upvote</button>
            <button onClick={downvote}>Downvote</button>
        </div>
    )
}

export default Message;