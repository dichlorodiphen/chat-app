# Nimble Chat

## Disclaimer

I ran out of time at the end, so this is massively unpolished. Minimum functionality, however, is there. That being said, most of the nontrivial work, in my opinion, has been done.

## What is this?

DISCLAIMER: This section is more of a general ramble, as I don't have time to clean it up. The information here is relevant though!

Nimble Chat is a simple chatroom application. The application is structured in three tiers: a database, a backend, and a frontend.

MongoDB is used for the database. The backend is implemented in Go mostly with the standard library. The frontend is implemented in TypeScript using React.

Local Kubernetes (minikube) is used for deployment. Manifests have been included in `deployment/`.

Please ignore `build.sh`. I used it during development, but it has no purpose for users, and I'd rather not delete it right now.

OH BEFORE I FORGET! The backend uses a REST API (fully documented in the README in `server/`) for the control plane and a websocket connection `localhost:3000/ws` to stream messages.

JWT tokens are used for authentication, and bcrypt is used to hash passwords for storage in the database.

Race conditions are handled by MongoDB transactions, which you can see if you look at the code in `messages.go` (please don't, that's probably the ugliest endpoint lol).

## Known issues and areas for future work

* Websocket heartbeats set up on backend but not frontend, so the server will disconnect a client after the heartbeat timeout.
* No explicit logout functionality, but the token is stored in volatile memory, so the user will be logged out on page refresh.
* Sessions don't persist through page refreshes. This can be fixed by using some form of more persistent storage, such as session data or cookies.
* A user cannot remove a cast vote; he or she can only switch an upvote to a downvote or vice versa. This is a result of the voting logic being unnecessarily complicated and can be fixed by changing the update endpoint to toggle upvotes and downvotes for the current user.
* A good portion of the code is ugly -- sorry, again, ran out of time so no time to polish.
* Where are the tests??? -- see above.
* Definitely more to write here, but I need to record the demo now.
* OH! Also the first message in the websocket is used for authentication. Right now it just sends a predefined string, but it would take little effort to make it send a JWT token that is then verified by the server.
* Also vote counts don't refresh dynamically -- only on page refresh. This can be fixed by streaming vote updates through the websocket connection.

## Running!

Fun part! First, start up minikube with at least two nodes. I found that `minikube start --nodes=2` didn't work for me, so if that is the case for you, start minikube with `minikube start` and then add a node with `minikube node add`. Oh don't forget to start up Docker before trying to start minikube!

Next, run `minikube tunnel` so that the frontend and backend have visible IP addresses we can connect to.

Next, apply all manifests with `kubectl apply -Rf deployment/`, which will recursively apply manifests in `deployment/`. The Docker images for the deployments are stored on Docker hub, so there is no need to build anything here. I'm not going to write instructions for how to do that because I currently have 14 minutes left in the challenge and still have to record the demo lol.

At this point, you should be able to connect to the frontend at `localhost:3000`.

When you are done testing, delete the resources with `kubectl delete -Rf deployment/`.



