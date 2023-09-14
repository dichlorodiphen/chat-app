import './Message.css'

import { useEffect, useState } from 'react';

type MessageProps = {
    id: string,
    author: string,
    content: string,
    votes: string,
    created: string,
    token: string
};

function Message({ id, author, content, votes, created, token }: MessageProps) {
    // FIXME: Create endpoint for getting current user to see if a message has been upvoted/downvoted.
    const [upvoted, setUpvoted] = useState(false)
    const [downvoted, setDownvoted] = useState(false)

    async function updateMessage(upvoted: boolean, downvoted: boolean) {
        const response = await fetch("http://127.0.0.1:8000/messages/" + id, {
            method: "PATCH",
            headers: {
                "Authorization": "Bearer " + token,
            },
            body: JSON.stringify({
                "upvoted": upvoted,
                "downvoted": downvoted
            })
        });
    }

    async function upvote() {
        if (downvoted) {
            setDownvoted(false)
        }
        setUpvoted(!upvoted);
    }

    async function downvote() {
        if (upvoted) {
            setUpvoted(false)
        }
        setDownvoted(!downvoted)
    }

    useEffect(() => {
        console.log(`sending updateMessage: upvoted: ${upvoted}, downvoted ${downvoted}`)
        updateMessage(upvoted, downvoted);
    }, [upvoted, downvoted]);

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